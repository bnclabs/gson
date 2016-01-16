//  Copyright (c) 2015 Couchbase, Inc.

// implements rfc-6901 to parse json-pointer text and
// work with golang arrays and maps

package gson

import "strconv"
import "strings"
import "unicode/utf8"

func parsePointer(in []byte, segments [][]byte) int {
	if len(in) == 0 {
		return 0
	}

	updateseg := func(segment []byte, j int) ([]byte, int) {
		segments[j] = segment
		j++
		return segments[j][:0], j
	}

	var j int
	var ch rune

	u, segment, escape := [6]byte{}, segments[j][:0], false

	for _, ch = range bytes2str(in) {
		if ch == '~' {
			escape = true

		} else if escape {
			switch ch {
			case '1':
				segment = append(segment, '/')
			case '0':
				segment = append(segment, '~')
			}
			escape = false

		} else if ch == '/' {
			segment, j = updateseg(segment, j)

		} else if ch < utf8.RuneSelf {
			segment = append(segment, byte(ch))

		} else {
			sz := utf8.EncodeRune(u[:], ch)
			segment = append(segment, u[:sz]...)
		}
	}
	segment, j = updateseg(segment, j)
	if in[len(in)-1] == '/' {
		_, j = updateseg(segment, j)
	}
	return j
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
				prefix = append(prefix, escapeJp(key)...)
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
				prefix = append(prefix, escapeJp(key)...)
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
