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

func newMempool(strlen, numkeys, itemlen, ptrlen int) mempools {
	m := mempools{}
	m.prefixPool = &sync.Pool{
		New: func() interface{} { return make([]byte, 0, ptrlen) },
	}
	m.segmentsPool = &sync.Pool{
		New: func() interface{} { return make([]string, 0, ptrlen) },
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
