package gson

import "sync"

// MaxStringLen maximum length of string value inside json document.
// Affects memory pool. Changing this value will affect all new
// configuration instances.
var MaxStringLen = 1024 * 1024

// MaxCollateLen maximum length of collated value. Affects memory pool.
// Changing this value will affect all new configuration instances.
var MaxCollateLen = 1024 * 1024

type memConfig struct {
	strlen  int // maximum length of string value inside JSON document
	numkeys int // maximum number of keys that a property object can have
	itemlen int // maximum length of collated value.
	ptrlen  int // maximum length of json-pointer can take
}

type mempools struct {
	keysPool *sync.Pool
	keypool  *sync.Pool
}

func newMempool(strlen, numkeys, itemlen, jptrlen int) mempools {
	m := mempools{}
	m.keysPool = &sync.Pool{
		New: func() interface{} { return make([]string, 0, numkeys) },
	}
	m.keypool = &sync.Pool{
		New: func() interface{} { return make(kvrefs, numkeys) },
	}
	return m
}
