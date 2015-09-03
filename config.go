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

// Config and access gson functions. All APIs to gson is defined via
// config. To quickly get started, use NewDefaultConfig() that will
// create a configuration with default values.
type Config struct {
	nk NumberKind
	ws SpaceKind
}

// NewDefaultConfig returns a new configuration with default values.
// NumberKind: FloatNumber
// SpaceKind: UnicodeSpace
func NewDefaultConfig() *Config {
	return NewConfig(FloatNumber, UnicodeSpace)
}

// NewConfig returns a new configuration.
func NewConfig(nk NumberKind, ws SpaceKind) *Config {
	return &Config{nk: nk, ws: ws}
}

// Parse input JSON text to a single go-native value. If text is
// invalid raises panic. Remaining unparsed text is returned,
// along with go-native value.
func (config *Config) Parse(txt string) (string, interface{}) {
	return scanValue(txt, config)
}

// ParseMany will parse input JSON text to one or more go native
// values. If text is invalid raises panic.
func (config *Config) ParseMany(txt string) []interface{} {
	var values []interface{}
	var tok interface{}
	for len(txt) > 0 {
		txt, tok = scanValue(txt, config)
		values = append(values, tok)
	}
	return values
}

// ParsePointer follows rfc-6901 allows ~0 and ~1 escapes, property
// lookup by specifying the key and array lookup by specifying the
// index. Also allows empty "" pointer and empty key "/".
func (config *Config) ParsePointer(pointer string, segments []string) []string {
	return parsePointer(pointer, segments)
}

// EncodePointer reverse of ParsePointer to convert parsed
// `segments` back to json-pointer. Converted pointer is available
// in the `pointer` array and returns the length of pointer-array.
func (config *Config) EncodePointer(segments []string, pointer []byte) int {
	return encodePointer(segments, pointer)
}

// ListPointers all possible pointers into object.
func (config *Config) ListPointers(object interface{}) []string {
	pointers := allpaths(object)
	pointers = append(pointers, "")
	return pointers
}

// Get field or nested field specified by json pointer.
func (config *Config) Get(ptr string, doc interface{}) (item interface{}) {
	segments := config.ParsePointer(ptr, []string{})
	return get(segments, doc)
}

// Set field or nested field specified by json pointer. While
// `newdoc` is gauranteed to contain the `item`, `doc` _may_ not be.
// Suggested usage,
//      doc := []interface{}{"hello"}
//      doc, _ = config.Set("/-", doc, "world")
func (config *Config) Set(ptr string, doc, item interface{}) (newdoc, old interface{}) {
	segments := config.ParsePointer(ptr, []string{})
	return set(segments, doc, item)
}

// Delete field or nested field specified by json pointer. While
// `newdoc` is gauranteed to be updated, `doc` _may_ not be.
// Suggested usage,
//      doc := []interface{}{"hello", "world"}
//      doc, _ = config.Delete("/1", doc)
func (config *Config) Delete(ptr string, doc interface{}) (newdoc, deleted interface{}) {
	segments := config.ParsePointer(ptr, []string{})
	return del(segments, doc)
}
