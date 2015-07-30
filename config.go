package gson

// NumberKind to parse JSON numbers.
type NumberKind byte

const (
	// StringNumber will keep the number text as is.
	StringNumber NumberKind = iota + 1
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

// Config for document factory
type Config struct {
	Nk NumberKind
	Ws SpaceKind
}

// NewDefaultConfig returns a new configuration factory
func NewDefaultConfig() *Config {
	return NewConfig(FloatNumber, UnicodeSpace)
}

// NewConfig returns a new configuration factory
func NewConfig(nk NumberKind, ws SpaceKind) *Config {
	return &Config{Nk: nk, Ws: ws}
}

// Parse input JSON text to a single go-native value.
func (config *Config) Parse(txt string) (interface{}, string) {
	return scanToken(txt, config)
}

// ParseMany will parse input JSON text to one or more go native
// values.
func (config *Config) ParseMany(txt string) ([]interface{}, string) {
	var values []interface{}
	var tok interface{}
	for len(txt) > 0 {
		tok, txt = scanToken(txt, config)
		values = append(values, tok)
	}
	return values, txt
}
