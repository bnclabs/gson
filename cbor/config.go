package cbor

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
	// JsonNumber will store number in JSON encoding, can be used while
	// converting json to cbor.
	JsonNumber
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

// ContainerEncoding, encoding method to use for arrays and maps.
type ContainerEncoding byte

const (
	// LengthPrefix encoding for composite types. That is, for arrays and maps
	// encode the number of contained items as well.
	LengthPrefix ContainerEncoding = iota + 1
	// Stream encoding for composite types. That is, for arrays and maps
	// use cbor's indefinite and break-stop to encode member items.
	Stream
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
	nk NumberKind
	ws SpaceKind
	ct ContainerEncoding
}

// NewDefaultConfig returns a new configuration factory, with default
// values,
// NumberKind: FloatNumber
// SpaceKind: UnicodeSpace
// ContainerEncoding: Stream
func NewDefaultConfig() *Config {
	return NewConfig(FloatNumber, UnicodeSpace, Stream)
}

// NewConfig returns a new configuration factory
func NewConfig(nk NumberKind, ws SpaceKind, ct ContainerEncoding) *Config {
	return &Config{nk: nk, ws: ws, ct: ct}
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

// IsBreakcodeBytes can be used to check whether chunks of
// byte-strings are ending with the current byte.
// Can be used by libraries that build on top of cbor.
func (config *Config) IsBreakcodeBytes(b byte) bool {
	return b == hdr(type2, itemBreak)
}

// IsBreakcodeText can be used to check whether chunks of
// text are ending with the current byte.
// Can be used by libraries that build on top of cbor.
func (config *Config) IsBreakcodeText(b byte) bool {
	return b == hdr(type3, itemBreak)
}

// IsBreakcodeArray can be used to check whether array items
// of indefinite length are coming to an end with the current
// byte.
// Can be used by libraries that build on top of cbor.
func (config *Config) IsBreakcodeArray(b byte) bool {
	return b == hdr(type4, itemBreak)
}

// IsBreakcodeMap can be used to check whether map items
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

// CborEncodeMapItems to encode key,value pairs into cbor
func (config *Config) CborEncodeMapItems(items [][2]interface{}, out []byte) int {
	return encodeMapItems(items, out, config)
}

// Decode cbor binary into golang data.
func (config *Config) Decode(buf []byte) (interface{}, int) {
	return decode(buf)
}

// Parse input JSON text to cbor binary. Returns length of
// `out`.
func (config *Config) ParseJson(txt string, out []byte) (string, int) {
	return scanToCbor(txt, out, config)
}

// ToJson converts CBOR binary data-item into JSON. Returns
// length of `out`.
func (config *Config) ToJson(in, out []byte) (int, int) {
	return decodeTojson(in, out)
}

// FromJsonPointer converts json path in RFC-6901 into cbor format.
// Returns length of `out`.
func (config *Config) FromJsonPointer(jsonptr, out []byte) int {
	if len(jsonptr) > 0 && jsonptr[0] != '/' {
		panic("cbor expectedJsonPointer")
	}
	return fromJsonPointer(jsonptr, out)
}

// ToJsonPointer coverts cbor encoded path into json path RFC-6901.
// Returns length of `out`.
func (config *Config) ToJsonPointer(cborptr, out []byte) int {
	if !config.IsIndefiniteText(Indefinite(cborptr[0])) {
		panic("cbor expectedCborPointer")
	}
	return toJsonPointer(cborptr, out)
}

// Get field or nested field specified by cbor-pointer. Returns
// length of `item`.
func (config *Config) Get(doc, cborptr, item []byte) int {
	if !config.IsIndefiniteText(Indefinite(cborptr[0])) {
		panic("cbor expectedCborPointer")
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
		panic("cbor expectedCborPointer")
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
		panic("cbor expectedCborPointer")
	}
	return prepend(doc, cborptr, item, newdoc, config)
}

// Delete field or nested field specified by json pointer. Returns
// length of `newdoc` and `deleted` item.
func (config *Config) Delete(doc, cborptr, newdoc, deleted []byte) (int, int) {
	if !config.IsIndefiniteText(Indefinite(cborptr[0])) {
		panic("cbor expectedCborPointer")
	} else if cborptr[1] == brkstp { // json-pointer is ""
		panic("cbor emptyPointer")
	}
	return del(doc, cborptr, newdoc, deleted)
}
