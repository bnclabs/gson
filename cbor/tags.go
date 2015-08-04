// Types from golang standard library and custom defined types in
// this package that are encoded using RFC-7049 cbor-tags.
//
//   * Epoch : in seconds since epoch.
//   * EpochMicro: in micro-seconds epoch.
//   * DecimalFraction: m*(10**e)
//   * BigFloat: m*(2**e)
//   * Cbor: a cbor encoded binary data item.
//   * CborPrefix: to self indentify a binary blog as cbor.
package cbor

import "time"
import "math/big"
import "regexp"
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

// Epoch tagged-type, seconds since 1970-01-01T00:00Z in UTC time.
type Epoch int64

// EpochMicro tagged-type, float64 since 1970-01-01T00:00Z in UTC time.
type EpochMicro float64

// DecimalFraction tagged-type, combine an integer mantissa with a
// base-10 scaling factor, m*(10**e). As int64{e,m}.
type DecimalFraction [2]interface{}

// BigFloat tagged-type, combine an integer mantissa with a base-2
// scaling factor, m*(2**e). As int64{e,m}.
type BigFloat [2]interface{}

// Cbor tagged-type, byte-string of cbor data-item.
type Cbor []byte

// CborPrefix tagged-type, byte-string of cbor data-item, that will be
// wrapped with a unique prefix before sending out.
type CborPrefix []byte

const ( // pre-defined tag values
	tagDateTime        = iota // datetime as utf-8 string
	tagEpoch                  // datetime as +/- int or +/- float
	tagPosBignum              // as []bytes
	tagNegBignum              // as []bytes
	tagDecimalFraction        // decimal fraction as array of [2]num
	tagBigFloat               // as array of [2]num
	// unassigned 6..20
	// TODO: tagBase64URL, tagBase64, tagBase16
	tagBase64URL = iota + 15 // interpret []byte as base64 format
	tagBase64                // interpret []byte as base64 format
	tagBase16                // interpret []byte as base16 format
	tagCborEnc               // embedd another CBOR message
	// unassigned 25..31
	tagURI          = iota + 32 // defined in rfc3986
	tagBase64URLEnc             // base64 encoded url as text strings
	tagBase64Enc                // base64 encoded byte-string as text strings
	tagRegexp                   // PCRE and ECMA262 regular expression
	tagMime                     // MIME defined by rfc2045
	// unassigned 37..55798
	tagCborPrefix = iota + 55799
	// unassigned 55800..
)

//---- encode functions

func encodeTag(tag uint64, buf []byte) int {
	n := encodeUint64(tag, buf)
	buf[0] = (buf[0] & 0x1f) | type6 // fix the type as tag.
	return n
}

func encodeDateTime(dt interface{}, buf []byte) int {
	n := 0
	switch v := dt.(type) {
	case time.Time: // rfc3339, as refined by section 3.3 rfc4287
		n += encodeTag(tagDateTime, buf)
		n += encode(v.Format(time.RFC3339), buf[n:]) // TODO: make this config.
	case Epoch:
		n += encodeTag(tagEpoch, buf)
		n += encode(int64(v), buf[n:])
	case EpochMicro:
		n += encodeTag(tagEpoch, buf)
		n += encode(float64(v), buf[n:])
	}
	return n
}

func encodeBigNum(num *big.Int, buf []byte) int {
	n := 0
	bytes := num.Bytes()
	if num.Sign() < 0 {
		n += encodeTag(tagNegBignum, buf)
	} else {
		n += encodeTag(tagPosBignum, buf)
	}
	n += encode(bytes, buf[n:])
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

//---- decode functions

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
	}
	// tagCborPrefix:
	item, m := decodeCborPrefix(buf[n:])
	return item, n + m
}

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
