package gson

import "encoding/binary"

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
	// TagDecimal aka decimal fraction as array of [2]num
	TagDecimal
	// TagBigFloat as array of [2]num
	TagBigFloat

	// unasigned 6..20

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
	// TagRegex for PCRE and ECMA262 (Javascript) regular expression
	TagRegex
	// TagMime as defined by rfc2045
	TagMime

	// unassigned 37..55798

	TagCborPrefix = iota + 55799

	// unassigned 55800..
)

const CborPrefix = uint16(0xd9f7)

func encodeTag(tag uint64, buf []byte) int {
	n := encodeUint64(tag, buf)
	buf[0] = (buf[0] & 0x1f) & Type6 // fix the type as tag.
	return n
}

func encodeDateTime(dt interface{}, buf []byte) int {
	n := 0
	switch dt.(type) {
	case string: // rfc3339, as refined by section 3.3 rfc4287
		n += encodeTag(TagDateTime, buf)
	default: // epoch, +/- int, +/- float
		n += encodeTag(TagEpoch, buf)
	}
	n += Encode(dt, buf[n:])
	return n
}

func encodeBigNum(num []byte, pos bool, buf []byte) int {
	n := 0
	if pos {
		n += encodeTag(TagPosBignum, buf)
	} else {
		n += encodeTag(TagNegBignum, buf)
	}
	n += Encode(num, buf[n:])
	return n
}

func encodeDecimal(m, e interface{}, buf []byte) int {
	n := encodeTag(TagDecimal, buf)
	n += Encode([]interface{}{m, e}, buf[n:])
	return n
}

func encodeBigFloat(m, e interface{}, buf []byte) int {
	n := encodeTag(TagBigFloat, buf)
	n += Encode([]interface{}{m, e}, buf[n:])
	return n
}

func encodeCbor(item, buf []byte) int {
	n := encodeTag(TagCborEnc, buf)
	n += encodeBytes(item, buf[n:])
	return n
}

func encodeRegex(item string, buf []byte) int {
	n := encodeTag(TagRegex, buf)
	n += encodeText(item, buf[n:])
	return n
}

func encodeMime(item string, buf []byte) int {
	n := encodeTag(TagMime, buf)
	n += encodeText(item, buf[n:])
	return n
}

func encodeCborPrefix(buf []byte) int {
	n := encodeTag(TagCborPrefix, buf)
	binary.BigEndian.PutUint16(buf[n:], CborPrefix)
	return n + 2
}
