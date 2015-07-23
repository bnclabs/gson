package cbor

import "fmt"

// NumberKind to parse JSON numbers.
type NumberKind byte

const (
	// IntNumber will use str.Atoi to parse JSON numbers.
	IntNumber NumberKind = iota + 1
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
	return NewConfig(FloatNumber, UnicodeSpace)
}

// NewConfig returns a new configuration factory
func NewConfig(nk NumberKind, ws SpaceKind) *Config {
	return &Config{Nk: nk, Ws: ws}
}

// Parse input JSON text to cbor binary.
func (config *Config) ParseJson(
	txt string, out []byte) (int, string, error) {

	n, remtxt, err := scanToken(txt, out, config)
	if err != nil {
		fmsg := "error `%v` before %v"
		err = fmt.Errorf(fmsg, err, len(txt)-len(remtxt))
		panic(err)
	}
	return n, remtxt, nil
}

// Encode golang data into cbor binary.
func (config *Config) Encode(item interface{}, buf []byte) int {
	return Encode(item, buf)
}

// Decode cbor binary into golang data.
func (config *Config) Decode(buf []byte) (interface{}, int) {
	return Decode(buf)
}
