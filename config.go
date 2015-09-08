package gson

import "bytes"
import "encoding/json"

// NumberKind to parse JSON numbers.
type NumberKind byte

const (
	// SmartNumber will either use str.Atoi to parse JSON numbers
	// or fall back to float32.
	SmartNumber32 NumberKind = iota + 1
	// SmartNumber will either use str.Atoi to parse JSON numbers
	// or fall back to float64.
	SmartNumber
	// IntNumber will use str.Atoi to parse JSON numbers.
	IntNumber
	// FloatNumber will use 32 bit strconv.ParseFloat to parse JSON numbers.
	FloatNumber32
	// FloatNumber will use 64 bit strconv.ParseFloat to parse JSON numbers.
	FloatNumber
	// JsonNumber will store number in JSON encoding.
	JsonNumber
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

// Config and access gson functions. All APIs to gson is defined via
// config. To quickly get started, use NewDefaultConfig() that will
// create a configuration with default values.
type Config struct {
	nk  NumberKind
	ws  SpaceKind
	ct  CborContainerEncoding
	enc *json.Encoder
	buf *bytes.Buffer
}

// NewDefaultConfig returns a new configuration with default values.
// NumberKind: FloatNumber
// SpaceKind: UnicodeSpace
// CborContainerEncoding: Stream
func NewDefaultConfig() *Config {
	return NewConfig(FloatNumber, UnicodeSpace, Stream)
}

// NewConfig returns a new configuration.
func NewConfig(nk NumberKind, ws SpaceKind, ct CborContainerEncoding) *Config {
	config := &Config{nk: nk, ws: ws, ct: ct}
	config.buf = bytes.NewBuffer(make([]byte, 0, 1024)) // TODO: no magic num.
	config.enc = json.NewEncoder(config.buf)
	return config
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
	prefix := prefixPool.Get()
	defer prefixPool.Put(prefix)
	ptrs = allpaths(object, ptrs, prefix.([]byte))
	ptrs = append(ptrs, "")
	return ptrs
}

// DocGet field or nested field specified by json pointer.
func (config *Config) DocGet(ptr string, doc interface{}) (item interface{}) {
	segments := segmentsPool.Get()
	defer segmentsPool.Put(segments)
	segs := config.ParseJsonPointer(ptr, segments.([]string))
	return docGet(segs, doc)
}

// DocSet field or nested field specified by json pointer. While
// `newdoc` is gauranteed to contain the `item`, `doc` _may_ not be.
// Suggested usage,
//      doc := []interface{}{"hello"}
//      doc, _ = config.Set("/-", doc, "world")
func (config *Config) DocSet(ptr string, doc, item interface{}) (newdoc, old interface{}) {
	segments := segmentsPool.Get()
	defer segmentsPool.Put(segments)
	segs := config.ParseJsonPointer(ptr, segments.([]string))
	return docSet(segs, doc, item)
}

// DocDelete field or nested field specified by json pointer. While
// `newdoc` is gauranteed to be updated, `doc` _may_ not be.
// Suggested usage,
//      doc := []interface{}{"hello", "world"}
//      doc, _ = config.Delete("/1", doc)
func (config *Config) DocDelete(ptr string, doc interface{}) (newdoc, deleted interface{}) {
	segments := segmentsPool.Get()
	defer segmentsPool.Put(segments)
	segs := config.ParseJsonPointer(ptr, segments.([]string))
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

// IsBreakcodeBytes can be used to check whether chunks of
// cbor byte-strings are ending with the current byte.
// Can be used by libraries that build on top of cbor.
func (config *Config) IsBreakcodeBytes(b byte) bool {
	return b == hdrBreakcodeBytes
}

// IsBreakcodeText can be used to check whether chunks of
// cbor text are ending with the current byte.
// Can be used by libraries that build on top of cbor.
func (config *Config) IsBreakcodeText(b byte) bool {
	return b == hdrBreakcodeText
}

// IsBreakcodeArray can be used to check whether cbor array
// items of indefinite length are coming to an end with the
// current byte.
// Can be used by libraries that build on top of cbor.
func (config *Config) IsBreakcodeArray(b byte) bool {
	return b == hdrBreakcodeArray
}

// IsBreakcodeMap can be used to check whether cbor map items
// of indefinite length are coming to an end with the current
// byte.
// Can be used by libraries that build on top of cbor.
func (config *Config) IsBreakcodeMap(b byte) bool {
	return b == hdrBreakcodeMap
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
	return cbor2value(buf)
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
