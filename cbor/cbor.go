// Package cbor implements RFC-7049 to encode golang data into
// binary format and vice-versa.
//
// Following golang native types are supported,
//
//   * nil, true, false.
//   * native integer types, and its alias, of all width.
//   * float32, float64.
//   * slice of bytes.
//   * native string.
//   * slice of interface - []interface{}.
//   * map of string to interface{} - map[string]interface{}.
//
// Custom types defined by this package can also be encoded using cbor.
//
//   * Undefined - to encode a data-item as undefined.
//
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

// Undefined type as part of simple-type codepoint-23.
type Undefined byte

// Indefinite code, first-byte of stream encoded data items.
type Indefinite byte

// BreakStop code, last-byte of stream encoded the data items.
type BreakStop byte

const ( // major types.
	type0 byte = iota << 5 // unsigned integer
	type1                  // negative integer
	type2                  // byte string
	type3                  // text string
	type4                  // array
	type5                  // map
	type6                  // tagged data-item
	type7                  // floating-point, simple-types and break-stop
)

const ( // associated information for type0 and type1.
	// 0..23 actual value
	info24 byte = iota + 24 // followed by 1-byte data-item
	info25                  // followed by 2-byte data-item
	info26                  // followed by 4-byte data-item
	info27                  // followed by 8-byte data-item
	// 28..30 reserved
	indefiniteLength = 31 // for byte-string, string, arr, map
)

const ( // simple types for type7
	// 0..19 unassigned
	simpleTypeFalse byte = iota + 20 // encodes nil type
	simpleTypeTrue
	simpleTypeNil
	simpleUndefined
	simpleTypeByte // the actual type in next byte 32..255
	flt16          // IEEE 754 Half-Precision Float
	flt32          // IEEE 754 Single-Precision Float
	flt64          // IEEE 754 Double-Precision Float
	// 28..30 reserved
	itemBreak = 31 // stop-code for indefinite-length items
)

// MaxSmallInt is the maximum integer value that can be stored
// as associative value.
const MaxSmallInt = 23

func major(b byte) byte {
	return b & 0xe0
}

func info(b byte) byte {
	return b & 0x1f
}

func hdr(major, info byte) byte {
	return (major & 0xe0) | (info & 0x1f)
}

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
	tagURI          = iota + 22 // defined in rfc3986
	tagBase64URLEnc             // base64 encoded url as text strings
	tagBase64Enc                // base64 encoded byte-string as text strings
	tagRegexp                   // PCRE and ECMA262 regular expression
	tagMime                     // MIME defined by rfc2045

	// tag 37 is un-assigned as per spec and used here to encode
	// json-string, the difficulty is that JSON string are
	// not really utf8 encoded string (mostly meant for human
	// readability).
	tagJsonString

	// tag 38 is un-assigned as per spec and used here to encode
	// number as json-string, which is more optimized by avoiding
	// atoi and itoa (or similar operations for float). can be used
	// while converting json->cbor
	tagJsonNumber

	// unassigned 38..55798
	tagCborPrefix = iota + 55783
	// unassigned 55800..
)

var brkstp byte = hdr(type7, itemBreak)
