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
//
// Package also provides encoding algorithm from json to cbor
// and vice-versa.
//
//   * number can be encoded as integer or float.
//   * string is wrapped as `tagJsonString` data-item, to avoid
//     marshalling and unmarshalling json-string to utf8.
//   * arrays and maps are encoded using indefinite encoding.
//   * byte-string encoding is not used.
package gson

// CborUndefined type as part of simple-type codepoint-23.
type CborUndefined byte

// CborIndefinite code, first-byte of stream encoded data items.
type CborIndefinite byte

// CborBreakStop code, last-byte of stream encoded the data items.
type CborBreakStop byte

// Cbor tagged-type, byte-string of cbor data-item.
type Cbor []byte

// CborPrefix tagged-type, byte-string of cbor data-item, that will be
// wrapped with a unique prefix before sending out.
type CborPrefix []byte

// CborEpoch tagged-type, seconds since 1970-01-01T00:00Z
// in UTC time.
type CborEpoch int64

// CborEpochMicro tagged-type, float64 since 1970-01-01T00:00Z
// in UTC time.
type CborEpochMicro float64

// CborDecimalFraction tagged-type, combine an integer mantissa
// with a base-10 scaling factor, m*(10**e). As int64{e,m}.
type CborDecimalFraction [2]interface{}

// CborBigFloat tagged-type, combine an integer mantissa with a base-2
// scaling factor, m*(2**e). As int64{e,m}.
type CborBigFloat [2]interface{}

const ( // major types.
	cborType0 byte = iota << 5 // unsigned integer
	cborType1                  // negative integer
	cborType2                  // byte string
	cborType3                  // text string
	cborType4                  // array
	cborType5                  // map
	cborType6                  // tagged data-item
	cborType7                  // floating-point, simple-types and break-stop
)

const ( // associated information for type0 and type1.
	// 0..23 actual value
	cborInfo24 byte = iota + 24 // followed by 1-byte data-item
	cborInfo25                  // followed by 2-byte data-item
	cborInfo26                  // followed by 4-byte data-item
	cborInfo27                  // followed by 8-byte data-item
	// 28..30 reserved
	cborIndefiniteLength = 31 // for byte-string, string, arr, map
)

const ( // simple types for type7
	// 0..19 unassigned
	cborSimpleTypeFalse byte = iota + 20 // encodes nil type
	cborSimpleTypeTrue
	cborSimpleTypeNil
	cborSimpleUndefined
	cborSimpleTypeByte // the actual type in next byte 32..255
	cborFlt16          // IEEE 754 Half-Precision Float
	cborFlt32          // IEEE 754 Single-Precision Float
	cborFlt64          // IEEE 754 Double-Precision Float
	// 28..30 reserved
	cborItemBreak = 31 // stop-code for indefinite-length items
)

// CborMaxSmallInt is the maximum integer value that can be
// stored as associative value.
const CborMaxSmallInt = 23

func cborMajor(b byte) byte {
	return b & 0xe0
}

func cborInfo(b byte) byte {
	return b & 0x1f
}

func cborHdr(major, info byte) byte {
	return (major & 0xe0) | (info & 0x1f)
}

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

var brkstp byte = cborHdr(cborType7, cborItemBreak)

var hdrIndefiniteBytes = cborHdr(cborType2, cborIndefiniteLength)
var hdrIndefiniteText = cborHdr(cborType3, cborIndefiniteLength)
var hdrIndefiniteArray = cborHdr(cborType4, cborIndefiniteLength)
var hdrIndefiniteMap = cborHdr(cborType5, cborIndefiniteLength)
