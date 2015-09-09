package gson

import "sync"
import "unsafe"
import "sync/atomic"

// maximum length of json pointer is restricted to 1024 bytes.
var prefixPool *sync.Pool

// pointer can contain a maximum of 1024 segments.
var segmentsPool *sync.Pool

// scratch pad for string objects
var stringPool *sync.Pool

// property keys
var keysPool *sync.Pool

func init() {
	prefixPool = &sync.Pool{
		New: func() interface{} {
			return make([]byte, 0, 1024) // TODO: no magic number
		},
	}
	segmentsPool = &sync.Pool{
		New: func() interface{} {
			return make([]string, 0, 1024) // TODO: no magic number
		},
	}
	stringPool = &sync.Pool{
		New: func() interface{} {
			return make([]byte, 1024*1024) // TODO: no magic number
		},
	}
	keysPool = &sync.Pool{
		New: func() interface{} {
			return make([]string, 0, 1024) // TODO: no magic number
		},
	}
}

type jsonPool struct {
	keypool  *sync.Pool
	codepool *sync.Pool
}

var jsonPools unsafe.Pointer = unsafe.Pointer(&map[int]jsonPool{})

func newJsonKeyPool(maxkeys int) jsonPool {
	keypool := &sync.Pool{New: func() interface{} {
		return make(kvrefs, maxkeys)
	}}
	codepool := &sync.Pool{New: func() interface{} {
		return make([]byte, maxkeys*1024)
	}}
	pool := jsonPool{keypool, codepool}
	for {
		opp := atomic.LoadPointer(&jsonPools)
		oldm := (*map[int]jsonPool)(opp) // type cast
		newm := map[int]jsonPool{maxkeys: pool}
		for k, pool := range *oldm {
			newm[k] = pool
		}
		if atomic.CompareAndSwapPointer(&jsonPools, opp, unsafe.Pointer(&newm)) {
			return pool
		}
	}
	panic("unreachable code")
}

func getJsonKeyPool(maxkeys int) jsonPool {
	m := *(*map[int]jsonPool)(atomic.LoadPointer(&jsonPools))
	if pool, ok := m[maxkeys]; ok {
		return pool
	}
	return newJsonKeyPool(maxkeys)
}
