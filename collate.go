//  Copyright (c) 2015 Couchbase, Inc.

package gson

import "encoding/json"
import "bytes"

// Collation order for supported types, to change the order set these
// values in your init() function.
var (
	Terminator  byte = 0
	TypeMissing byte = 1
	TypeNull    byte = 2
	TypeFalse   byte = 3
	TypeTrue    byte = 4
	TypeNumber  byte = 5
	TypeString  byte = 6
	TypeLength  byte = 7
	TypeArray   byte = 8
	TypeObj     byte = 9
	TypeBinary  byte = 10
)

// Missing denotes a special type for an item that evaluates
// to _nothing_, used for collation.
type Missing string

// MissingLiteral is special string to denote missing item:
// IMPORTANT: we are assuming that MissingLiteral will not occur in
// the keyspace.
const MissingLiteral = Missing("~[]{}falsenilNA~")

type collateConfig struct {
	doMissing         bool // handle missing values (for N1QL)
	arrayLenPrefix    bool // first sort arrays based on its length
	propertyLenPrefix bool // first sort properties based on length
	enc               *json.Encoder
	buf               *bytes.Buffer
}

// SortbyArrayLen setting to sort array of smaller-size before larger ones.
func (config Config) SortbyArrayLen(what bool) *Config {
	config.arrayLenPrefix = what
	return &config
}

// SortbyPropertyLen setting to sort properties of smaller size before
// larger ones.
func (config Config) SortbyPropertyLen(what bool) *Config {
	config.propertyLenPrefix = what
	return &config
}

// UseMissing setting to use TypeMissing collation.
func (config Config) UseMissing(what bool) *Config {
	config.doMissing = what
	return &config
}

// Collate abstraction for value encoded into binary-collation.
type Collate struct {
	config *Config
	data   []byte
	n      int
}

// Bytes return a reference slice to encapsulated buffer.
func (clt *Collate) Bytes() []byte {
	return clt.data[:clt.n]
}

// Reset buffer to zero-length.
func (clt *Collate) Reset(data []byte) *Collate {
	if data == nil {
		clt.n = 0
		return clt
	}
	clt.data, clt.n = data, len(data)
	return clt
}

// Tovalue convert to golang native value.
func (clt *Collate) Tovalue() interface{} {
	if clt.n == 0 {
		return nil
	}
	value, _ /*rb*/ := collate2gson(clt.data[:clt.n], clt.config)
	return value
}

// Tojson convert to json encoded text.
func (clt *Collate) Tojson(jsn *Json) *Json {
	if clt.n == 0 {
		return nil
	}
	in := clt.data[:clt.n]
	_ /*rb*/, m /*wb*/ := collate2json(in, jsn.data[jsn.n:], clt.config)
	jsn.n += m
	return jsn
}

// Tocbor convert to cbor encoded value.
func (clt *Collate) Tocbor(cbr *Cbor) *Cbor {
	if clt.n == 0 {
		return nil
	}
	in := clt.data[:clt.n]
	_ /*rb*/, m /*wb*/ := collate2cbor(in, cbr.data[cbr.n:], clt.config)
	cbr.n += m
	return cbr
}

// Equal checks wether n is MissingLiteral
func (m Missing) Equal(n string) bool {
	s := string(m)
	if len(n) == len(s) && n[0] == '~' && n[1] == '[' {
		return s == n
	}
	return false
}
