//  Copyright (c) 2015 Couchbase, Inc.

package gson

const ( // major types (3 most significant bits in the first byte)
	cborType0 byte = iota << 5 // unsigned integer
	cborType1                  // negative integer
	cborType2                  // byte string
	cborType3                  // text string
	cborType4                  // array
	cborType5                  // map
	cborType6                  // tagged data-item
	cborType7                  // floating-point, simple-types and break-stop
)

// CborMaxSmallInt maximum integer value that can be stored as associative value
// for cborType0 or cborType1.
const CborMaxSmallInt = 23

const ( // for cborType0 cborType1 (5 least significant bits in the first byte)
	// 0..23 actual value
	cborInfo24 byte = iota + 24 // followed by 1-byte data-item
	cborInfo25                  // followed by 2-byte data-item
	cborInfo26                  // followed by 4-byte data-item
	cborInfo27                  // followed by 8-byte data-item
	// 28..30 reserved
	cborIndefiniteLength = 31 // for cborType2/cborType3/cborType4/cborType5
)

// CborIndefinite code, {cborType2,Type3,Type4,Type5}/cborIndefiniteLength
type CborIndefinite byte

const ( // simple types for cborType7
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

// CborUndefined simple type, cborType7/cborSimpleUndefined
type CborUndefined byte

// CborBreakStop code, cborType7/cborItemBreak
type CborBreakStop byte

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

// CborTagEpoch, codepoint-1, followed by int64 of seconds since
// 1970-01-01T00:00Z in UTC time.
type CborTagEpoch int64

// CborTagEpochMicro, codepoint-1, followed by float64 of seconds/us since
// 1970-01-01T00:00Z in UTC time.
type CborTagEpochMicro float64

// CborTagFraction, codepoint-4, followed by [2]int64{e,m} => m*(10**e).
type CborTagFraction [2]int64

// CborTagFloat codepoint-5, followed by [2]int64{e,m} => m*(2**e).
type CborTagFloat [2]int64

// CborTagBytes codepoint-24, bytes in cbor format.
type CborTagBytes []byte

// CborTagPrefix, codepoint-5579, followed by byte-string.
type CborTagPrefix []byte

var brkstp byte = cborHdr(cborType7, cborItemBreak)

var hdrIndefiniteBytes = cborHdr(cborType2, cborIndefiniteLength)
var hdrIndefiniteText = cborHdr(cborType3, cborIndefiniteLength)
var hdrIndefiniteArray = cborHdr(cborType4, cborIndefiniteLength)
var hdrIndefiniteMap = cborHdr(cborType5, cborIndefiniteLength)

// Cbor encapsulates configuration and a cbor buffer.
type Cbor struct {
	config *Config
	data   []byte
	n      int
}

// Bytes return a reference slice to encapsulated cbor-buffer.
func (cbr *Cbor) Bytes() []byte {
	return cbr.data[:cbr.n]
}

// Reset buffer to zero-length.
func (cbr *Cbor) Reset(data []byte) *Cbor {
	if data == nil {
		cbr.n = 0
		return cbr
	}
	cbr.data, cbr.n = data, len(data)
	return cbr
}

// Tovalue convert to golang native value.
func (cbr *Cbor) Tovalue() interface{} {
	value, _ /*rb*/ := cbor2value(cbr.data[:cbr.n], cbr.config)
	return value
}

// Tojson convert to json encoded value.
func (cbr *Cbor) Tojson(jsn *Json) *Json {
	in := cbr.data[:cbr.n]
	_ /*rb*/, m /*wb*/ := cbor2json(in, jsn.data[jsn.n:], cbr.config)
	jsn.n += m
	return jsn
}

// Tocollate convert to binary-collation.
func (cbr *Cbor) Tocollate(clt *Collate) *Collate {
	in := cbr.data[:cbr.n]
	_ /*rb*/, m /*wb*/ := cbor2collate(in, clt.data[clt.n:], cbr.config)
	clt.n += m
	return clt
}

// EncodeSimpletype to encode simple type into cbor buffer.
// Code points 0..19 and 32..255 are un-assigned.
func (cbr *Cbor) EncodeSimpletype(typcode byte) *Cbor {
	cbr.n += simpletypeToCbor(typcode, cbr.data[cbr.n:])
	return cbr
}

// EncodeSmallint to encode tiny integers between -23..+23 into cbor buffer.
func (cbr *Cbor) EncodeSmallint(item int8) *Cbor {
	if item < 0 {
		cbr.data[cbr.n] = cborHdr(cborType1, byte(-(item + 1))) // -23 to -1
	} else {
		cbr.data[cbr.n] = cborHdr(cborType0, byte(item)) // 0 to 23
	}
	cbr.n++
	return cbr
}

// JsonPointerToCbor converts json path in RFC-6901 into cbor format.
//func (cbor *Cbor) EncodeJsonpointer(jsonptr []byte) *Cbor {
//	if len(jsonptr) > 0 && jsonptr[0] != '/' {
//		panic("cbor expectedJsonPointer")
//	}
//	cbor.n = jptrToCbor(jsonptr, cbor.data)
//	return cbor
//}

// ToJsonpointer converts cbor encoded path into json path RFC-6901.
//func (cbr *Cbor) ToJsonpointer(out []byte) int {
//	if cbr.n > 0 {
//		if !cbr.IsIndefiniteText() {
//			panic("cbor expectedCborPointer")
//		}
//		return cborToJptr(cbr.data[:cbr.n], out)
//	}
//	return 0
//}

// Get field or nested field specified by cbor-pointer.
//func (cbr *Cbor) Get(jptr *Jsonpointer, item *Cbor) *Cbor {
//	if len(jptr.path) == 1 && jptr.path[0] == '/' {
//		item.n = copy(item.data, cbr.data[:cbr.n])
//		return item
//	}
//	item.n = cborGet(cbr.data[:cbr.n], jptr, item.data)
//	return item
//}

// Set field or nested field specified by cbor-pointer.
//func (cbr *Cbor) Set(jptr *Jsonpointer, item, newdoc, old *Cbor) *Cbor {
//	if len(jptr.path) == 1 && jptr.path[0] == '/' {
//		newdoc.n = copy(newdoc.data, item.data[:item.n])
//		old.n = copy(old.data, cbr.data[:cbr.n])
//		return cbr
//	}
//	docdata, itemdata := cbr.data[:cbr.n], item.data[:item.n]
//	newdoc.n, old.n = cborSet(docdata, jptr, itemdata, newdoc.data, old.data)
//	return cbr
//}

// Prepend item into a array or property container specified by cbor-pointer.
//func (cbr *Cbor) Prepend(jptr *Jsonpointer, item, newdoc *Cbor) *Cbor {
//	docdata, itemdata := cbr.data[:cbr.n], item.data[:item.n]
//	newdoc.n = cborPrepend(docdata, jptr, itemdata, newdoc.data, cbr.config)
//	return cbr
//}

// Delete field or nested field specified by cbor-pointer.
//func (cbr *Cbor) Delete(jptr *Jsonpointer, newdoc, deleted *Cbor) *Cbor {
//	if len(jptr.path) == 1 && jptr.path[0] == '/' {
//		panic("cbor emptyPointer")
//	}
//	docdata = cbr.data[:cbr.n]
//	newdoc.n, deleted.n = cborDel(docdata, jptr, newdoc.data, deleted.data)
//	return cbr
//}

//---- help functions.

func cborMajor(b byte) byte {
	return b & 0xe0
}

func cborInfo(b byte) byte {
	return b & 0x1f
}

func cborHdr(major, info byte) byte {
	return (major & 0xe0) | (info & 0x1f)
}
