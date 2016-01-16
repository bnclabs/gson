package gson

const MaxJsonpointerLen = 2048
const MaxSegmentLen = 256

// Jsonpointer abstracts rfc-6901 into a type.
// allows ~0 and ~1 escapes, property lookup by specifying the key,
// and array lookup by specifying the index.
// Also allows empty "" pointer and empty key "/".
type Jsonpointer struct {
	path     []byte
	segments [][]byte
	segln    int
}

// NewJsonpointer to create a new instance of Jsonpointer allocate
// necessary memory from heap.
func NewJsonpointer(path string) *Jsonpointer {
	jptr := &Jsonpointer{
		path:     make([]byte, MaxJsonpointerLen),
		segments: make([][]byte, 0),
	}
	n := copy(jptr.path, path)
	jptr.path = jptr.path[:n]
	return jptr
}

// ResetPath to reuse the Jsonpointer object for a new path.
func (jptr *Jsonpointer) ResetPath(path string) *Jsonpointer {
	n := copy(jptr.path, path)
	jptr.path = jptr.path[:n]
	jptr.segln = 0
	return jptr
}

// ResetSegments variant of ResetPath to reconstruct the path from segments.
func (jptr *Jsonpointer) ResetSegments(segments []string) *Jsonpointer {
	n := encodePointer(segments, jptr.path[:cap(jptr.path)])
	jptr.path = jptr.path[:n]
	for i, segment := range segments {
		if i < len(jptr.segments) {
			jptr.segments[i] = append(jptr.segments[i][:0], segment...)
		} else {
			jptr.segments = append(jptr.segments, make([]byte, MaxSegmentLen))
			n := copy(jptr.segments[i], segment)
			jptr.segments[i] = jptr.segments[i][:n]
		}
	}
	jptr.segln = len(segments)
	return jptr
}

// Segments return path segments, segments in a path are separated by "/"
func (jptr *Jsonpointer) Segments() [][]byte {
	if len(jptr.path) > 0 && jptr.segln == 0 {
		jptr.segln = parsePointer(jptr.path, jptr.segments)
	}
	return jptr.segments[:jptr.segln]
}
