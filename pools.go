package gson

import "sync"

// MaxStringLen maximum length of string value inside json document. Affects
// memory pool. Changing this value will affect all new configuration objects.
var MaxStringLen = 1024 * 1024

// MaxCollateLen maximum length of collated value. Affects memory pool.
// Changing this value will affect all new configuration objects.
var MaxCollateLen = 1024 * 1024

type memConfig struct {
	strlen  int // maximum length of string value inside JSON document
	numkeys int // maximum number of keys that a property object can have
	itemlen int // maximum length of collated value.
	ptrlen  int // maximum length of json-pointer can take
}

type mempools struct {
	prefixPool *sync.Pool // for computing list of possible pointers
	stringPool *sync.Pool
	keysPool   *sync.Pool
	keypool    *sync.Pool
	codepool   *sync.Pool
}

func newMempool(strlen, numkeys, itemlen, jptrlen int) mempools {
	m := mempools{}
	m.prefixPool = &sync.Pool{
		New: func() interface{} { return make([]byte, 0, jptrlen) },
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
