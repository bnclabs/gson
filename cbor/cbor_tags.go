package cbor

import "time"
import "math/big"
import "regexp"
import "fmt"

// Notes:
//
// 1. TagBase64URL, TagBase64, TagBase16 are used to reduce the message size.
//   a. if following data-item is other than []byte then it applies
//     to all []byte contained in the data-time.
//
// 2. TagBase64URL/TagBase64 carry item in raw-byte string while
//   TagBase64URLEnc/TagBase64Enc carry item in base64 encoded text-string.
//
// 3. TODO, yet to encode/decode TagBase* data-items and TagURI item.

const (
	// TagDateTime as utf-8 string
	TagDateTime = iota
	// TagEpoch as +/- int or +/- float
	TagEpoch
	// TagPosBignum as []bytes
	TagPosBignum
	// TagNegBignum as []bytes
	TagNegBignum
	// TagDecimalFraction aka decimal fraction as array of [2]num
	TagDecimalFraction
	// TagBigFloat as array of [2]num
	TagBigFloat

	// unasigned 6..20

	// TODO: TagBase64URL, TagBase64, TagBase16

	// TagBase64URL tells decoder that []byte to surface up in base64 format
	TagBase64URL = iota + 21
	// TagBase64 tells decoder that []byte to surface up in base64 format
	TagBase64
	// TagBase64 tells decoder that []byte to surface up in base16 format
	TagBase16

	// TagCborEnc embedds another CBOR message
	TagCborEnc

	// unassigned 25..31

	// TagURI as defined in rfc3986
	TagURI = iota + 32
	// TagBase64URLEnc base64 encoded url as text strings
	TagBase64URLEnc
	// TagBase64Enc base64 encoded byte-string as text strings
	TagBase64Enc
	// TagRegexp for PCRE and ECMA262 (Javascript) regular expression
	TagRegexp
	// TagMime as defined by rfc2045
	TagMime

	// unassigned 37..55798

	TagCborPrefix = iota + 55799

	// unassigned 55800..
)

//---- encode functions

func encodeTag(tag uint64, buf []byte) int {
	n := encodeUint64(tag, buf)
	buf[0] = (buf[0] & 0x1f) & Type6 // fix the type as tag.
	return n
}

func encodeDateTime(dt interface{}, buf []byte) int {
	n := 0
	switch v := dt.(type) {
	case time.Time: // rfc3339, as refined by section 3.3 rfc4287
		n += encodeTag(TagDateTime, buf)
		n += Encode(v.Format(time.RFC3339), buf[n:]) // TODO: make this config.
	case Epoch:
		n += encodeTag(TagEpoch, buf)
		n += Encode(int64(v), buf[n:])
	case EpochMicro:
		n += encodeTag(TagEpoch, buf)
		n += Encode(float64(v), buf[n:])
	}
	return n
}

func encodeBigNum(num *big.Int, buf []byte) int {
	n := 0
	bytes := num.Bytes()
	if num.Sign() < 0 {
		n += encodeTag(TagPosBignum, buf)
	} else {
		n += encodeTag(TagNegBignum, buf)
	}
	n += Encode(bytes, buf[n:])
	return n
}

func encodeDecimalFraction(item [2]interface{}, buf []byte) int {
	n := encodeTag(TagDecimalFraction, buf)
	n += Encode(item, buf[n:])
	return n
}

func encodeBigFloat(item [2]interface{}, buf []byte) int {
	n := encodeTag(TagBigFloat, buf)
	n += Encode(item, buf[n:])
	return n
}

func encodeCbor(item, buf []byte) int {
	n := encodeTag(TagCborEnc, buf)
	n += encodeBytes(item, buf[n:])
	return n
}

func encodeRegexp(item *regexp.Regexp, buf []byte) int {
	n := encodeTag(TagRegexp, buf)
	n += encodeText(item.String(), buf[n:])
	return n
}

func encodeCborPrefix(item, buf []byte) int {
	n := encodeTag(TagCborPrefix, buf)
	n += encodeBytes(item, buf[n:])
	return n
}

//---- decode functions

func decodeTag(buf []byte) (interface{}, int) {
	byt := (buf[0] & 0x1f) & Type0 // fix as positive num
	item, n := cborDecoders[byt](buf)
	switch item.(uint64) {
	case TagDateTime:
		item, m := decodeDateTime(buf[n:])
		return item, n + m

	case TagEpoch:
		item, m := decodeEpoch(buf[n:])
		return item, n + m

	case TagPosBignum:
		item, m := decodeBigNum(buf[n:])
		return item, n + m

	case TagNegBignum:
		item, m := decodeBigNum(buf[n:])
		return item, n + m

	case TagDecimalFraction:
		item, m := decodeDecimalFraction(buf[n:])
		return item, m + n

	case TagBigFloat:
		item, m := decodeBigFloat(buf[n:])
		return item, m + n

	case TagCborEnc:
		item, m := decodeCborEnc(buf[n:])
		return item, m + n

	case TagRegexp:
		item, m := decodeRegexp(buf[n:])
		return item, m + n

	case TagCborPrefix:
		item, m := decodeCborPrefix(buf[n:])
		return item, m + n
	}
	return nil, 0
}

func decodeDateTime(buf []byte) (interface{}, int) {
	item, n := Decode(buf)
	item, err := time.Parse(time.RFC3339, item.(string))
	if err != nil {
		panic("decodeDateTime(): malformed time.RFC3339")
	}
	return item, n
}

func decodeEpoch(buf []byte) (interface{}, int) {
	item, n := Decode(buf)
	switch v := item.(type) {
	case int64:
		return Epoch(v), n
	case float64:
		return EpochMicro(v), n
	default:
		panic("decodeEpoch(): neither int64 nor float64")
	}
	return nil, 0
}

func decodeBigNum(buf []byte) (interface{}, int) {
	item, n := Decode(buf)
	num := big.NewInt(0).SetBytes(item.([]byte))
	return num, n
}

func decodeDecimalFraction(buf []byte) (interface{}, int) {
	item, n := Decode(buf)
	x := item.([]interface{})
	return DecimalFraction([2]interface{}{x[0], x[1]}), n
}

func decodeBigFloat(buf []byte) (interface{}, int) {
	item, n := Decode(buf)
	x := item.([]interface{})
	return BigFloat([2]interface{}{x[0], x[1]}), n
}

func decodeCborEnc(buf []byte) (interface{}, int) {
	item, n := Decode(buf)
	return item, n
}

func decodeRegexp(buf []byte) (interface{}, int) {
	item, n := Decode(buf)
	s := bytes2str(item.([]byte))
	re, err := regexp.Compile(s)
	if err != nil {
		panic(fmt.Errorf("compiling regexp %q: %v", s, err))
	}
	return re, n
}

func decodeCborPrefix(buf []byte) (interface{}, int) {
	item, n := Decode(buf)
	return item, n
}
