//  Copyright (c) 2015 Couchbase, Inc.

package gson

import "sync"

type mempools struct {
	prefixPool   *sync.Pool // maximum length of json pointer
	segmentsPool *sync.Pool // number of segments in json-pointer
	stringPool   *sync.Pool // scratch pad for string objects
	keysPool     *sync.Pool // property keys

	keypool  *sync.Pool
	codepool *sync.Pool
}

func newMempool(strlen, numkeys, itemlen, jptrlen int) mempools {
	// strlen:  maximum length a string value can take in the JSON document.
	// numkeys: maximum number of keys that a property obj. can have.
	// itemlen: maximum length a collated value can take.
	// jptrlen: maximum length a json-pointer can take.
	m := mempools{}
	m.prefixPool = &sync.Pool{
		New: func() interface{} { return make([]byte, 0, jptrlen) },
	}
	m.segmentsPool = &sync.Pool{
		New: func() interface{} { return make([]string, 0, jptrlen) },
	}
	m.stringPool = &sync.Pool{
		New: func() interface{} { return make([]byte, strlen) },
	}
	m.keysPool = &sync.Pool{
		New: func() interface{} { return make([]string, 0, numkeys) },
	}
	m.keypool = &sync.Pool{
		New: func() interface{} { return make(kvrefs, numkeys) },
	}
	m.codepool = &sync.Pool{
		New: func() interface{} { return make([]byte, itemlen) },
	}
	return m
}
