package cbor

import "time"
import "math/big"
import "regexp"

// Undefined type as part of simple-type code23
type Undefined byte

// Indefinite code, 1st-byte of data item.
type Indefinite byte

// BreakStop code, last-byte of the data item.
type BreakStop byte

// Epoch tagged-type, seconds since 1970-01-01T00:00Z in UTC time.
type Epoch int64

// EpochMicro tagged-type, float64 since 1970-01-01T00:00Z in UTC time.
type EpochMicro float64

// DecimalFraction tagged-type, combine an integer mantissa with a
// base-10 scaling factor, m*(10**e). As int64{e,m}.
type DecimalFraction [2]interface{}

// BigFloat tagged-type, combine an integer mantissa with a base-2
// scaling factor, m*(10**e). As int64{e,m}.
type BigFloat [2]interface{}

// Cbor tagged-type, byte-string of cbor data-item.
type Cbor []byte

// CborPrefix tagged-type, byte-string of cbor data-item, that will be
// wrapped with a unique prefix before sending out.
type CborPrefix []byte

// MaxSmallInt is the maximum value that can be stored
// as assiative value to any major type.
const MaxSmallInt = 23

// Major types.
const (
	// Type0 is major type Unsigned integer
	Type0 byte = iota << 5
	// Type1 is major type Negative integer
	Type1
	// Type2 is major type Byte string
	Type2
	// Type3 is major type Text string
	Type3
	// Type4 is major type Array
	Type4
	// Type5 is major type Map
	Type5
	// Type6 is major type Tagging
	Type6
	// Type7 is major type floating-point, simple-types and break-stop
	Type7
)

// Associated information for Type0 and Type1.
const (
	// 0..23 actual value

	// Info24 follows 1-byte data-item
	Info24 byte = iota + 24
	// Info25 follows 2-byte data-item
	Info25
	// Info26 follows 4-byte data-item
	Info26
	// Info27 follows 8-byte data-item
	Info27

	// 28..30 reserved

	// IndefiniteLength for byte-string, string, arr, map
	IndefiniteLength = 31
)

// Simple types defined by Type7
const (
	// 0..19 unassigned

	// SimpleTypeFalse encodes nil type
	SimpleTypeFalse byte = iota + 20
	// SimpleTypeTrue encodes true type
	SimpleTypeTrue
	// SimpleTypeNil encodes nil type
	SimpleTypeNil
	// SimpleUndefined type
	SimpleUndefined
	// SimpleTypeByte says the actual type in next byte 32..255
	SimpleTypeByte
	// Float16 encodes IEEE 754 Half-Precision Float
	Float16
	// Float32 encodes IEEE 754 Single-Precision Float
	Float32
	// Float64 encodes IEEE 754 Double-Precision Float
	Float64

	// 28..30 un-assigned

	// ItemBreak encodes stop-code for indefinite-length items
	ItemBreak = 31
)

// EncodeSmallInt integers -23..+23
func EncodeSmallInt(item int8, buf []byte) int {
	if item < 0 {
		buf[0] = hdr(Type1, byte(-(item + 1))) // -23 to -1
	} else {
		buf[0] = hdr(Type0, byte(item)) // 0 to 23
	}
	return 1
}

// EncodeSimpleType that falls outside the golang native type.
// code points 0..19 and 32..255
func EncodeSimpleType(typcode byte, buf []byte) int {
	if typcode < 20 {
		buf[0] = hdr(Type7, typcode)
		return 1
	} else if typcode < 32 {
		panic("simpletype.lessthan32")
	}
	buf[0] = hdr(Type7, SimpleTypeByte)
	buf[1] = typcode
	return 2
}

// EncodeUndefined for simple type undefined.
func EncodeUndefined(buf []byte) int {
	buf[0] = hdr(Type7, SimpleUndefined)
	return 1
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
	case int64:
		n += encodeInt64(v, buf)
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
