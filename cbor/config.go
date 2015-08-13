package cbor

// MaxSmallInt is the maximum integer value that can be stored
// as associative value.
const MaxSmallInt = 23

// Undefined type as part of simple-type codepoint-23.
type Undefined byte

// Indefinite code, first-byte of data item.
type Indefinite byte

// BreakStop code, last-byte of the data item.
type BreakStop byte

// NumberKind to parse JSON numbers.
type NumberKind byte

var brkstp byte = hdr(type7, itemBreak)

const (
	// SmartNumber will either use str.Atoi to parse JSON numbers
	// or fall back to float32. Default.
	SmartNumber32 NumberKind = iota + 1
	// SmartNumber will either use str.Atoi to parse JSON numbers
	// or fall back to float64. Default.
	SmartNumber
	// IntNumber will use str.Atoi to parse JSON numbers.
	IntNumber
	// FloatNumber will use 32 bit strconv.ParseFloat to parse JSON numbers.
	FloatNumber32
	// FloatNumber will use 64 bit strconv.ParseFloat to parse JSON numbers.
	FloatNumber
)

// SpaceKind to skip white-spaces in JSON text.
type SpaceKind byte

const (
	// AnsiSpace will skip white space characters defined by ANSI spec.
	AnsiSpace SpaceKind = iota + 1
	// UnicodeSpace will skip white space characters defined by Unicode spec.
	// Default.
	UnicodeSpace
)

// Config and access cbor functions. All APIs to Cbor is
// defined via config. To quickly get started, use NewDefaultConfig()
// that will create a configuration with default values.
//
// Conventions in APIs.
//
//   * out []byte, if present, saves o/p. must be sufficiently large.
//   * buf []byte, if present, provides i/p.
type Config struct {
	// Nk number kind
	Nk NumberKind
	// Ws whitespace type
	Ws SpaceKind
	// Stream to use indefinite encoding for arrays and maps
	Stream bool
}

// NewDefaultConfig returns a new configuration factory, with default
// values,
//      Nk: FloatNumber
//      Ws: UnicodeSpace
//      Stream: true
func NewDefaultConfig() *Config {
	return NewConfig(FloatNumber, UnicodeSpace, true)
}

// NewConfig returns a new configuration factory
func NewConfig(nk NumberKind, ws SpaceKind, s bool) *Config {
	return &Config{Nk: nk, Ws: ws, Stream: s}
}

// EncodeSmallInt encode tiny integers between -23..+23.
// Can be used by libraries that build on top of cbor.
func (config *Config) EncodeSmallInt(item int8, out []byte) int {
	if item < 0 {
		out[0] = hdr(type1, byte(-(item + 1))) // -23 to -1
	} else {
		out[0] = hdr(type0, byte(item)) // 0 to 23
	}
	return 1
}

// EncodeSimpleType that falls outside golang native type,
// code points 0..19 and 32..255 are un-assigned.
// Can be used by libraries that build on top of cbor.
func (config *Config) EncodeSimpleType(typcode byte, out []byte) int {
	return encodeSimpleType(typcode, out)
}

// IsIndefiniteBytes can be used to check the shape of
// data-item, like byte-string, string, array or map, that
// is going to come afterwards.
// Can be used by libraries that build on top of cbor.
func (config *Config) IsIndefiniteBytes(b Indefinite) bool {
	return b == Indefinite(hdr(type2, indefiniteLength))
}

// IsIndefiniteText can be used to check the shape of
// data-item, like byte-string, string, array or map, that
// is going to come afterwards.
// Can be used by libraries that build on top of cbor.
func (config *Config) IsIndefiniteText(b Indefinite) bool {
	return b == Indefinite(hdr(type3, indefiniteLength))
}

// IsIndefiniteArray can be used to check the shape of
// data-item, like byte-string, string, array or map, that
// is going to come afterwards.
// Can be used by libraries that build on top of cbor.
func (config *Config) IsIndefiniteArray(b Indefinite) bool {
	return b == Indefinite(hdr(type4, indefiniteLength))
}

// IsIndefiniteMap can be used to check the shape of
// data-item, like byte-string, string, array or map, that
// is going to come afterwards.
// Can be used by libraries that build on top of cbor.
func (config *Config) IsIndefiniteMap(b Indefinite) bool {
	return b == Indefinite(hdr(type5, indefiniteLength))
}

// Encode golang data into cbor binary.
func (config *Config) Encode(item interface{}, out []byte) int {
	return encode(item, out)
}

// EncodeMapItems to encode key,value pairs into cbor
func (config *Config) EncodeMapItems(items [][2]interface{}, out []byte) int {
	return encodeMapItems(items, out)
}

// Decode cbor binary into golang data.
func (config *Config) Decode(buf []byte) (interface{}, int) {
	return decode(buf)
}

// Parse input JSON text to cbor binary. Returns length of
// `out`.
func (config *Config) ParseJson(txt string, out []byte) (string, int) {
	return scanToken(txt, out, config)
}

// ToJson converts CBOR binary data-item into JSON. Returns
// length of `out`.
func (config *Config) ToJson(in, out []byte) (int, int) {
	n, m := cborTojson[in[0]](in, out)
	return n, m
}

// FromJsonPointer converts json path in RFC-6901 into cbor format.
// Returns length of `out`.
func (config *Config) FromJsonPointer(jsonptr, out []byte) int {
	if len(jsonptr) > 0 && jsonptr[0] != '/' {
		panic(ErrorExpectedJsonPointer)
	}
	return fromJsonPointer(jsonptr, out)
}

// ToJsonPointer coverts cbor encoded path into json path RFC-6901.
// Returns length of `out`.
func (config *Config) ToJsonPointer(cborptr, out []byte) int {
	if !config.IsIndefiniteText(Indefinite(cborptr[0])) {
		panic(ErrorExpectedCborPointer)
	}
	return toJsonPointer(cborptr, out)
}

// Get field or nested field specified by cbor-pointer. Returns
// length of `item`.
func (config *Config) Get(doc, cborptr, item []byte) int {
	if !config.IsIndefiniteText(Indefinite(cborptr[0])) {
		panic(ErrorExpectedCborPointer)
	} else if cborptr[1] == brkstp {
		copy(item, doc)
		return len(doc)
	}
	return get(doc, cborptr, item)
}

// Set field or nested field specified by cbor-pointer. Returns
// length of `newdoc` and `old` item.
func (config *Config) Set(doc, cborptr, item, newdoc, old []byte) (int, int) {
	if !config.IsIndefiniteText(Indefinite(cborptr[0])) {
		panic(ErrorExpectedCborPointer)
	} else if cborptr[1] == brkstp { // json-pointer is ""
		copy(newdoc, item)
		copy(old, doc)
		return len(item), len(doc)
	}
	return set(doc, cborptr, item, newdoc, old)
}

// Prepend item into a array or property container specified by cbor-pointer.
// Returns length of `newdoc`.
func (config *Config) Prepend(doc, cborptr, item, newdoc []byte) int {
	if !config.IsIndefiniteText(Indefinite(cborptr[0])) {
		panic(ErrorExpectedCborPointer)
	}
	return prepend(doc, cborptr, item, newdoc)
}

// Delete field or nested field specified by json pointer. Returns
// length of `newdoc` and `deleted` item.
func (config *Config) Delete(doc, cborptr, newdoc, deleted []byte) (int, int) {
	if !config.IsIndefiniteText(Indefinite(cborptr[0])) {
		panic(ErrorExpectedCborPointer)
	} else if cborptr[1] == brkstp { // json-pointer is ""
		panic(ErrorEmptyPointer)
	}
	return del(doc, cborptr, newdoc, deleted)
}
