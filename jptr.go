package gson

// MaxJsonpointerLen size of json-pointer path
const MaxJsonpointerLen = 2048

type jptrConfig struct {
	jptrMaxlen int
	jptrMaxseg int
}

// Jsonpointer abstracts rfc-6901 into a type.
// allows ~0 and ~1 escapes, property lookup by specifying the key,
// and array lookup by specifying the index.
// Also allows empty "" pointer and empty key "/".
type Jsonpointer struct {
	config   *Config
	path     []byte
	segments [][]byte
	segln    int
}

// ResetPath to reuse the Jsonpointer object for a new path.
func (jptr *Jsonpointer) ResetPath(path string) *Jsonpointer {
	if len(path) > jptr.config.jptrMaxlen {
		panic("jsonpointer path exceeds configured length")
	}
	n := copy(jptr.path[:cap(jptr.path)], path)
	jptr.path = jptr.path[:n]
	jptr.segln = 0
	return jptr
}

// ResetSegments variant of ResetPath to reconstruct the path from segments.
func (jptr *Jsonpointer) ResetSegments(segments []string) *Jsonpointer {
	if len(segments) > jptr.config.jptrMaxseg {
		panic("no. of segments in jsonpointer-path exceeds configured limit")
	}
	n := encodePointer(segments, jptr.path[:cap(jptr.path)])
	jptr.path = jptr.path[:n]
	for i, segment := range segments {
		jptr.segments[i] = append(jptr.segments[i][:0], segment...)
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

// Path return the path value.
func (jptr *Jsonpointer) Path() []byte {
	return jptr.path
}
