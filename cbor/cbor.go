// Package cbor implements RFC-7049 to encode golang data into
// binary format and vice-versa.
package cbor

import "time"
import "math/big"
import "regexp"

// MaxSmallInt is the maximum integer value that can be stored
// as associative value.
const MaxSmallInt = 23

// Undefined type as part of simple-type codepoint-23.
type Undefined byte

// Indefinite code, first-byte of data item.
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

// EncodeSimpleType that falls outside golang native type.
// code points 0..19 and 32..255 are un-assigned.
func EncodeSimpleType(typcode byte, buf []byte) int {
	return encodeSimpleType(typcode, buf)
}

// Encode golang data into cbor binary.
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
	default:
		panic(ErrorUnknownType)
	}
	return n
}

// Decode cbor binary into golang data.
func Decode(buf []byte) (interface{}, int) {
	item, n := cborDecoders[buf[0]](buf)
	return item, n
}

// IsIndefiniteBytes can be used to check the shape of
// data-item, like byte-string, string, array or map, that
// is going to come afterwards.
func IsIndefiniteBytes(b Indefinite) bool {
	return b == Indefinite(hdr(type2, indefiniteLength))
}

// IsIndefiniteText can be used to check the shape of
// data-item, like byte-string, string, array or map, that
// is going to come afterwards.
func IsIndefiniteText(b Indefinite) bool {
	return b == Indefinite(hdr(type3, indefiniteLength))
}

// IsIndefiniteArray can be used to check the shape of
// data-item, like byte-string, string, array or map, that
// is going to come afterwards.
func IsIndefiniteArray(b Indefinite) bool {
	return b == Indefinite(hdr(type4, indefiniteLength))
}

// IsIndefiniteMap can be used to check the shape of
// data-item, like byte-string, string, array or map, that
// is going to come afterwards.
func IsIndefiniteMap(b Indefinite) bool {
	return b == Indefinite(hdr(type5, indefiniteLength))
}
