package gson

import "strconv"
import "unicode/utf8"

func parsePointer(in string) []string {
	segments := make([]string, 0, 2)
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

func allpaths(value interface{}) []string {
	pointers := make([]string, 0, 4)

	switch v := value.(type) {
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
