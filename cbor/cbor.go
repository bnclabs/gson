package cbor

import "time"
import "math/big"
import "regexp"

// MaxSmallInt is the maximum value that can be stored
// as assiative value to any major type.
const MaxSmallInt = 23

// Undefined type as part of simple-type code23
type Undefined byte

// Indefinite code, 1st-byte of data item.
type Indefinite byte

// BreakStop code, last-byte of the data item.
type BreakStop byte

// EncodeSmallInt integers -23..+23
func EncodeSmallInt(item int8, buf []byte) int {
	if item < 0 {
		buf[0] = hdr(type1, byte(-(item + 1))) // -23 to -1
	} else {
		buf[0] = hdr(type0, byte(item)) // 0 to 23
	}
	return 1
}

// EncodeSimpleType that falls outside the golang native type.
// code points 0..19 and 32..255
func EncodeSimpleType(typcode byte, buf []byte) int {
	if typcode < 20 {
		buf[0] = hdr(type7, typcode)
		return 1
	} else if typcode < 32 {
		panic("simpletype.lessthan32")
	}
	buf[0] = hdr(type7, simpleTypeByte)
	buf[1] = typcode
	return 2
}

// Encode null, true, false,
// 8/16/32/64 bit ints and uints, 32/64 bit floats,
// byte string and string.
func Encode(item interface{}, buf []byte) int {
	n := 0
	switch v := item.(type) {
	case nil:
		n += encodeNull(buf)
	case bool:
		if v {
			n += encodeTrue(buf)
		} else {
			n += encodeFalse(buf)
		}
	case int8:
		n += encodeInt8(v, buf)
	case uint8:
		n += encodeUint8(v, buf)
	case int16:
		n += encodeInt16(v, buf)
	case uint16:
		n += encodeUint16(v, buf)
	case int32:
		n += encodeInt32(v, buf)
	case uint32:
		n += encodeUint32(v, buf)
	case int:
		n += encodeInt64(int64(v), buf)
	case int64:
		n += encodeInt64(v, buf)
	case uint:
		n += encodeUint64(uint64(v), buf)
	case uint64:
		n += encodeUint64(v, buf)
	case float32:
		n += encodeFloat32(v, buf)
	case float64:
		n += encodeFloat64(v, buf)
	case []byte:
		n += encodeBytes(v, buf)
	case string:
		n += encodeText(v, buf)
	case []interface{}:
		n += encodeArray(v, buf)
	case [][2]interface{}:
		n += encodeMap(v, buf)
	// simple types
	case Undefined:
		n += encodeUndefined(buf)
	// tagged encoding
	case time.Time: // tag-0
		n += encodeDateTime(v, buf)
	case Epoch: // tag-1
		n += encodeDateTime(v, buf)
	case EpochMicro: // tag-1
		n += encodeDateTime(v, buf)
	case *big.Int:
		n += encodeBigNum(v, buf)
	case DecimalFraction:
		n += encodeDecimalFraction(v, buf)
	case BigFloat:
		n += encodeBigFloat(v, buf)
	case Cbor:
		n += encodeCbor(v, buf)
	case *regexp.Regexp:
		n += encodeRegexp(v, buf)
	case CborPrefix:
		n += encodeCborPrefix(v, buf)
		// tagged encoding for custom data-type
		//default:
	}
	return n
}

func Decode(buf []byte) (interface{}, int) {
	item, n := cborDecoders[buf[0]](buf)
	return item, n
}
