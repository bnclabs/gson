package gson

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
	nk NumberKind
	ws SpaceKind
	ct CborContainerEncoding
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
	return &Config{nk: nk, ws: ws, ct: ct}
}

// ParseToValue input JSON text to a single go-native value. If text is
// invalid raises panic. Remaining unparsed text is returned,
// along with go-native value.
func (config *Config) ParseToValue(txt string) (string, interface{}) {
	return scanValue(txt, config)
}

// ParseToValues will parse input JSON text to one or more go native
// values. If text is invalid raises panic.
func (config *Config) ParseToValues(txt string) []interface{} {
	var values []interface{}
	var tok interface{}
	for len(txt) > 0 {
		txt, tok = scanValue(txt, config)
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

// Get field or nested field specified by json pointer.
func (config *Config) DocGet(ptr string, doc interface{}) (item interface{}) {
	segments := segmentsPool.Get()
	defer segmentsPool.Put(segments)
	segs := config.ParseJsonPointer(ptr, segments.([]string))
	return get(segs, doc)
}

// Set field or nested field specified by json pointer. While
// `newdoc` is gauranteed to contain the `item`, `doc` _may_ not be.
// Suggested usage,
//      doc := []interface{}{"hello"}
//      doc, _ = config.Set("/-", doc, "world")
func (config *Config) DocSet(ptr string, doc, item interface{}) (newdoc, old interface{}) {
	segments := segmentsPool.Get()
	defer segmentsPool.Put(segments)
	segs := config.ParseJsonPointer(ptr, segments.([]string))
	return set(segs, doc, item)
}

// Delete field or nested field specified by json pointer. While
// `newdoc` is gauranteed to be updated, `doc` _may_ not be.
// Suggested usage,
//      doc := []interface{}{"hello", "world"}
//      doc, _ = config.Delete("/1", doc)
func (config *Config) DocDelete(ptr string, doc interface{}) (newdoc, deleted interface{}) {
	segments := segmentsPool.Get()
	defer segmentsPool.Put(segments)
	segs := config.ParseJsonPointer(ptr, segments.([]string))
	return del(segs, doc)
}

// CborEncodeSmallInt encode tiny integers between -23..+23.
// Can be used by libraries that build on top of cbor.
func (config *Config) CborEncodeSmallInt(item int8, out []byte) int {
	if item < 0 {
		out[0] = hdr(type1, byte(-(item + 1))) // -23 to -1
	} else {
		out[0] = hdr(type0, byte(item)) // 0 to 23
	}
	return 1
}

// CborEncodeSimpleType that falls outside golang native type,
// code points 0..19 and 32..255 are un-assigned.
// Can be used by libraries that build on top of cbor.
func (config *Config) CborEncodeSimpleType(typcode byte, out []byte) int {
	return encodeSimpleType(typcode, out)
}

// IsIndefiniteBytes can be used to check the shape of cbor
// data-item, like byte-string, string, array or map, that
// is going to come afterwards.
// Can be used by libraries that build on top of cbor.
func (config *Config) IsIndefiniteBytes(b Indefinite) bool {
	return b == Indefinite(hdr(type2, indefiniteLength))
}

// IsIndefiniteText can be used to check the shape of cbor
// data-item, like byte-string, string, array or map, that
// is going to come afterwards.
// Can be used by libraries that build on top of cbor.
func (config *Config) IsIndefiniteText(b Indefinite) bool {
	return b == Indefinite(hdr(type3, indefiniteLength))
}

// IsIndefiniteArray can be used to check the shape of cbor
// data-item, like byte-string, string, array or map, that
// is going to come afterwards.
// Can be used by libraries that build on top of cbor.
func (config *Config) IsIndefiniteArray(b Indefinite) bool {
	return b == Indefinite(hdr(type4, indefiniteLength))
}

// IsIndefiniteMap can be used to check the shape of cbor
// data-item, like byte-string, string, array or map, that
// is going to come afterwards.
// Can be used by libraries that build on top of cbor.
func (config *Config) IsIndefiniteMap(b Indefinite) bool {
	return b == Indefinite(hdr(type5, indefiniteLength))
}

// IsBreakcodeBytes can be used to check whether chunks of
// cbor byte-strings are ending with the current byte.
// Can be used by libraries that build on top of cbor.
func (config *Config) IsBreakcodeBytes(b byte) bool {
	return b == hdr(type2, itemBreak)
}

// IsBreakcodeText can be used to check whether chunks of
// cbor text are ending with the current byte.
// Can be used by libraries that build on top of cbor.
func (config *Config) IsBreakcodeText(b byte) bool {
	return b == hdr(type3, itemBreak)
}

// IsBreakcodeArray can be used to check whether cbor array
// items of indefinite length are coming to an end with the
// current byte.
// Can be used by libraries that build on top of cbor.
func (config *Config) IsBreakcodeArray(b byte) bool {
	return b == hdr(type4, itemBreak)
}

// IsBreakcodeMap can be used to check whether cbor map items
// of indefinite length are coming to an end with the current
// byte.
// Can be used by libraries that build on top of cbor.
func (config *Config) IsBreakcodeMap(b byte) bool {
	return b == hdr(type5, itemBreak)
}

// CborEncode golang data into cbor binary.
func (config *Config) CborEncode(item interface{}, out []byte) int {
	return encode(item, out, config)
}

// EncodeMapItems to encode key,value pairs into cbor
func (config *Config) CborEncodeMapItems(items [][2]interface{}, out []byte) int {
	return encodeMapItems(items, out, config)
}
