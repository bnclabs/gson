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

const (
	// SmartNumber will either use str.Atoi to parse JSON numbers
	// or fall back to float. Default.
	SmartNumber NumberKind = iota + 1
	// IntNumber will use str.Atoi to parse JSON numbers.
	IntNumber
	// FloatNumber will use strconv.ParseFloat to parse JSON numbers.
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
	Nk NumberKind
	Ws SpaceKind
}

// NewDefaultConfig returns a new configuration factory, with default
// values,
//      Nk: FloatNumber
//      Ws: UnicodeSpace
func NewDefaultConfig() *Config {
	return NewConfig(SmartNumber, UnicodeSpace)
}

// NewConfig returns a new configuration factory
func NewConfig(nk NumberKind, ws SpaceKind) *Config {
	return &Config{Nk: nk, Ws: ws}
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

// Decode cbor binary into golang data.
func (config *Config) Decode(buf []byte) (interface{}, int) {
	return decode(buf)
}

// Parse input JSON text to cbor binary.
func (config *Config) ParseJson(txt string, out []byte) (string, int) {
	return scanToken(txt, out, config)
}

// ToJson converts CBOR binary data-item into JSON.
func (config *Config) ToJson(in, out []byte) (int, int) {
	n, m := cborTojson[in[0]](in, out)
	return n, m
}

// FromJsonPointer converts json path in RFC-6901 into cbor format,
func (config *Config) FromJsonPointer(path []byte, out []byte) int {
	var part [maxPartSize]byte

	if len(path) > 0 && path[0] != '/' {
		panic(ErrorExpectedJsonPointer)
	}

	n, off := encodeTextStart(out), 0
	for i := 0; i < len(path); {
		if path[i] == '~' {
			if path[i+1] == '1' {
				part[off] = '/'
				off, i = off+1, i+2

			} else if path[i+1] == '0' {
				part[off] = '~'
				off, i = off+1, i+2
			}

		} else if path[i] == '/' {
			if off > 0 {
				n += encodeTag(uint64(tagJsonString), out[n:])
				n += encodeText(bytes2str(part[:off]), out[n:])
				off = 0
			}
			i++

		} else {
			part[off] = path[i]
			i, off = i+1, off+1
		}
	}
	if off > 0 || (len(path) > 0 && path[len(path)-1] == '/') {
		n += encodeTag(uint64(tagJsonString), out[n:])
		n += encodeText(bytes2str(part[:off]), out[n:])
	}

	n += encodeBreakStop(out[n:])
	return n
}

// ToJsonPointer coverts cbor encoded path into json path RFC-6901
func (config *Config) ToJsonPointer(bin []byte, out []byte) int {
	if !config.IsIndefiniteText(Indefinite(bin[0])) {
		panic(ErrorExpectedCborPointer)
	}

	i, n, brkstp := 1, 0, hdr(type7, itemBreak)
	for {
		if bin[i] == hdr(type6, info24) && bin[i+1] == tagJsonString {
			i, out[n] = i+2, '/'
			n += 1
			ln, j := decodeLength(bin[i:])
			ln, i = ln+i+j, i+j
			for i < ln {
				switch bin[i] {
				case '/':
					out[n], out[n+1] = '~', '1'
					n += 2
				case '~':
					out[n], out[n+1] = '~', '0'
					n += 2
				default:
					out[n] = bin[i]
					n += 1
				}
				i++
			}
		}
		if bin[i] == brkstp {
			break
		}
	}
	return n
}
