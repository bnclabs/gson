// implements rfc-6901 to parse json-pointer text and
// work with golang arrays and maps

package gson

import "strconv"
import "errors"
import "unicode/utf8"

// ErrorInvalidPointer supplied pointer is not valid.
var ErrorInvalidPointer = errors.New("gson.invalidPointer")

// ErrorInvalidIndex supplied pointer index to array is not valid.
var ErrorInvalidIndex = errors.New("gson.invalidIndex")

// ErrorOutofRange supplied pointer index to array is out of bounds.
var ErrorOutofRange = errors.New("gson.outofRange")

// ErrorInvalidKey supplied pointer index to map is not valid.
var ErrorInvalidKey = errors.New("gson.invalidKey")

func parsePointer(in string, segments []string) []string {
	if len(in) == 0 {
		return segments
	}

	i, j, s, u, part := 0, 0, str2bytes(in), [6]byte{}, [2048]byte{}
	for i < len(s) {
		if s[i] == '~' {
			if s[i+1] == '1' {
				part[j] = '/'
				i, j = i+2, j+1

			} else if s[i+1] == '0' {
				part[j] = '~'
				i, j = i+2, j+1
			}

		} else if s[i] == '/' {
			if j > 0 {
				segments = append(segments, string(part[:j]))
				j = 0
			}
			i++

		} else if s[i] < utf8.RuneSelf {
			part[j] = s[i]
			i, j = i+1, j+1

		} else {
			r, size := utf8.DecodeRune(s[i:])
			sizej := utf8.EncodeRune(u[:], r)
			copy(part[j:], u[:sizej])
			i, j = i+size, j+sizej
		}
	}
	if s[len(s)-1] == '/' || j > 0 {
		segments = append(segments, string(part[:j]))
	}
	return segments
}

func encodePointer(p []string, out []byte) int {
	n := 0
	for _, s := range p {
		out[n] = '/'
		n++
		for _, c := range []byte(s) {
			switch c {
			case '/':
				out[n] = '~'
				out[n+1] = '1'
				n += 2
			case '~':
				out[n] = '~'
				out[n+1] = '0'
				n += 2
			default:
				out[n] = c
				n++
			}
		}
	}
	return n
}

func allpaths(doc interface{}) []string {
	pointers := make([]string, 0, 4)

	switch v := doc.(type) {
	case []interface{}:
		if len(v) > 0 {
			for i, val := range v {
				prefix := "/" + strconv.Itoa(i)
				pointers = append(pointers, prefix)
				for _, pointer := range allpaths(val) {
					pointers = append(pointers, prefix+pointer)
				}
			}
		}
		pointers = append(pointers, "/-")

	case map[string]interface{}:
		if len(v) > 0 {
			for key, val := range v {
				prefix := "/" + key
				pointers = append(pointers, prefix)
				for _, pointer := range allpaths(val) {
					pointers = append(pointers, prefix+pointer)
				}
			}
		}
		pointers = append(pointers, "/-")
	}
	return pointers
}

func get(segments []string, doc interface{}) interface{} {
	if len(segments) == 0 { // exit recursion.
		return doc
	}

	switch val := doc.(type) {
	case []interface{}:
		if segments[0] == "-" {
			return get(segments[1:], val[len(val)-1])
		} else if idx, err := strconv.Atoi(segments[0]); err != nil {
			panic(ErrorInvalidIndex)
		} else if idx >= len(val) {
			panic(ErrorOutofRange)
		} else {
			return get(segments[1:], val[idx])
		}
	case map[string]interface{}:
		if doc, ok := val[segments[0]]; !ok {
			panic(ErrorInvalidKey)
		} else {
			return get(segments[1:], doc)
		}
	default:
		panic(ErrorInvalidPointer)
	}
}

func set(segments []string, doc, item interface{}) (newdoc, old interface{}) {
	ln := len(segments)
	container := doc // if ln == 1
	if ln == 0 {
		panic(ErrorInvalidPointer)
	} else if ln > 1 {
		container = get(segments[:ln-1], doc)
	}

	key := segments[ln-1]

	var ok bool
	switch cont := container.(type) {
	case []interface{}:
		if key == "-" {
			old = item
			cont = append(cont, item)
			if ln > 1 {
				set(segments[:ln-1], doc, cont)
			} else { // edge case !
				return cont, item
			}
		} else if idx, err := strconv.Atoi(key); err != nil {
			panic(ErrorInvalidIndex)
		} else if idx >= len(cont) {
			panic(ErrorOutofRange)
		} else {
			old, cont[idx] = cont[idx], item
		}
	case map[string]interface{}:
		if old, ok = cont[key]; !ok {
			old = item
		}
		cont[key] = item
	default:
		panic(ErrorInvalidPointer)
	}
	return doc, old
}

func del(segments []string, doc interface{}) (newdoc, old interface{}) {
	ln := len(segments)
	container := doc // if ln == 1
	if ln == 0 {
		panic(ErrorInvalidPointer)
	} else if ln > 1 {
		container = get(segments[:ln-1], doc)
	}

	key := segments[ln-1]

	switch cont := container.(type) {
	case []interface{}:
		if idx, err := strconv.Atoi(key); err != nil {
			panic(ErrorInvalidIndex)
		} else if idx >= len(cont) {
			panic(ErrorOutofRange)
		} else {
			old = cont[idx]
			copy(cont[idx:], cont[idx+1:])
			cont = cont[:len(cont)-1]
			if ln > 1 {
				set(segments[:ln-1], doc, cont)
			} else { // edge case !!
				return cont, old
			}
		}
	case map[string]interface{}:
		old, _ = cont[key]
		delete(cont, key)
	default:
		panic(ErrorInvalidPointer)
	}
	return doc, old
}
