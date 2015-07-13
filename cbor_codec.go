package gson

import "encoding/binary"
import "bytes"

//import "fmt"

func major(b byte) byte {
	return b & 0xe0
}

func info(b byte) byte {
	return b & 0x1f
}

func hdr(major, info byte) byte {
	return (major & 0xe0) | (info & 0x1f)
}

//---- encode functions
//
//  * all encode functions shall optionally take an input value to encode, and
//   o/p byte-slice to save the o/p.
//  * all encode functions shall return the number of bytes encoded into the
//   o/p byte-slice.

func encodeNull(buf []byte) int {
	buf[0] = hdr(Type7, SimpleTypeNil)
	return 1
}

func encodeTrue(buf []byte) int {
	buf[0] = hdr(Type7, SimpleTypeTrue)
	return 1
}

func encodeFalse(buf []byte) int {
	buf[0] = hdr(Type7, SimpleTypeFalse)
	return 1
}

func encodeUint8(item byte, buf []byte) int {
	if item <= MaxSmallInt {
		buf[0] = hdr(Type0, item) // 0..23
		return 1
	}
	buf[0] = hdr(Type0, Info24)
	buf[1] = item // 24..255
	return 2
}

func encodeInt8(item int8, buf []byte) int {
	if item > MaxSmallInt {
		buf[0] = hdr(Type0, Info24)
		buf[1] = byte(item) // 24..127
		return 2
	} else if item < -MaxSmallInt {
		buf[0] = hdr(Type1, Info24)
		buf[1] = byte(-(item + 1)) // -128..-24
		return 2
	} else if item < 0 {
		buf[0] = hdr(Type1, byte(-(item + 1))) // -23..-1
		return 1
	}
	buf[0] = hdr(Type0, byte(item)) // 0..23
	return 1
}

func encodeUint16(item uint16, buf []byte) int {
	if item < 256 {
		return encodeUint8(byte(item), buf)
	}
	buf[0] = hdr(Type0, Info25)
	binary.BigEndian.PutUint16(buf[1:], item) // 256..65535
	return 3
}

func encodeInt16(item int16, buf []byte) int {
	if item > 127 {
		if item < 256 {
			buf[0] = hdr(Type0, Info24)
			buf[1] = byte(item) // 128..255
			return 2
		}
		buf[0] = hdr(Type0, Info25)
		binary.BigEndian.PutUint16(buf[1:], uint16(item)) // 256..32767
		return 3

	} else if item < -128 {
		if item > -256 {
			buf[0] = hdr(Type1, Info24)
			buf[1] = byte(-(item + 1)) // -255..-129
			return 2
		}
		buf[0] = hdr(Type1, Info25) // -32768..-256
		binary.BigEndian.PutUint16(buf[1:], uint16(-(item + 1)))
		return 3
	}
	return encodeInt8(int8(item), buf)
}

func encodeUint32(item uint32, buf []byte) int {
	if item < 65536 {
		return encodeUint16(uint16(item), buf) // 0..65535
	}
	buf[0] = hdr(Type0, Info26)
	binary.BigEndian.PutUint32(buf[1:], item) // 65536 to 4294967295
	return 5
}

func encodeInt32(item int32, buf []byte) int {
	if item > 32767 {
		if item < 65536 {
			buf[0] = hdr(Type0, Info25)
			binary.BigEndian.PutUint16(buf[1:], uint16(item)) // 32768..65535
			return 3
		}
		buf[0] = hdr(Type0, Info26) // 65536 to 2147483647
		binary.BigEndian.PutUint32(buf[1:], uint32(item))
		return 5

	} else if item < -32768 {
		if item > -65536 {
			buf[0] = hdr(Type1, Info25) // -65535..-32769
			binary.BigEndian.PutUint16(buf[1:], uint16(-(item + 1)))
			return 3
		}
		buf[0] = hdr(Type1, Info26) // -2147483648..-65536
		binary.BigEndian.PutUint32(buf[1:], uint32(-(item + 1)))
		return 5
	}
	return encodeInt16(int16(item), buf)
}

func encodeUint64(item uint64, buf []byte) int {
	if item < 4294967296 {
		return encodeUint16(uint16(item), buf) // 0..4294967295
	}
	buf[0] = hdr(Type0, Info27) // 4294967296 to 18446744073709551615
	binary.BigEndian.PutUint64(buf[1:], item)
	return 9
}

func encodeInt64(item int64, buf []byte) int {
	if item > 2147483647 {
		if item < 4294967296 {
			buf[0] = hdr(Type0, Info26) // 2147483647..4294967296
			binary.BigEndian.PutUint32(buf[1:], uint32(item))
			return 5
		}
		buf[0] = hdr(Type0, Info27) // 4294967296..9223372036854775807
		binary.BigEndian.PutUint64(buf[1:], uint64(item))
		return 9

	} else if item < -2147483648 {
		if item > -4294967296 {
			buf[0] = hdr(Type1, Info26) // -4294967295..-2147483649
			binary.BigEndian.PutUint32(buf[1:], uint32(-(item + 1)))
			return 5
		}
		buf[0] = hdr(Type1, Info25) // -9223372036854775808..-4294967296
		binary.BigEndian.PutUint64(buf[1:], uint64(-(item + 1)))
		return 9
	}
	return encodeInt32(int32(item), buf)
}

func encodeFloat32(item float32, buf []byte) int {
	buf[0] = Type7 | Float32
	iobuf := bytes.NewBuffer(buf[1:])
	binary.Write(iobuf, binary.BigEndian, item)
	return 5
}

func encodeFloat64(item float64, buf []byte) int {
	buf[0] = hdr(Type7, Float64)
	iobuf := bytes.NewBuffer(buf[1:])
	binary.Write(iobuf, binary.BigEndian, item)
	return 9
}

func encodeBytes(item []byte, buf []byte) int {
	n := encodeUint64(uint64(len(item)), buf)
	buf[0] = (buf[0] & 0x1f) | Type2 // fix the type from Type0->Type2
	copy(buf[n:], item)
	return n + len(item)
}

func encodeBytesStart(buf []byte) int {
	// indefinite chunks of byte string
	buf[0] = hdr(Type2, byte(IndefiniteLength))
	return 1
}

func encodeText(item string, buf []byte) int {
	n := encodeBytes(str2bytes(item), buf)
	buf[0] = (buf[0] & 0x1f) | Type3 // fix the type from Type2->Type3
	return n
}

func encodeTextStart(buf []byte) int {
	buf[0] = hdr(Type3, byte(IndefiniteLength)) // indefinite chunks of text
	return 1
}

func encodeArray(items []interface{}, buf []byte) int {
	n := encodeUint64(uint64(len(items)), buf)
	buf[0] = (buf[0] & 0x1f) | Type4 // fix the type from Type0->Type4
	return encodeArrayItems(items, buf[n:])
}

func encodeArrayItems(items []interface{}, buf []byte) int {
	n := 0
	for _, item := range items {
		n += Encode(item, buf[n:])
	}
	return n
}

func encodeArrayStart(buf []byte) int {
	buf[0] = hdr(Type4, byte(IndefiniteLength)) // indefinite length array
	return 1
}

func encodeMap(items [][2]interface{}, buf []byte) int {
	n := encodeUint64(uint64(len(items)), buf)
	buf[0] = (buf[0] & 0x1f) | Type5 // fix the type from Type0->Type5
	return encodeMapItems(items, buf[n:])
}

func encodeMapItems(items [][2]interface{}, buf []byte) int {
	n := 0
	for _, item := range items {
		n += Encode(item[0], buf[n:])
		n += Encode(item[1], buf[n:])
	}
	return n
}

func encodeMapStart(buf []byte) int {
	buf[0] = hdr(Type5, byte(IndefiniteLength)) // indefinite length map
	return 1
}

func encodeBreakStop(buf []byte) int {
	// break stop for indefinite array or map
	buf[0] = hdr(Type7, byte(ItemBreak))
	return 1
}

//---- decode functions
//

func decodeFalse(buf []byte) (interface{}, int) {
	return false, 1
}

func decodeTrue(buf []byte) (interface{}, int) {
	return true, 1
}

func decodeNil(buf []byte) (interface{}, int) {
	return nil, 1
}

func decodeUndefined(buf []byte) (interface{}, int) {
	return Undefined(SimpleUndefined), 1
}

func decodeSimpleTypeByte(buf []byte) (interface{}, int) {
	if buf[1] < 32 {
		panic("simpletype.malformed")
	}
	return buf[1], 2
}

func decodeFloat16(buf []byte) (interface{}, int) {
	// TODO: how to implement this ???
	return 0, 2
}

func decodeFloat32(buf []byte) (interface{}, int) {
	var item float32
	iobuf := bytes.NewBuffer(buf[1:5])
	binary.Write(iobuf, binary.BigEndian, &item)
	return item, 5
}

func decodeFloat64(buf []byte) (interface{}, int) {
	var item float64
	iobuf := bytes.NewBuffer(buf[1:])
	binary.Write(iobuf, binary.BigEndian, &item)
	return item, 9
}

func decodeType0SmallInt(buf []byte) (interface{}, int) {
	return int64(info(buf[0])), 1
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
		panic("number exceeds the limit of int64")
	}
	return int64(-x) - 1, 9
}

func decodeLength(buf []byte) (int, int) {
	lbyte := (buf[0] & 0x1f) | Type0 // fix the type from Type*->Type0
	ln, n := cborDecoders[lbyte](buf)
	if v, ok := ln.(int64); ok {
		return int(v), n
	}
	return int(ln.(uint64)), n
}

func decodeType2(buf []byte) (interface{}, int) {
	ln, n := decodeLength(buf)
	return buf[n : n+ln], n + ln
}

func decodeType2Indefinite(buf []byte) (interface{}, int) {
	return Indefinite(buf[0]), 1
}

func decodeType3(buf []byte) (interface{}, int) {
	ln, n := decodeLength(buf)
	return bytes2str(buf[n : n+ln]), n + ln
}

func decodeType3Indefinite(buf []byte) (interface{}, int) {
	return Indefinite(buf[0]), 1
}

func decodeType4(buf []byte) (interface{}, int) {
	ln, n := decodeLength(buf)
	arr := make([]interface{}, ln)
	for i := 0; i < ln; i++ {
		item, n1 := Decode(buf[n:])
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
		key, n1 := Decode(buf[n:])
		value, n2 := Decode(buf[n+n1:])
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

var cborDecoders = make(map[byte]func([]byte) (interface{}, int))

func init() {
	makePanic := func(msg string) func([]byte) (interface{}, int) {
		return func(_ []byte) (interface{}, int) { panic(msg); return nil, 0 }
	}
	//-- Type0                  (unsigned integer)
	// 1st-byte 0..23
	for i := byte(0); i < Info24; i++ {
		cborDecoders[hdr(Type0, i)] = decodeType0SmallInt
	}
	// 1st-byte 24..27
	cborDecoders[hdr(Type0, Info24)] = decodeType0Info24
	cborDecoders[hdr(Type0, Info25)] = decodeType0Info25
	cborDecoders[hdr(Type0, Info26)] = decodeType0Info26
	cborDecoders[hdr(Type0, Info27)] = decodeType0Info27
	// 1st-byte 28..31
	cborDecoders[hdr(Type0, 28)] = makePanic("reserved")
	cborDecoders[hdr(Type0, 29)] = makePanic("reserved")
	cborDecoders[hdr(Type0, 30)] = makePanic("reserved")
	cborDecoders[hdr(Type0, IndefiniteLength)] = makePanic("indefinite -na-")

	//-- Type1                  (signed integer)
	// 1st-byte 0..23
	for i := byte(0); i < Info24; i++ {
		cborDecoders[hdr(Type1, i)] = decodeType1SmallInt
	}
	// 1st-byte 24..27
	cborDecoders[hdr(Type1, Info24)] = decodeType1Info24
	cborDecoders[hdr(Type1, Info25)] = decodeType1Info25
	cborDecoders[hdr(Type1, Info26)] = decodeType1Info26
	cborDecoders[hdr(Type1, Info27)] = decodeType1Info27
	// 1st-byte 28..31
	cborDecoders[hdr(Type1, 28)] = makePanic("reserved")
	cborDecoders[hdr(Type1, 29)] = makePanic("reserved")
	cborDecoders[hdr(Type1, 30)] = makePanic("reserved")
	cborDecoders[hdr(Type1, IndefiniteLength)] = makePanic("indefinite -na-")

	//-- Type2                  (byte string)
	// 1st-byte 0..27
	for i := 0; i < 28; i++ {
		cborDecoders[hdr(Type2, byte(i))] = decodeType2
	}
	// 1st-byte 28..31
	cborDecoders[hdr(Type2, 28)] = makePanic("reserved")
	cborDecoders[hdr(Type2, 29)] = makePanic("reserved")
	cborDecoders[hdr(Type2, 30)] = makePanic("reserved")
	cborDecoders[hdr(Type2, IndefiniteLength)] = decodeType2Indefinite

	//-- Type3                  (string)
	// 1st-byte 0..27
	for i := 0; i < 28; i++ {
		cborDecoders[hdr(Type3, byte(i))] = decodeType3
	}
	// 1st-byte 28..31
	cborDecoders[hdr(Type3, 28)] = makePanic("reserved")
	cborDecoders[hdr(Type3, 29)] = makePanic("reserved")
	cborDecoders[hdr(Type3, 30)] = makePanic("reserved")
	cborDecoders[hdr(Type3, IndefiniteLength)] = decodeType3Indefinite

	//-- Type4                  (array)
	// 1st-byte 0..27
	for i := 0; i < 28; i++ {
		cborDecoders[hdr(Type4, byte(i))] = decodeType4
	}
	// 1st-byte 28..31
	cborDecoders[hdr(Type4, 28)] = makePanic("reserved")
	cborDecoders[hdr(Type4, 29)] = makePanic("reserved")
	cborDecoders[hdr(Type4, 30)] = makePanic("reserved")
	cborDecoders[hdr(Type4, IndefiniteLength)] = decodeType4Indefinite

	//-- Type5                  (map)
	// 1st-byte 0..27
	for i := 0; i < 28; i++ {
		cborDecoders[hdr(Type5, byte(i))] = decodeType5
	}
	// 1st-byte 28..31
	cborDecoders[hdr(Type5, 28)] = makePanic("reserved")
	cborDecoders[hdr(Type5, 29)] = makePanic("reserved")
	cborDecoders[hdr(Type5, 30)] = makePanic("reserved")
	cborDecoders[hdr(Type5, IndefiniteLength)] = decodeType5Indefinite

	//-- Type7                  (simple values / floats / break-stop)
	// 1st-byte 0..19
	for i := byte(0); i < 20; i++ {
		cborDecoders[hdr(Type7, i)] =
			func(i byte) func([]byte) (interface{}, int) {
				return func(buf []byte) (interface{}, int) { return i, 1 }
			}(i)
	}
	// 1st-byte 20..23
	cborDecoders[hdr(Type7, SimpleTypeFalse)] = decodeFalse
	cborDecoders[hdr(Type7, SimpleTypeTrue)] = decodeTrue
	cborDecoders[hdr(Type7, SimpleTypeNil)] = decodeNil
	cborDecoders[hdr(Type7, SimpleUndefined)] = decodeUndefined

	cborDecoders[hdr(Type7, SimpleTypeByte)] = decodeSimpleTypeByte
	cborDecoders[hdr(Type7, Float16)] = decodeFloat16
	cborDecoders[hdr(Type7, Float32)] = decodeFloat32
	cborDecoders[hdr(Type7, Float64)] = decodeFloat64
	// 1st-byte 28..31
	cborDecoders[hdr(Type7, 28)] = makePanic("unassigned")
	cborDecoders[hdr(Type7, 29)] = makePanic("unassigned")
	cborDecoders[hdr(Type7, 30)] = makePanic("unassigned")
	cborDecoders[hdr(Type7, ItemBreak)] = decodeBreakCode
}
