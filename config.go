//  Copyright (c) 2015 Couchbase, Inc.

// Package Gson provides a toolkit for JSON representation, collation
// and transformation.
//
// Package provides APIs to convert data representation from one format
// to another. Supported formats are:
//   * Json
//   * Golang value
//   * CBOR - Consice Binary Object Representation
//   * binary-collation
//
// CBOR:
//
// Package also provides a RFC-7049 (CBOR) implementation, to encode
// golang data into machine friendly binary format. Following golang
// native types are supported:
//   * nil, true, false.
//   * native integer types, and its alias, of all width.
//   * float32, float64.
//   * slice of bytes.
//   * native string.
//   * slice of interface - []interface{}.
//   * map of string to interface{} - map[string]interface{}.
//
// Types from golang's standard library and custom types provided
// by this package that can be encoded using CBOR:
//   * CborTagBytes: a cbor encoded []bytes treated as value.
//   * CborUndefined: encode a data-item as undefined.
//   * CborIndefinite: encode bytes, string, array and map of unspefied length.
//   * CborBreakStop: to encode end of CborIndefinite length item.
//   * CborTagEpoch: in seconds since epoch.
//   * CborTagEpochMicro: in micro-seconds epoch.
//   * CborTagFraction: m*(10**e)
//   * CborTagFloat: m*(2**e)
//   * CborTagPrefix: to self indentify a binary blog as CBOR.
//
// Package also provides an implementation for encoding json to CBOR
// and vice-versa:
//   * number can be encoded as integer or float.
//   * arrays and maps are encoded using indefinite encoding.
//   * byte-string encoding is not used.
//
// Json-Pointer:
//
// Package also provides a RFC-6901 (JSON-pointers) implementation.
package gson

import "bytes"
import "fmt"
import "encoding/json"

// NumberKind how to treat numbers.
type NumberKind byte

const (
	// SmartNumber32 to treat number as either integer or fall back to float32.
	SmartNumber32 NumberKind = iota + 1

	// SmartNumber to treat number as either integer or fall back to float64.
	SmartNumber

	// IntNumber to treat number as int64.
	IntNumber

	// FloatNumber32 to treat number as float32.
	FloatNumber32

	// FloatNumber to treat number as float64.
	FloatNumber

	// Decimal to collate input numbers as N, where -1 < N < 1
	Decimal
)

// MaxKeys maximum number of keys allowed in a property object.
const MaxKeys = 1024

// Config is the primary object to access the APIs exported by this package.
// Before calling any of the config-methods, make sure to initialize
// them with desired settings and don't change them afterwards.
type Config struct {
	nk      NumberKind
	maxKeys int
	pools   mempools

	cborConfig
	jsonConfig
	collateConfig
	jptrConfig
	memConfig

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

// NewDefaultConfig returns a new configuration with default settings:
//		FloatNumber         Stream
//		MaxKeys
//		UnicodeSpace        -strict
//		+doMissing          -arrayLenPrefix
//		+propertyLenPrefix
//		MaxJsonpointerLen
//		MaxStringLen        MaxKeys
//		MaxCollateLen       MaxJsonpointerLen
func NewDefaultConfig() *Config {
	config := &Config{
		nk:      FloatNumber,
		maxKeys: MaxKeys,
		cborConfig: cborConfig{
			ct: Stream,
		},
		jsonConfig: jsonConfig{
			ws:     UnicodeSpace,
			strict: false,
		},
		collateConfig: collateConfig{
			doMissing:         true,
			arrayLenPrefix:    false,
			propertyLenPrefix: true,
		},
		memConfig: memConfig{
			strlen:  MaxStringLen,
			numkeys: MaxKeys,
			itemlen: MaxCollateLen,
			ptrlen:  MaxJsonpointerLen,
		},
	}
	config = config.SetJptrlen(MaxJsonpointerLen)

	config.buf = bytes.NewBuffer(make([]byte, 0, 1024)) // start with 1K
	config.enc = json.NewEncoder(config.buf)
	a, b, c, d := config.strlen, config.numkeys, config.itemlen, config.ptrlen
	config.pools = newMempool(a, b, c, d)

	return config
}

// NumberKind setting to interpret number values.
func (config Config) SetNumberKind(nk NumberKind) *Config {
	config.nk = nk
	return &config
}

// SetContainerEncoding setting to encode / decode cbor arrays and maps.
func (config Config) SetContainerEncoding(ct ContainerEncoding) *Config {
	config.ct = ct
	return &config
}

// SetMaxkeys will set the maximum number of keys allowed in property item.
func (config Config) SetMaxkeys(n int) *Config {
	config.maxKeys = n
	return &config
}

// SetJptrlen will set the maximum size for jsonpointer path.
func (config Config) SetJptrlen(n int) *Config {
	config.jptrMaxlen = n
	config.jptrMaxseg = n / 8
	return &config
}

// ResetPools will create a new set of pools with specified size.
//	   strlen  - maximum length of string value inside JSON document
//	   numkeys - maximum number of keys that a property object can have
//	   itemlen - maximum length of collated value.
//	   ptrlen  - maximum length of json-pointer can take
func (config Config) ResetPools(strlen, numkeys, itemlen, ptrlen int) *Config {
	config.memConfig = memConfig{
		strlen: strlen, numkeys: numkeys, itemlen: itemlen, ptrlen: ptrlen,
	}
	config.pools = newMempool(strlen, numkeys, itemlen, ptrlen)
	return &config
}

// NewCbor create a new Cbor instance.
func (config *Config) NewCbor(buffer []byte, ln int) *Cbor {
	if ln == -1 {
		ln = len(buffer)
	}
	return &Cbor{config: config, data: buffer, n: ln}
}

// NewJson create a new Json instance.
func (config *Config) NewJson(buffer []byte, ln int) *Json {
	if ln == -1 {
		ln = len(buffer)
	}
	return &Json{config: config, data: buffer, n: ln}
}

// NewCollate create a new Collate instance.
func (config *Config) NewCollate(buffer []byte, ln int) *Collate {
	if ln == -1 {
		ln = len(buffer)
	}
	return &Collate{config: config, data: buffer, n: ln}
}

// NewValue create a new Value instance.
func (config *Config) NewValue(value interface{}) *Value {
	return &Value{config: config, data: value}
}

// NewJsonpointer create a instance of Jsonpointer allocate necessary memory.
func (config *Config) NewJsonpointer(path string) *Jsonpointer {
	if len(path) > config.jptrMaxlen {
		panic("jsonpointer path exceeds configured length")
	}
	jptr := &Jsonpointer{
		config:   config,
		path:     make([]byte, config.jptrMaxlen+16),
		segments: make([][]byte, config.jptrMaxseg),
	}
	for i := 0; i < config.jptrMaxseg; i++ {
		jptr.segments[i] = make([]byte, 0, 16)
	}
	n := copy(jptr.path, path)
	jptr.path = jptr.path[:n]
	return jptr
}

func (config *Config) String() string {
	return fmt.Sprintf(
		"nk:%v, ws:%v, ct:%v, arrayLenPrefix:%v, "+
			"propertyLenPrefix:%v, doMissing:%v, maxKeys:%v",
		config.nk, config.ws, config.ct,
		config.arrayLenPrefix, config.propertyLenPrefix,
		config.doMissing, config.maxKeys)
}

func (nk NumberKind) String() string {
	switch nk {
	case SmartNumber32:
		return "SmartNumber32"
	case SmartNumber:
		return "SmartNumber"
	case IntNumber:
		return "IntNumber"
	case FloatNumber32:
		return "FloatNumber32"
	case FloatNumber:
		return "FloatNumber"
	case Decimal:
		return "Decimal"
	default:
		panic("new number-kind")
	}
}

func (ct ContainerEncoding) String() string {
	switch ct {
	case LengthPrefix:
		return "LengthPrefix"
	case Stream:
		return "Stream"
	default:
		panic("new space-kind")
	}
}
