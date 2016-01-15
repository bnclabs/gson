//  Copyright (c) 2015 Couchbase, Inc.

// transform golang native value into json encoded value.
// cnf: -

package gson

func value2json(value interface{}, out []byte, config *Config) int {
	config.buf.Reset()
	if err := config.enc.Encode(value); err != nil {
		panic(err)
	}
	s := config.buf.Bytes()
	return copy(out, s[:len(s)-1]) // -1 to strip \n
}
