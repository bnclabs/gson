package cbor

import "math"
import "math/big"
import "regexp"
import "time"
import "encoding/binary"
import "fmt"

// Notes:
//
// 1. tagBase64URL, tagBase64, tagBase16 are used to reduce the
//   message size.
//   a. if following data-item is other than []byte then it applies
//     to all []byte contained in the data-time.
//
// 2. tagBase64URL/tagBase64 carry item in raw-byte string while
//   tagBase64URLEnc/tagBase64Enc carry item in base64 encoded
//   text-string.
//
// 3. TODO, yet to encode/decode tagBase* data-items and tagURI item.

func encode(item interface{}, out []byte, config *Config) int {
	n := 0
	switch v := item.(type) {
	case nil:
		n += encodeNull(out)
	case bool:
		if v {
			n += encodeTrue(out)
		} else {
			n += encodeFalse(out)
		}
	case int8:
		n += encodeInt8(v, out)
	case uint8:
		n += encodeUint8(v, out)
	case int16:
		n += encodeInt16(v, out)
	case uint16:
		n += encodeUint16(v, out)
	case int32:
		n += encodeInt32(v, out)
	case uint32:
		n += encodeUint32(v, out)
	case int:
		n += encodeInt64(int64(v), out)
	case int64:
		n += encodeInt64(v, out)
	case uint:
		n += encodeUint64(uint64(v), out)
	case uint64:
		n += encodeUint64(v, out)
	case float32:
		n += encodeFloat32(v, out)
	case float64:
		n += encodeFloat64(v, out)
	case []byte:
		n += encodeBytes(v, out)
	case string:
		n += encodeText(v, out)
	case []interface{}:
		n += encodeArray(v, out, config)
	case [][2]interface{}:
		n += encodeMap(v, out, config)
	// simple types
	case Undefined:
		n += encodeUndefined(out)
	// tagged encoding
	case time.Time: // tag-0
		n += encodeDateTime(v, out, config)
	case Epoch: // tag-1
		n += encodeDateTime(v, out, config)
	case EpochMicro: // tag-1
		n += encodeDateTime(v, out, config)
	case *big.Int:
		n += encodeBigNum(v, out, config)
	case DecimalFraction:
		n += encodeDecimalFraction(v, out)
	case BigFloat:
		n += encodeBigFloat(v, out)
	case Cbor:
		n += encodeCbor(v, out)
	case *regexp.Regexp:
		n += encodeRegexp(v, out)
	case CborPrefix:
		n += encodeCborPrefix(v, out)
	default:
		panic("cbor encode unknownType")
	}
	return n
}

func encodeTag(tag uint64, buf []byte) int {
	n := encodeUint64(tag, buf)
	buf[0] = (buf[0] & 0x1f) | type6 // fix the type as tag.
	return n
}

func decode(buf []byte) (interface{}, int) {
	item, n := cborDecoders[buf[0]](buf)
	if _, ok := item.(Indefinite); ok {
		switch major(buf[0]) {
		case type4:
			arr := make([]interface{}, 0, 2)
			for buf[n] != brkstp {
				item, n1 := decode(buf[n:])
				arr = append(arr, item)
				n += n1
			}
			return arr, n + 1

		case type5:
			pairs := make([][2]interface{}, 0, 2)
			for buf[n] != brkstp {
				key, n1 := decode(buf[n:])
				value, n2 := decode(buf[n+n1:])
				pairs = append(pairs, [2]interface{}{key, value})
				n = n + n1 + n2
			}
			return pairs, n + 1
		}
	}
	return item, n
}

func decodeTag(buf []byte) (interface{}, int) {
	byt := (buf[0] & 0x1f) | type0 // fix as positive num
	item, n := cborDecoders[byt](buf)
	switch item.(uint64) {
	case tagDateTime:
		item, m := decodeDateTime(buf[n:])
		return item, n + m

	case tagEpoch:
		item, m := decodeEpoch(buf[n:])
		return item, n + m

	case tagPosBignum:
		item, m := decodeBigNum(buf[n:])
		return item, n + m

	case tagNegBignum:
		item, m := decodeBigNum(buf[n:])
		return big.NewInt(0).Mul(item.(*big.Int), big.NewInt(-1)), n + m

	case tagDecimalFraction:
		item, m := decodeDecimalFraction(buf[n:])
		return item, n + m

	case tagBigFloat:
		item, m := decodeBigFloat(buf[n:])
		return item, n + m

	case tagCborEnc:
		item, m := decodeCborEnc(buf[n:])
		return item, n + m

	case tagRegexp:
		item, m := decodeRegexp(buf[n:])
		return item, n + m

	case tagJsonString:
		ln, m := decodeLength(buf[n:])
		return string(buf[n+m : n+m+ln]), n + m + ln

	case tagJsonNumber:
		ln, m := decodeLength(buf[n:])
		return string(buf[n+m : n+m+ln]), n + m + ln

	case tagCborPrefix:
		item, m := decodeCborPrefix(buf[n:])
		return item, n + m
	}
	// skip tags
	item, m := decode(buf[n:])
	return item, n + m
}

//---- encode basic data types

func encodeNull(buf []byte) int {
	buf[0] = hdr(type7, simpleTypeNil)
	return 1
}

func encodeTrue(buf []byte) int {
	buf[0] = hdr(type7, simpleTypeTrue)
	return 1
}

func encodeFalse(buf []byte) int {
	buf[0] = hdr(type7, simpleTypeFalse)
	return 1
}

func encodeUint8(item byte, buf []byte) int {
	if item <= MaxSmallInt {
		buf[0] = hdr(type0, item) // 0..23
		return 1
	}
	buf[0] = hdr(type0, info24)
	buf[1] = item // 24..255
	return 2
}

func encodeInt8(item int8, buf []byte) int {
	if item > MaxSmallInt {
		buf[0] = hdr(type0, info24)
		buf[1] = byte(item) // 24..127
		return 2
	} else if item < -MaxSmallInt {
		buf[0] = hdr(type1, info24)
		buf[1] = byte(-(item + 1)) // -128..-24
		return 2
	} else if item < 0 {
		buf[0] = hdr(type1, byte(-(item + 1))) // -23..-1
		return 1
	}
	buf[0] = hdr(type0, byte(item)) // 0..23
	return 1
}

func encodeUint16(item uint16, buf []byte) int {
	if item < 256 {
		return encodeUint8(byte(item), buf)
	}
	buf[0] = hdr(type0, info25)
	binary.BigEndian.PutUint16(buf[1:], item) // 256..65535
	return 3
}

func encodeInt16(item int16, buf []byte) int {
	if item > 127 {
		if item < 256 {
			buf[0] = hdr(type0, info24)
			buf[1] = byte(item) // 128..255
			return 2
		}
		buf[0] = hdr(type0, info25)
		binary.BigEndian.PutUint16(buf[1:], uint16(item)) // 256..32767
		return 3

	} else if item < -128 {
		if item > -256 {
			buf[0] = hdr(type1, info24)
			buf[1] = byte(-(item + 1)) // -255..-129
			return 2
		}
		buf[0] = hdr(type1, info25) // -32768..-256
		binary.BigEndian.PutUint16(buf[1:], uint16(-(item + 1)))
		return 3
	}
	return encodeInt8(int8(item), buf)
}

func encodeUint32(item uint32, buf []byte) int {
	if item < 65536 {
		return encodeUint16(uint16(item), buf) // 0..65535
	}
	buf[0] = hdr(type0, info26)
	binary.BigEndian.PutUint32(buf[1:], item) // 65536 to 4294967295
	return 5
}

func encodeInt32(item int32, buf []byte) int {
	if item > 32767 {
		if item < 65536 {
			buf[0] = hdr(type0, info25)
			binary.BigEndian.PutUint16(buf[1:], uint16(item)) // 32768..65535
			return 3
		}
		buf[0] = hdr(type0, info26) // 65536 to 2147483647
		binary.BigEndian.PutUint32(buf[1:], uint32(item))
		return 5

	} else if item < -32768 {
		if item > -65536 {
			buf[0] = hdr(type1, info25) // -65535..-32769
			binary.BigEndian.PutUint16(buf[1:], uint16(-(item + 1)))
			return 3
		}
		buf[0] = hdr(type1, info26) // -2147483648..-65536
		binary.BigEndian.PutUint32(buf[1:], uint32(-(item + 1)))
		return 5
	}
	return encodeInt16(int16(item), buf)
}

func encodeUint64(item uint64, buf []byte) int {
	if item < 4294967296 {
		return encodeUint32(uint32(item), buf) // 0..4294967295
	}
	buf[0] = hdr(type0, info27) // 4294967296 to 18446744073709551615
	binary.BigEndian.PutUint64(buf[1:], item)
	return 9
}

func encodeInt64(item int64, buf []byte) int {
	if item > 2147483647 {
		if item < 4294967296 {
			buf[0] = hdr(type0, info26) // 2147483647..4294967296
			binary.BigEndian.PutUint32(buf[1:], uint32(item))
			return 5
		}
		buf[0] = hdr(type0, info27) // 4294967296..9223372036854775807
		binary.BigEndian.PutUint64(buf[1:], uint64(item))
		return 9

	} else if item < -2147483648 {
		if item > -4294967296 {
			buf[0] = hdr(type1, info26) // -4294967295..-2147483649
			binary.BigEndian.PutUint32(buf[1:], uint32(-(item + 1)))
			return 5
		}
		buf[0] = hdr(type1, info27) // -9223372036854775808..-4294967296
		binary.BigEndian.PutUint64(buf[1:], uint64(-(item + 1)))
		return 9
	}
	return encodeInt32(int32(item), buf)
}

func encodeLength(item interface{}, buf []byte) int {
	switch v := item.(type) {
	case uint8:
		buf[0] = hdr(type0, info24)
		buf[1] = v
		return 2
	case uint16:
		buf[0] = hdr(type0, info25)
		binary.BigEndian.PutUint16(buf[1:], v)
		return 3
	case uint32:
		buf[0] = hdr(type0, info26)
		binary.BigEndian.PutUint32(buf[1:], v)
		return 5
	case uint64:
		buf[0] = hdr(type0, info27)
		binary.BigEndian.PutUint64(buf[1:], v)
		return 9
	}
	v := item.(int)
	buf[0] = hdr(type0, info27)
	binary.BigEndian.PutUint64(buf[1:], uint64(v))
	return 9
}

func encodeFloat32(item float32, buf []byte) int {
	buf[0] = hdr(type7, flt32)
	binary.BigEndian.PutUint32(buf[1:], math.Float32bits(item))
	return 5
}

func encodeFloat64(item float64, buf []byte) int {
	buf[0] = hdr(type7, flt64)
	binary.BigEndian.PutUint64(buf[1:], math.Float64bits(item))
	return 9
}

func encodeBytes(item []byte, buf []byte) int {
	n := encodeUint64(uint64(len(item)), buf)
	buf[0] = (buf[0] & 0x1f) | type2 // fix the type from type0->type2
	copy(buf[n:], item)
	return n + len(item)
}

func encodeBytesStart(buf []byte) int {
	// indefinite chunks of byte string
	buf[0] = hdr(type2, byte(indefiniteLength))
	return 1
}

func encodeText(item string, buf []byte) int {
	n := encodeBytes(str2bytes(item), buf)
	buf[0] = (buf[0] & 0x1f) | type3 // fix the type from type2->type3
	return n
}

func encodeTextStart(buf []byte) int {
	buf[0] = hdr(type3, byte(indefiniteLength)) // indefinite chunks of text
	return 1
}

func encodeArray(items []interface{}, buf []byte, config *Config) int {
	if config.Ct == LengthPrefix {
		n := encodeUint64(uint64(len(items)), buf)
		buf[0] = (buf[0] & 0x1f) | type4 // fix the type from type0->type4
		n += encodeArrayItems(items, buf[n:], config)
		return n
	}
	// Stream encoding
	n := encodeArrayStart(buf)
	n += encodeArrayItems(items, buf[n:], config)
	n += encodeBreakStop(buf[n:])
	return n
}

func encodeArrayItems(items []interface{}, buf []byte, config *Config) int {
	n := 0
	for _, item := range items {
		n += encode(item, buf[n:], config)
	}
	return n
}

func encodeArrayStart(buf []byte) int {
	buf[0] = hdr(type4, byte(indefiniteLength)) // indefinite length array
	return 1
}

func encodeMap(items [][2]interface{}, buf []byte, config *Config) int {
	if config.Ct == LengthPrefix {
		n := encodeUint64(uint64(len(items)), buf)
		buf[0] = (buf[0] & 0x1f) | type5 // fix the type from type0->type5
		n += encodeMapItems(items, buf[n:], config)
		return n
	}
	// Stream encoding
	n := encodeMapStart(buf)
	n += encodeMapItems(items, buf[n:], config)
	n += encodeBreakStop(buf[n:])
	return n
}

func encodeMapItems(items [][2]interface{}, buf []byte, config *Config) int {
	n := 0
	for _, item := range items {
		n += encode(item[0], buf[n:], config)
		n += encode(item[1], buf[n:], config)
	}
	return n
}

func encodeMapStart(buf []byte) int {
	buf[0] = hdr(type5, byte(indefiniteLength)) // indefinite length map
	return 1
}

func encodeBreakStop(buf []byte) int {
	// break stop for indefinite array or map
	buf[0] = hdr(type7, byte(itemBreak))
	return 1
}

func encodeUndefined(buf []byte) int {
	buf[0] = hdr(type7, simpleUndefined)
	return 1
}

func encodeSimpleType(typcode byte, buf []byte) int {
	if typcode < 32 {
		buf[0] = hdr(type7, typcode)
		return 1
	}
	buf[0] = hdr(type7, simpleTypeByte)
	buf[1] = typcode
	return 2
}

//---- encode tags

func encodeDateTime(dt interface{}, buf []byte, config *Config) int {
	n := 0
	switch v := dt.(type) {
	case time.Time: // rfc3339, as refined by section 3.3 rfc4287
		n += encodeTag(tagDateTime, buf)
		// TODO: make rfc3339 as config.
		n += encode(v.Format(time.RFC3339), buf[n:], config)
	case Epoch:
		n += encodeTag(tagEpoch, buf)
		n += encode(int64(v), buf[n:], config)
	case EpochMicro:
		n += encodeTag(tagEpoch, buf)
		n += encode(float64(v), buf[n:], config)
	}
	return n
}

func encodeBigNum(num *big.Int, buf []byte, config *Config) int {
	n := 0
	bytes := num.Bytes()
	if num.Sign() < 0 {
		n += encodeTag(tagNegBignum, buf)
	} else {
		n += encodeTag(tagPosBignum, buf)
	}
	n += encode(bytes, buf[n:], config)
	return n
}

func encodeDecimalFraction(item interface{}, buf []byte) int {
	n := encodeTag(tagDecimalFraction, buf)
	x := item.(DecimalFraction)
	n += encodeInt64(x[0].(int64), buf[n:])
	n += encodeInt64(x[1].(int64), buf[n:])
	return n
}

func encodeBigFloat(item interface{}, buf []byte) int {
	n := encodeTag(tagBigFloat, buf)
	x := item.(BigFloat)
	n += encodeInt64(x[0].(int64), buf[n:])
	n += encodeInt64(x[1].(int64), buf[n:])
	return n
}

func encodeCbor(item, buf []byte) int {
	n := encodeTag(tagCborEnc, buf)
	n += encodeBytes(item, buf[n:])
	return n
}

func encodeRegexp(item *regexp.Regexp, buf []byte) int {
	n := encodeTag(tagRegexp, buf)
	n += encodeText(item.String(), buf[n:])
	return n
}

func encodeCborPrefix(item, buf []byte) int {
	n := encodeTag(tagCborPrefix, buf)
	n += encodeBytes(item, buf[n:])
	return n
}

//---- decode basic data types

var cborDecoders = make(map[byte]func([]byte) (interface{}, int))

func decodeNull(buf []byte) (interface{}, int) {
	return nil, 1
}

func decodeFalse(buf []byte) (interface{}, int) {
	return false, 1
}

func decodeTrue(buf []byte) (interface{}, int) {
	return true, 1
}

func decodeSimpleTypeByte(buf []byte) (interface{}, int) {
	return buf[1], 2
}

func decodeFloat16(buf []byte) (interface{}, int) {
	panic("decodeFloat16 not supported")
}

func decodeFloat32(buf []byte) (interface{}, int) {
	item, n := binary.BigEndian.Uint32(buf[1:]), 5
	return math.Float32frombits(item), n
}

func decodeFloat64(buf []byte) (interface{}, int) {
	item, n := binary.BigEndian.Uint64(buf[1:]), 9
	return math.Float64frombits(item), n
}

func decodeType0SmallInt(buf []byte) (interface{}, int) {
	return uint64(info(buf[0])), 1
}

func decodeType1SmallInt(buf []byte) (interface{}, int) {
	return -int64(info(buf[0]) + 1), 1
}

func decodeType0Info24(buf []byte) (interface{}, int) {
	return uint64(buf[1]), 2
}

func decodeType1Info24(buf []byte) (interface{}, int) {
	return -int64(buf[1] + 1), 2
}

func decodeType0Info25(buf []byte) (interface{}, int) {
	return uint64(binary.BigEndian.Uint16(buf[1:])), 3
}

func decodeType1Info25(buf []byte) (interface{}, int) {
	return -int64(binary.BigEndian.Uint16(buf[1:]) + 1), 3
}

func decodeType0Info26(buf []byte) (interface{}, int) {
	return uint64(binary.BigEndian.Uint32(buf[1:])), 5
}

func decodeType1Info26(buf []byte) (interface{}, int) {
	return -int64(binary.BigEndian.Uint32(buf[1:]) + 1), 5
}

func decodeType0Info27(buf []byte) (interface{}, int) {
	return uint64(binary.BigEndian.Uint64(buf[1:])), 9
}

func decodeType1Info27(buf []byte) (interface{}, int) {
	x := uint64(binary.BigEndian.Uint64(buf[1:]))
	if x > 9223372036854775807 {
		panic("cbor decoding integer exceeds int64")
	}
	return int64(-x) - 1, 9
}

func decodeLength(buf []byte) (int, int) {
	if y := info(buf[0]); y < info24 {
		return int(y), 1
	} else if y == info24 {
		return int(buf[1]), 2
	} else if y == info25 {
		return int(binary.BigEndian.Uint16(buf[1:])), 3
	} else if y == info26 {
		return int(binary.BigEndian.Uint32(buf[1:])), 5
	}
	return int(binary.BigEndian.Uint64(buf[1:])), 9 // info27
}

func decodeType2(buf []byte) (interface{}, int) {
	ln, n := decodeLength(buf)
	dst := make([]byte, ln)
	copy(dst, buf[n:n+ln])
	return dst, n + ln
}

func decodeType2Indefinite(buf []byte) (interface{}, int) {
	return Indefinite(buf[0]), 1
}

func decodeType3(buf []byte) (interface{}, int) {
	ln, n := decodeLength(buf)
	dst := make([]byte, ln)
	copy(dst, buf[n:n+ln])
	return bytes2str(dst), n + ln
}

func decodeType3Indefinite(buf []byte) (interface{}, int) {
	return Indefinite(buf[0]), 1
}

func decodeType4(buf []byte) (interface{}, int) {
	ln, n := decodeLength(buf)
	arr := make([]interface{}, ln)
	for i := 0; i < ln; i++ {
		item, n1 := decode(buf[n:])
		arr[i], n = item, n+n1
	}
	return arr, n
}

func decodeType4Indefinite(buf []byte) (interface{}, int) {
	return Indefinite(buf[0]), 1
}

func decodeType5(buf []byte) (interface{}, int) {
	ln, n := decodeLength(buf)
	pairs := make([][2]interface{}, ln)
	for i := 0; i < ln; i++ {
		key, n1 := decode(buf[n:])
		value, n2 := decode(buf[n+n1:])
		pairs[i] = [2]interface{}{key, value}
		n = n + n1 + n2
	}
	return pairs, n
}

func decodeType5Indefinite(buf []byte) (interface{}, int) {
	return Indefinite(buf[0]), 1
}

func decodeBreakCode(buf []byte) (interface{}, int) {
	return BreakStop(buf[0]), 1
}

func decodeUndefined(buf []byte) (interface{}, int) {
	return Undefined(simpleUndefined), 1
}

//---- decode tags

func decodeDateTime(buf []byte) (interface{}, int) {
	item, n := decode(buf)
	item, err := time.Parse(time.RFC3339, item.(string))
	if err != nil {
		panic("decodeDateTime(): malformed time.RFC3339")
	}
	return item, n
}

func decodeEpoch(buf []byte) (interface{}, int) {
	item, n := decode(buf)
	switch v := item.(type) {
	case int64:
		return Epoch(v), n
	case uint64:
		return Epoch(v), n
	case float64:
		return EpochMicro(v), n
	}
	panic(fmt.Errorf("decodeEpoch(): neither int64 nor float64: %T", item))
}

func decodeBigNum(buf []byte) (interface{}, int) {
	item, n := decode(buf)
	num := big.NewInt(0).SetBytes(item.([]byte))
	return num, n
}

func decodeDecimalFraction(buf []byte) (interface{}, int) {
	e, x := decode(buf)
	m, y := decode(buf[x:])
	if a, ok := e.(uint64); ok {
		if b, ok := m.(uint64); ok {
			return DecimalFraction([2]interface{}{int64(a), int64(b)}), x + y
		}
		return DecimalFraction([2]interface{}{int64(a), m.(int64)}), x + y

	} else if b, ok := m.(uint64); ok {
		return DecimalFraction([2]interface{}{e.(int64), int64(b)}), x + y
	}
	return DecimalFraction([2]interface{}{e.(int64), m.(int64)}), x + y
}

func decodeBigFloat(buf []byte) (interface{}, int) {
	e, x := decode(buf)
	m, y := decode(buf[x:])
	if a, ok := e.(uint64); ok {
		if b, ok := m.(uint64); ok {
			return BigFloat([2]interface{}{int64(a), int64(b)}), x + y
		}
		return BigFloat([2]interface{}{int64(a), m.(int64)}), x + y

	} else if b, ok := m.(uint64); ok {
		return BigFloat([2]interface{}{e.(int64), int64(b)}), x + y
	}
	return BigFloat([2]interface{}{e.(int64), m.(int64)}), x + y
}

func decodeCborEnc(buf []byte) (interface{}, int) {
	item, n := decode(buf)
	return Cbor(item.([]uint8)), n
}

func decodeRegexp(buf []byte) (interface{}, int) {
	item, n := decode(buf)
	s := item.(string)
	re, err := regexp.Compile(s)
	if err != nil {
		panic(fmt.Errorf("compiling regexp %q: %v", s, err))
	}
	return re, n
}

func decodeCborPrefix(buf []byte) (interface{}, int) {
	item, n := decode(buf)
	return CborPrefix(item.([]byte)), n
}

func init() {
	makePanic := func(msg string) func([]byte) (interface{}, int) {
		return func(_ []byte) (interface{}, int) { panic(msg) }
	}
	//-- type0                  (unsigned integer)
	// 1st-byte 0..23
	for i := byte(0); i < info24; i++ {
		cborDecoders[hdr(type0, i)] = decodeType0SmallInt
	}
	// 1st-byte 24..27
	cborDecoders[hdr(type0, info24)] = decodeType0Info24
	cborDecoders[hdr(type0, info25)] = decodeType0Info25
	cborDecoders[hdr(type0, info26)] = decodeType0Info26
	cborDecoders[hdr(type0, info27)] = decodeType0Info27
	// 1st-byte 28..31
	msg := "cbor decode type0 reserved info"
	cborDecoders[hdr(type0, 28)] = makePanic(msg)
	cborDecoders[hdr(type0, 29)] = makePanic(msg)
	cborDecoders[hdr(type0, 30)] = makePanic(msg)
	msg := "cbor decode type0 indefnite"
	cborDecoders[hdr(type0, indefiniteLength)] = makePanic(msg)

	//-- type1                  (signed integer)
	// 1st-byte 0..23
	for i := byte(0); i < info24; i++ {
		cborDecoders[hdr(type1, i)] = decodeType1SmallInt
	}
	// 1st-byte 24..27
	cborDecoders[hdr(type1, info24)] = decodeType1Info24
	cborDecoders[hdr(type1, info25)] = decodeType1Info25
	cborDecoders[hdr(type1, info26)] = decodeType1Info26
	cborDecoders[hdr(type1, info27)] = decodeType1Info27
	// 1st-byte 28..31
	msg := "cbor decode type1 reserved info"
	cborDecoders[hdr(type1, 28)] = makePanic(msg)
	cborDecoders[hdr(type1, 29)] = makePanic(msg)
	cborDecoders[hdr(type1, 30)] = makePanic(msg)
	msg := "cbor decode type1 indefnite"
	cborDecoders[hdr(type1, indefiniteLength)] = makePanic(msg)

	//-- type2                  (byte string)
	// 1st-byte 0..27
	for i := 0; i < 28; i++ {
		cborDecoders[hdr(type2, byte(i))] = decodeType2
	}
	// 1st-byte 28..31
	msg := "cbor decode type2 reserved info"
	cborDecoders[hdr(type2, 28)] = makePanic(msg)
	cborDecoders[hdr(type2, 29)] = makePanic(msg)
	cborDecoders[hdr(type2, 30)] = makePanic(msg)
	msg := "cbor decode type2 indefnite"
	cborDecoders[hdr(type2, indefiniteLength)] = decodeType2Indefinite

	//-- type3                  (string)
	// 1st-byte 0..27
	for i := 0; i < 28; i++ {
		cborDecoders[hdr(type3, byte(i))] = decodeType3
	}
	// 1st-byte 28..31
	msg := "cbor decode type3 reserved info"
	cborDecoders[hdr(type3, 28)] = makePanic(msg)
	cborDecoders[hdr(type3, 29)] = makePanic(msg)
	cborDecoders[hdr(type3, 30)] = makePanic(msg)
	cborDecoders[hdr(type3, indefiniteLength)] = decodeType3Indefinite

	//-- type4                  (array)
	// 1st-byte 0..27
	for i := 0; i < 28; i++ {
		cborDecoders[hdr(type4, byte(i))] = decodeType4
	}
	// 1st-byte 28..31
	msg := "cbor decode type4 reserved info"
	cborDecoders[hdr(type4, 28)] = makePanic(msg)
	cborDecoders[hdr(type4, 29)] = makePanic(msg)
	cborDecoders[hdr(type4, 30)] = makePanic(msg)
	cborDecoders[hdr(type4, indefiniteLength)] = decodeType4Indefinite

	//-- type5                  (map)
	// 1st-byte 0..27
	for i := 0; i < 28; i++ {
		cborDecoders[hdr(type5, byte(i))] = decodeType5
	}
	// 1st-byte 28..31
	msg := "cbor decode type5 reserved info"
	cborDecoders[hdr(type5, 28)] = makePanic(msg)
	cborDecoders[hdr(type5, 29)] = makePanic(msg)
	cborDecoders[hdr(type5, 30)] = makePanic(msg)
	cborDecoders[hdr(type5, indefiniteLength)] = decodeType5Indefinite

	//-- type6
	// 1st-byte 0..23
	for i := byte(0); i < info24; i++ {
		cborDecoders[hdr(type6, i)] = decodeTag
	}
	// 1st-byte 24..27
	cborDecoders[hdr(type6, info24)] = decodeTag
	cborDecoders[hdr(type6, info25)] = decodeTag
	cborDecoders[hdr(type6, info26)] = decodeTag
	cborDecoders[hdr(type6, info27)] = decodeTag
	// 1st-byte 28..31
	msg := "cbor decode type6 reserved info"
	cborDecoders[hdr(type6, 28)] = makePanic(msg)
	cborDecoders[hdr(type6, 29)] = makePanic(msg)
	cborDecoders[hdr(type6, 30)] = makePanic(msg)
	msg := "cbor decode type6 indefnite"
	cborDecoders[hdr(type6, indefiniteLength)] = makePanic(msg)

	//-- type7                  (simple types / floats / break-stop)
	// 1st-byte 0..19
	for i := byte(0); i < 20; i++ {
		cborDecoders[hdr(type7, i)] =
			func(i byte) func([]byte) (interface{}, int) {
				return func(buf []byte) (interface{}, int) { return i, 1 }
			}(i)
	}
	// 1st-byte 20..23
	cborDecoders[hdr(type7, simpleTypeFalse)] = decodeFalse
	cborDecoders[hdr(type7, simpleTypeTrue)] = decodeTrue
	cborDecoders[hdr(type7, simpleTypeNil)] = decodeNull
	cborDecoders[hdr(type7, simpleUndefined)] = decodeUndefined

	cborDecoders[hdr(type7, simpleTypeByte)] = decodeSimpleTypeByte
	cborDecoders[hdr(type7, flt16)] = decodeFloat16
	cborDecoders[hdr(type7, flt32)] = decodeFloat32
	cborDecoders[hdr(type7, flt64)] = decodeFloat64
	// 1st-byte 28..31
	msg := "cbor decode type7 simple type"
	cborDecoders[hdr(type7, 28)] = makePanic(msg)
	cborDecoders[hdr(type7, 29)] = makePanic(msg)
	cborDecoders[hdr(type7, 30)] = makePanic(msg)
	cborDecoders[hdr(type7, itemBreak)] = decodeBreakCode
}
