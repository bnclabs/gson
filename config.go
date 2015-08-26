package gson

// Number placeholder type when number is represented in str format,
// used for delayed parsing.
type Number string

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

// Config and access gson functions. All APIs to gson is defined via
// config. To quickly get started, use NewDefaultConfig() that will
// create a configuration with default values.
type Config struct {
	Nk NumberKind
	Ws SpaceKind
}

// NewDefaultConfig returns a new configuration with default values.
// Nk: FloatNumber
// Ws: UnicodeSpace
func NewDefaultConfig() *Config {
	return NewConfig(FloatNumber, UnicodeSpace)
}

// NewConfig returns a new configuration.
func NewConfig(nk NumberKind, ws SpaceKind) *Config {
	return &Config{Nk: nk, Ws: ws}
}

// Parse input JSON text to a single go-native value. If text is
// invalid raises panic. Along with go-native value, remaining
// unparsed text is returned.
func (config *Config) Parse(txt string) (interface{}, string) {
	return scanToken(txt, config)
}

// ParseMany will parse input JSON text to one or more go native
// values. If text is invalid raises panic.
func (config *Config) ParseMany(txt string) []interface{} {
	var values []interface{}
	var tok interface{}
	for len(txt) > 0 {
		tok, txt = scanToken(txt, config)
		values = append(values, tok)
	}
	return values
}

// ParsePointer parse input JSON pointer into segments.
func (config *Config) ParsePointer(pointer string, segments []string) []string {
	return parsePointer(pointer, segments)
}

// EncodePointer compliments ParsePointer to convert parsed
// `segments` back to json-pointer. Converted pointer is available
// in the `pointer` array and returns the length of pointer.
func (config *Config) EncodePointer(segments []string, pointer []byte) int {
	return encodePointer(segments, pointer)
}

// ListPointers from value.
func (config *Config) ListPointers(value interface{}) []string {
	pointers := allpaths(value)
	pointers = append(pointers, "")
	return pointers
}

// Get field or nested field specified by json pointer.
func (config *Config) Get(ptr string, doc interface{}) (item interface{}) {
	segments := config.ParsePointer(ptr, []string{})
	return get(segments, doc)
}

// Set field or nested field specified by json pointer. If input
// `doc` is of type []interface{}, `newdoc` return the updated slice,
// along with the old value.
func (config *Config) Set(
	ptr string, doc, item interface{}) (newdoc, old interface{}) {

	segments := config.ParsePointer(ptr, []string{})
	return set(segments, doc, item)
}

// Delete field or nested field specified by json pointer.
// If input `doc` is of type []interface{}, `newdoc` return
// the updated slice, along with the deleted value.
func (config *Config) Delete(
	ptr string, doc interface{}) (newdoc, deleted interface{}) {

	segments := config.ParsePointer(ptr, []string{})
	return del(segments, doc)
}
