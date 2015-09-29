//  Copyright (c) 2015 Couchbase, Inc.

// Document representation, collation and transformation toolkit
//
// Package provides APIs to convert data representation from one format
// to another. Supported formats are:
//   * Json
//   * Golang value
//   * CBOR - Consice Binary Object Representation
//   * collation
//
// Package also provides a RFC-7049 (CBOR) implementation, to encode
// golang data into machine friendly binary format and vice-versa.
// Following golang native types are supported:
//   * nil, true, false.
//   * native integer types, and its alias, of all width.
//   * float32, float64.
//   * slice of bytes.
//   * native string.
//   * slice of interface - []interface{}.
//   * map of string to interface{} - map[string]interface{}.
//
// Types from golang's standard library and custom types provided
// by this package that can be encoded using CBOR:
//   * `CborUndefined` to encode a data-item as undefined.
//   * `CborIndefinite` and `CborBreakStop` to encode indefinite
//     length of bytes, string, array and map
//
//   * CborEpoch : in seconds since epoch.
//   * CborEpochMicro: in micro-seconds epoch.
//   * CborDecimalFraction: m*(10**e)
//   * CborBigFloat: m*(2**e)
//   * Cbor: a cbor encoded binary data item.
//   * CborPrefix: to self indentify a binary blog as CBOR.
//
// Package also provides an implementation for encoding json to CBOR
// and vice-versa:
//   * number can be encoded as integer or float.
//   * string is wrapped as `tagJsonString` data-item, to avoid
//     marshalling and unmarshalling json-string to utf8.
//   * arrays and maps are encoded using indefinite encoding.
//   * byte-string encoding is not used.
//
// Package also provides a RFC-6901 (JSON-pointers) implementation.
// Pointers themself can be encoded into cbor format and
// vice-versa:
//
//   cbor-path        : text-chunk-start segments break-stop
//   segments         : tagJsonString | len | cbor-text
//                    | segments segment
//   text-chunk-start : 0xdf
//   tagJsonString    : 0x25
//   break-stop       : 0xff
//   len              : <encoded as cbor integer>
//   cbor-text        : <encoded as cbor text>
package gson

import "bytes"
import "fmt"
import "encoding/json"

// NumberKind to parse JSON numbers.
type NumberKind byte

const (
	// SmartNumber32 will treat number as either integer or
	// fall back to float32.
	SmartNumber32 NumberKind = iota + 1

	// SmartNumber will treat number as either integer or
	// fall back to float64.
	SmartNumber

	// IntNumber will treat number as int64.
	IntNumber

	// FloatNumber will treat number as float32.
	FloatNumber32

	// FloatNumber will treat number as float64.
	FloatNumber

	// JsonNumber will store number in JSON encoding.
	JsonNumber

	// Decimal to collate input numbers as N, where -1 < N < 1
	Decimal
)

// SpaceKind to skip white-spaces in JSON text.
type SpaceKind byte

const (
	// AnsiSpace will skip white space characters defined by ANSI spec.
	AnsiSpace SpaceKind = iota + 1

	// UnicodeSpace will skip white space characters defined by Unicode spec.
	UnicodeSpace
)

// CborContainerEncoding, encoding method to use for arrays and maps.
type CborContainerEncoding byte

const (
	// LengthPrefix encoding for composite types. That is, for arrays and maps
	// encode the number of contained items as well.
	LengthPrefix CborContainerEncoding = iota + 1

	// Stream encoding for composite types. That is, for arrays and maps
	// use cbor's indefinite and break-stop to encode member items.
	Stream
)

// MaxKeys maximum number of keys allowed in a property object.
const MaxKeys = 1000

// Config and access gson functions. All APIs to gson is defined via
// config. To quickly get started, use NewDefaultConfig() that will
// create a configuration with default values.
type Config struct {
	nk                NumberKind
	ws                SpaceKind
	ct                CborContainerEncoding
	jsonString        bool
	arrayLenPrefix    bool // first sort arrays based on its length
	propertyLenPrefix bool // first sort properties based on length
	doMissing         bool // handle missing values (for N1QL)
	enc               *json.Encoder
	buf               *bytes.Buffer
	maxKeys           int
	// if `strict` is false then configurations with IntNumber
	// will parse floating numbers and then convert it to int64.
	// else will panic when detecting floating numbers.
	strict bool
	// memory pools
	pools mempools
	//-- unicode
	//backwards        bool
	//hiraganaQ        bool
	//caseLevel        bool
	//numeric          bool
	//nfkd              bool
	//utf8              bool
	//strength          colltab.Level
	//alternate         collate.AlternateHandling
	//language          language.Tag
}

// NewDefaultConfig returns a new configuration with default values.
// NumberKind: FloatNumber
// SpaceKind: UnicodeSpace
// CborContainerEncoding: Stream
func NewDefaultConfig() *Config {
	config := &Config{
		nk:                FloatNumber,
		ws:                UnicodeSpace,
		ct:                Stream,
		jsonString:        false,
		arrayLenPrefix:    false,
		propertyLenPrefix: true,
		doMissing:         true,
		maxKeys:           MaxKeys,
		strict:            true,
	}
	config.buf = bytes.NewBuffer(make([]byte, 0, 1024)) // TODO: no magic num.
	config.enc = json.NewEncoder(config.buf)
	strlen, numkeys, itemlen, ptrlen := 1024*1024, 1024, 1024*1024, 1024
	config.pools = newMempool(strlen, numkeys, itemlen, ptrlen)
	return config
}

// NewConfig returns a new configuration.
func NewConfig(nk NumberKind, ws SpaceKind) *Config {
	config := NewDefaultConfig()
	config.nk = nk
	config.ws = ws
	return config
}

// NumberKind representation for number types.
func (config Config) NumberKind(nk NumberKind) *Config {
	config.nk = nk
	return &config
}

// SpaceKind representation for interpreting whitespace.
func (config Config) SpaceKind(ws SpaceKind) *Config {
	config.ws = ws
	return &config
}

// ContainerEncoding for cbor.
func (config Config) ContainerEncoding(ct CborContainerEncoding) *Config {
	config.ct = ct
	return &config
}

// JsonString treat json string as it is, avoid un-quoting.
func (config Config) JsonString(what bool) *Config {
	config.jsonString = what
	return &config
}

// SortbyArrayLen sorts array by length before sorting by array elements.
func (config Config) SortbyArrayLen(what bool) *Config {
	config.arrayLenPrefix = what
	return &config
}

// SortbyPropertyLen sorts property by length before sorting by property items.
func (config Config) SortbyPropertyLen(what bool) *Config {
	config.propertyLenPrefix = what
	return &config
}

// UseMissing set or reset TypeMissing collation.
func (config Config) UseMissing(what bool) *Config {
	config.doMissing = what
	return &config
}

// SetMaxkeys will set the maximum number of keys allowed in property item.
func (config Config) SetMaxkeys(n int) *Config {
	config.maxKeys = n
	return &config
}

// Strict will set or reset the strict transforms.
func (config Config) Strict(what bool) *Config {
	config.strict = what
	return &config
}

// ResetPools will create a new set of pools with specified size.
func (config Config) ResetPools(strlen, numkeys, itemlen, ptrlen int) *Config {
	config.pools = newMempool(strlen, numkeys, itemlen, ptrlen)
	return &config
}

// JsonToValue input JSON text to a single go-native value. If text is
// invalid raises panic. Remaining unparsed text is returned,
// along with go-native value.
func (config *Config) JsonToValue(txt string) (string, interface{}) {
	return json2value(txt, config)
}

// JsonToValues will parse input JSON text to one or more go native
// values. Same as JsonToValue except that API will expect to parse
// full txt hoping to get more json values.
func (config *Config) JsonToValues(txt string) []interface{} {
	var values []interface{}
	var tok interface{}
	for len(txt) > 0 {
		txt, tok = json2value(txt, config)
		values = append(values, tok)
	}
	return values
}

// ValueToJson will convert json compatible golang value into JSON string.
// Returns the number of bytes written into `out`.
func (config *Config) ValueToJson(value interface{}, out []byte) int {
	config.buf.Reset()
	if err := config.enc.Encode(value); err != nil {
		panic(err)
	}
	s := config.buf.Bytes()
	return copy(out, s[:len(s)-1]) // -1 to strip \n
}

// ParseJsonPointer follows rfc-6901 allows ~0 and ~1 escapes, property
// lookup by specifying the key and array lookup by specifying the
// index. Also allows empty "" pointer and empty key "/".
func (config *Config) ParseJsonPointer(pointer string, sgmts []string) []string {
	return parsePointer(pointer, sgmts)
}

// ToJsonPointer reverse of ParseJsonPointer to convert parsed
// `segments` back to json-pointer. Converted pointer is available
// in the `pointer` array and returns the length of pointer-array.
func (config *Config) ToJsonPointer(segments []string, pointer []byte) int {
	return encodePointer(segments, pointer)
}

// ListPointers all possible pointers into object.
func (config *Config) ListPointers(object interface{}, ptrs []string) []string {
	prefix := config.pools.prefixPool.Get().([]byte)
	defer config.pools.prefixPool.Put(prefix[:0])
	ptrs = allpaths(object, ptrs, prefix)
	ptrs = append(ptrs, "")
	return ptrs
}

// DocGet field or nested field specified by json pointer.
func (config *Config) DocGet(ptr string, doc interface{}) (item interface{}) {
	segments := config.pools.segmentsPool.Get().([]string)
	defer config.pools.segmentsPool.Put(segments[:0])
	segs := config.ParseJsonPointer(ptr, segments)
	return docGet(segs, doc)
}

// DocSet field or nested field specified by json pointer. While
// `newdoc` is gauranteed to contain the `item`, `doc` _may_ not be.
// Suggested usage,
//      doc := []interface{}{"hello"}
//      doc, _ = config.Set("/-", doc, "world")
func (config *Config) DocSet(ptr string, doc, item interface{}) (newdoc, old interface{}) {
	segments := config.pools.segmentsPool.Get().([]string)
	defer config.pools.segmentsPool.Put(segments[:0])
	segs := config.ParseJsonPointer(ptr, segments)
	return docSet(segs, doc, item)
}

// DocDelete field or nested field specified by json pointer. While
// `newdoc` is gauranteed to be updated, `doc` _may_ not be.
// Suggested usage,
//      doc := []interface{}{"hello", "world"}
//      doc, _ = config.Delete("/1", doc)
func (config *Config) DocDelete(ptr string, doc interface{}) (newdoc, deleted interface{}) {
	segments := config.pools.segmentsPool.Get().([]string)
	defer config.pools.segmentsPool.Put(segments[:0])
	segs := config.ParseJsonPointer(ptr, segments)
	return docDel(segs, doc)
}

// SmallintToCbor encode tiny integers between -23..+23.
// Can be used by libraries that build on top of cbor.
func (config *Config) SmallintToCbor(item int8, out []byte) int {
	if item < 0 {
		out[0] = cborHdr(cborType1, byte(-(item + 1))) // -23 to -1
	} else {
		out[0] = cborHdr(cborType0, byte(item)) // 0 to 23
	}
	return 1
}

// SimpletypeToCbor that falls outside golang native type,
// code points 0..19 and 32..255 are un-assigned.
// Can be used by libraries that build on top of cbor.
func (config *Config) SimpletypeToCbor(typcode byte, out []byte) int {
	return simpletypeToCbor(typcode, out)
}

// IsIndefiniteBytes can be used to check the shape of cbor
// data-item, like byte-string, string, array or map, that
// is going to come afterwards.
// Can be used by libraries that build on top of cbor.
func (config *Config) IsIndefiniteBytes(b CborIndefinite) bool {
	return b == CborIndefinite(hdrIndefiniteBytes)
}

// IsIndefiniteText can be used to check the shape of cbor
// data-item, like byte-string, string, array or map, that
// is going to come afterwards.
// Can be used by libraries that build on top of cbor.
func (config *Config) IsIndefiniteText(b CborIndefinite) bool {
	return b == CborIndefinite(hdrIndefiniteText)
}

// IsIndefiniteArray can be used to check the shape of cbor
// data-item, like byte-string, string, array or map, that
// is going to come afterwards.
// Can be used by libraries that build on top of cbor.
func (config *Config) IsIndefiniteArray(b CborIndefinite) bool {
	return b == CborIndefinite(hdrIndefiniteArray)
}

// IsIndefiniteMap can be used to check the shape of cbor
// data-item, like byte-string, string, array or map, that
// is going to come afterwards.
// Can be used by libraries that build on top of cbor.
func (config *Config) IsIndefiniteMap(b CborIndefinite) bool {
	return b == CborIndefinite(hdrIndefiniteMap)
}

// IsBreakstop can be used to check whether chunks of
// cbor bytes, or texts, or array items or map items
// ending with the current byte. Can be used by libraries
// that build on top of cbor.
func (config *Config) IsBreakstop(b byte) bool {
	return b == brkstp
}

// ValueToCbor golang data into cbor binary.
func (config *Config) ValueToCbor(item interface{}, out []byte) int {
	return value2cbor(item, out, config)
}

// MapsliceToCbor to encode key,value pairs into cbor
func (config *Config) MapsliceToCbor(items [][2]interface{}, out []byte) int {
	return mapl2cbor(items, out, config)
}

// CborToValue cbor binary into golang data.
func (config *Config) CborToValue(buf []byte) (interface{}, int) {
	return cbor2value(buf, config)
}

// JsonToCbor input JSON text to cbor binary. Returns length of
// `out`.
func (config *Config) JsonToCbor(txt string, out []byte) (string, int) {
	return json2cbor(txt, out, config)
}

// CborToJson converts CBOR binary data-item into JSON. Returns
// length of `out`.
func (config *Config) CborToJson(in, out []byte) (int, int) {
	return cbor2json(in, out, config)
}

// JsonPointerToCbor converts json path in RFC-6901 into cbor format.
// Returns length of `out`.
func (config *Config) JsonPointerToCbor(jsonptr, out []byte) int {
	if len(jsonptr) > 0 && jsonptr[0] != '/' {
		panic("cbor expectedJsonPointer")
	}
	return jptrToCbor(jsonptr, out)
}

// CborToJsonPointer coverts cbor encoded path into json path RFC-6901.
// Returns length of `out`.
func (config *Config) CborToJsonPointer(cborptr, out []byte) int {
	if !config.IsIndefiniteText(CborIndefinite(cborptr[0])) {
		panic("cbor expectedCborPointer")
	}
	return cborToJptr(cborptr, out)
}

// CborGet field or nested field specified by cbor-pointer. Returns
// length of `item`.
func (config *Config) CborGet(doc, cborptr, item []byte) int {
	if !config.IsIndefiniteText(CborIndefinite(cborptr[0])) {
		panic("cbor expectedCborPointer")
	} else if cborptr[1] == brkstp {
		copy(item, doc)
		return len(doc)
	}
	return cborGet(doc, cborptr, item)
}

// CborSet field or nested field specified by cbor-pointer. Returns
// length of `newdoc` and `old` item.
func (config *Config) CborSet(doc, cborptr, item, newdoc, old []byte) (int, int) {
	if !config.IsIndefiniteText(CborIndefinite(cborptr[0])) {
		panic("cbor expectedCborPointer")
	} else if cborptr[1] == brkstp { // json-pointer is ""
		copy(newdoc, item)
		copy(old, doc)
		return len(item), len(doc)
	}
	return cborSet(doc, cborptr, item, newdoc, old)
}

// CborPrepend item into a array or property container specified by cbor-pointer.
// Returns length of `newdoc`.
func (config *Config) CborPrepend(doc, cborptr, item, newdoc []byte) int {
	if !config.IsIndefiniteText(CborIndefinite(cborptr[0])) {
		panic("cbor expectedCborPointer")
	}
	return cborPrepend(doc, cborptr, item, newdoc, config)
}

// CborDelete field or nested field specified by json pointer. Returns
// length of `newdoc` and `deleted` item.
func (config *Config) CborDelete(doc, cborptr, newdoc, deleted []byte) (int, int) {
	if !config.IsIndefiniteText(CborIndefinite(cborptr[0])) {
		panic("cbor expectedCborPointer")
	} else if cborptr[1] == brkstp { // json-pointer is ""
		panic("cbor emptyPointer")
	}
	return cborDel(doc, cborptr, newdoc, deleted)
}

// ValueToCollate encode input golang object to order preserving
// binary representation.
func (config *Config) ValueToCollate(obj interface{}, code []byte) int {
	return gson2collate(obj, code, config)
}

// CollateToValue will decode collated object back to golang object.
func (config *Config) CollateToValue(code []byte) (interface{}, int) {
	if len(code) == 0 {
		return nil, 0
	}
	return collate2gson(code, config)
}

// JsonToCollate encode input json text into order preserving
// binary representation.
func (config *Config) JsonToCollate(text string, code []byte) int {
	_, n := json2collate(text, code, config)
	return n
}

// CollateToJson will decode collated text back to JSON.
func (config *Config) CollateToJson(code, text []byte) (int, int) {
	if len(code) == 0 {
		return 0, 0
	}
	return collate2json(code, text, config)
}

// CborToCollate encode input cbor encoded item into order preserving
// binary representation.
func (config *Config) CborToCollate(cborin, code []byte) (int, int) {
	return cbor2collate(cborin, code, config)
}

// CollateToCbor will decode collated item back to Cbor format.
func (config *Config) CollateToCbor(code, cborout []byte) (int, int) {
	if len(code) == 0 {
		return 0, 0
	}
	return collate2cbor(code, cborout, config)
}

func (config *Config) ConfigString() string {
	return fmt.Sprintf(
		"nk:%v, ws:%v, ct:%v, jsonString:%v, arrayLenPrefix:%v, "+
			"propertyLenPrefix:%v, doMissing:%v, maxKeys:%v",
		config.nk, config.ws, config.ct, config.jsonString,
		config.arrayLenPrefix, config.propertyLenPrefix,
		config.doMissing, config.maxKeys)
}

func (nk NumberKind) String() string {
	switch nk {
	case SmartNumber32:
		return "SmartNumber32"
	case SmartNumber:
		return "SmartNumber"
	case IntNumber:
		return "IntNumber"
	case FloatNumber32:
		return "FloatNumber32"
	case FloatNumber:
		return "FloatNumber"
	case JsonNumber:
		return "JsonNumber"
	case Decimal:
		return "Decimal"
	default:
		panic("new number-kind")
	}
}

func (ws SpaceKind) String() string {
	switch ws {
	case AnsiSpace:
		return "AnsiSpace"
	case UnicodeSpace:
		return "UnicodeSpace"
	default:
		panic("new space-kind")
	}
}

func (ct CborContainerEncoding) String() string {
	switch ct {
	case LengthPrefix:
		return "LengthPrefix"
	case Stream:
		return "Stream"
	default:
		panic("new space-kind")
	}
}
