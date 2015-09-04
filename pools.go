package gson

import "sync"

// maximum length of json pointer is restricted to 1024 bytes.
var prefixPool *sync.Pool

// pointer can contain a maximum of 1024 segments.
var segmentsPool *sync.Pool

// scratch pad for string objects
var stringPool *sync.Pool

func init() {
	prefixPool = &sync.Pool{
		New: func() interface{} {
			return make([]byte, 0, 1024) // TODO: export 1024
		},
	}
	segmentsPool = &sync.Pool{
		New: func() interface{} {
			return make([]string, 0, 1024) // TODO: export 1024
		},
	}
	stringPool = &sync.Pool{
		New: func() interface{} {
			return make([]byte, 1024*1024) // TODO: export 1024
		},
	}
}
