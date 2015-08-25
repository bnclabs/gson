// Package collatejson supplies encoders and decoders to transform
// JSON text, or golang representation of JSON text, cbor encoded
// JSON text into binary representation without loosing information.
// That is,
//
// * binary representation should preserve the sort order such
//   that sorting binary encoded, using memcmp, JSON document
//   should follow a desired sort order.
// * it must be possible to get back the original document, in
//   semantically correct form, from its binary representation.
//
// Notes:
//
// * items in a property object are sorted by its property name
//   before they are compared with property's value.
//
// * strings are collated as it is received from the input without
//   un-quoting the JSON-string and without unicode collation.
package collate

import "errors"

// ErrorNumberType means configured number type is not supported.
var ErrorNumberType = errors.New("collatejson.numberType")

// ErrorOutputLen means output buffer has insufficient length.
var ErrorOutputLen = errors.New("collatejson.outputLen")

// Length is an internal type used for prefixing length
// of arrays and properties.
type Length int64

// Missing denotes a special type for an item that evaluates
// to _nothing_.
type Missing string

// MissingLiteral is special string to denote missing item.
// IMPORTANT: we are assuming that MissingLiteral will not
// occur in the keyspace.
const MissingLiteral = Missing("~[]{}falsenilNA~")

// NumberType to choose for number collation
type NumberType byte

const (
	// Float64 to collate input numbers as 64-bit floating point.
	Float64 NumberType = iota + 1
	// Int64 to collate input numbers as 64-bit signed-integer.
	Int64
	// Decimal to collate input numbers as N, where -1 < N < 1
	Decimal
)

// Config for new collation and de-collation.
type Config struct {
	arrayLenPrefix    bool       // first sort arrays based on its length
	propertyLenPrefix bool       // first sort properties based on length
	doMissing         bool       // handle missing values (for N1QL)
	nt                NumberType // encode numbers as "float64" or "int64" or "decimal"
	//-- unicode
	//backwards        bool
	//hiraganaQ        bool
	//caseLevel        bool
	//numeric          bool
	//nfkd              bool
	//utf8              bool
	//strength          colltab.Level
	//alternate         collate.AlternateHandling
	//language          language.Tag
}

// NewDefaultConfig creates a configuration instance with default
// parameters to collate and de-collate gson, json and cbor.
func NewDefaultConfig() *Config {
	return &Config{
		arrayLenPrefix:    false,
		propertyLenPrefix: true,
		doMissing:         true,
		nt:                Float64,
	}
}

// NewConfig creates a configuration instance to collate
// and de-collate gson, json and cbor.
func NewConfig(al, pl bool, nt NumberType) *Config {
	config := NewDefaultConfig()
	config.arrayLenPrefix = al
	config.propertyLenPrefix = pl
	config.nt = nt
	return config
}

// SortbyArrayLen sorts array by length before sorting by array
// elements. Use `false` to sort only by array elements.
// Default is `true`.
func (config *Config) SortbyArrayLen(what bool) *Config {
	config.arrayLenPrefix = what
	return config
}

// SortbyPropertyLen sorts property by length before sorting by
// property items. Use `false` to sort only by proprety items.
// Default is `true`.
func (config *Config) SortbyPropertyLen(what bool) *Config {
	config.propertyLenPrefix = what
	return config
}

// UseMissing will interpret special string MissingLiteral and
// encode them as TypeMissing.
// Default is `true`.
func (config *Config) UseMissing(what bool) *Config {
	config.doMissing = what
	return config
}

// NumberType chooses type of encoding / decoding for JSON
// numbers. Can be "float64", "int64", "decimal".
// Default is "float64"
func (config *Config) NumberType(what string) *Config {
	switch what {
	case "float64":
		config.nt = Float64
	case "int64":
		config.nt = Int64
	case "decimal":
		config.nt = Decimal
	}
	return config
}

// CollateGson encode input golang object to order preserving
// binary representation. `code` is the output buffer for
// encoding and expected to be adequately size.
func (config *Config) CollateGson(obj interface{}, code []byte) int {
	return gson2collate(obj, code, config)
}

// Gson will decode an already collated object back to golang
// representation of JSON.
func (config *Config) Gson(code []byte) (interface{}, int) {
	if len(code) == 0 {
		return nil, 0
	}
	return collate2gson(code, config)
}

// Equal checks wether n is MissingLiteral
func (m Missing) Equal(n string) bool {
	s := string(m)
	if len(n) == len(s) && n[0] == '~' && n[1] == '[' {
		return s == n
	}
	return false
}
