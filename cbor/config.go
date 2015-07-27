package cbor

// NumberKind to parse JSON numbers.
type NumberKind byte

const (
	// SmartNumber will either use str.Atoi to parse JSON numbers or
	// fall back to float.
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
	UnicodeSpace
)

// Config for cbor interface.
type Config struct {
	Nk NumberKind
	Ws SpaceKind
}

// NewDefaultConfig returns a new configuration factory, with default
// values,
//      Nk: FloatNumber
//      Ws: SpaceKind
func NewDefaultConfig() *Config {
	return NewConfig(SmartNumber, UnicodeSpace)
}

// NewConfig returns a new configuration factory
func NewConfig(nk NumberKind, ws SpaceKind) *Config {
	return &Config{Nk: nk, Ws: ws}
}

// Encode golang data into cbor binary.
func (config *Config) Encode(item interface{}, buf []byte) int {
	return Encode(item, buf)
}

// Decode cbor binary into golang data.
func (config *Config) Decode(buf []byte) (interface{}, int) {
	return Decode(buf)
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
