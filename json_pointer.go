//  Copyright (c) 2015 Couchbase, Inc.

// implements rfc-6901 to parse json-pointer text and
// work with golang arrays and maps

package gson

import "strconv"
import "strings"
import "unicode/utf8"

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
		for _, c := range str2bytes(s) {
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

func allpaths(doc interface{}, pointers []string, prefix []byte) []string {
	var scratch [64]byte

	n := len(prefix)
	prefix = append(prefix, '/', '-')
	switch v := doc.(type) {
	case []interface{}:
		pointers = append(pointers, string(prefix)) // new allocation
		if len(v) > 0 {
			for i, val := range v {
				prefix = prefix[:n]
				dst := strconv.AppendInt(scratch[:0], int64(i), 10)
				prefix = append(prefix, '/')
				prefix = append(prefix, dst...)
				pointers = append(pointers, string(prefix)) // new allocation
				pointers = allpaths(val, pointers, prefix)
			}
		}

	case map[string]interface{}:
		pointers = append(pointers, string(prefix)) // new allocation
		if len(v) > 0 {
			for key, val := range v {
				prefix = prefix[:n]
				prefix = append(prefix, '/')
				prefix = append(prefix, str2bytes(escapeJp(key))...)
				pointers = append(pointers, string(prefix)) // new allocation
				pointers = allpaths(val, pointers, prefix)
			}
		}

	case [][2]interface{}:
		pointers = append(pointers, string(prefix)) // new allocation
		if len(v) > 0 {
			for _, pairs := range v {
				prefix = prefix[:n]
				key, val := pairs[0].(string), pairs[1]
				prefix = append(prefix, '/')
				prefix = append(prefix, str2bytes(escapeJp(key))...)
				pointers = append(pointers, string(prefix)) // new allocation
				pointers = allpaths(val, pointers, prefix)
			}
		}

	}
	return pointers
}

func escapeJp(key string) string {
	if strings.ContainsAny(key, "~/") {
		return strings.Replace(strings.Replace(key, "~", "~0", -1), "/", "~1", -1)
	}
	return key
}
