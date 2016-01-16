//  Copyright (c) 2015 Couchbase, Inc.

package gson

// SpaceKind to skip white-spaces in JSON text.
type SpaceKind byte

const (
	// AnsiSpace will skip white space characters defined by ANSI spec.
	AnsiSpace SpaceKind = iota + 1

	// UnicodeSpace will skip white space characters defined by Unicode spec.
	UnicodeSpace
)

type jsonConfig struct {
	// if `strict` is false then configurations with IntNumber
	// will parse floating numbers and then convert it to int64.
	// else will panic when detecting floating numbers.
	strict     bool
	ws         SpaceKind
	jsonString bool
}

// SpaceKind setting to interpret whitespaces in json text.
func (config Config) SpaceKind(ws SpaceKind) *Config {
	config.ws = ws
	return &config
}

// Strict setting to enforce strict transforms.
// TODO: describe this more.
func (config Config) Strict(what bool) *Config {
	config.strict = what
	return &config
}

// JsonString settings to parse json string as it is, avoid un-quoting.
func (config Config) JsonString(what bool) *Config {
	config.jsonString = what
	return &config
}

// Json abstraction for value encoded as json text.
type Json struct {
	config *Config
	data   []byte
	n      int
}

// Bytes return the json encoded date.
func (jsn *Json) Bytes() []byte {
	return jsn.data[:jsn.n]
}

// Reset buffer to zero-length.
func (jsn *Json) Reset(data []byte) *Json {
	if data == nil {
		jsn.n = 0
		return jsn
	}
	jsn.data, jsn.n = data, len(data)
	return jsn
}

// Tovalue parse json text to golang native value. Return remaining text.
func (jsn *Json) Tovalue() (*Json, interface{}) {
	remaining, value := json2value(bytes2str(jsn.data[:jsn.n]), jsn.config)
	if remaining != "" {
		return jsn.config.NewJson(str2bytes(remaining), len(remaining)), value
	}
	return nil, value
}

// ToValues parse json text to one or more go native values.
func (jsn *Json) Tovalues() []interface{} {
	var values []interface{}
	var tok interface{}
	txt := bytes2str(jsn.data[:jsn.n])
	for len(txt) > 0 {
		txt, tok = json2value(txt, jsn.config)
		values = append(values, tok)
	}
	return values
}

// Tocbor convert json encoded value into cbor encoded binary string.
func (jsn *Json) Tocbor(cbr *Cbor) *Cbor {
	in := bytes2str(jsn.data[:jsn.n])
	_ /*remning*/, m := json2cbor(in, cbr.data[cbr.n:], jsn.config)
	cbr.n += m
	return cbr
}

// Tocollate convert json encoded value into binary-collation.
func (jsn *Json) Tocollate(clt *Collate) *Collate {
	in := bytes2str(jsn.data[:jsn.n])
	_ /*remn*/, m := json2collate(in, clt.data[clt.n:], jsn.config)
	clt.n += m
	return clt
}

func (ws SpaceKind) String() string {
	switch ws {
	case AnsiSpace:
		return "AnsiSpace"
	case UnicodeSpace:
		return "UnicodeSpace"
	default:
		panic("new space-kind")
	}
}
