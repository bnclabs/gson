package gson

import "strconv"
import "unicode"
import "encoding/json"
import "unicode/utf8"
import "unicode/utf16"

// primary interface to scan JSON text and return,
// a. text remaining to be parsed.
// b. as go-native value.
// calling this function will scan for exactly one JSON value
func scanToken(txt string, config *Config) (string, interface{}) {
	txt = skipWS(txt, config.ws)

	if len(txt) < 1 {
		panic("gson scanner jsonEmpty")
	}

	if digitCheck[txt[0]] == 1 {
		return scanNum(txt, config.nk)
	}

	switch txt[0] {
	case 'n':
		if len(txt) >= 4 && txt[:4] == "null" {
			return txt[4:], nil
		}
		panic("gson scanner expectedNil")

	case 't':
		if len(txt) >= 4 && txt[:4] == "true" {
			return txt[4:], true
		}
		panic("gson scanner expectedTrue")

	case 'f':
		if len(txt) >= 5 && txt[:5] == "false" {
			return txt[5:], false
		}
		panic("gson scanner expectedFalse")

	case '"':
		s, remtxt := scanString(str2bytes(txt))
		return bytes2str(remtxt), s

	case '[':
		if txt = skipWS(txt[1:], config.ws); len(txt) == 0 {
			panic("gson scanner expectedCloseArray")
		} else if txt[0] == ']' {
			return txt[1:], []interface{}{}
		}
		arr := make([]interface{}, 0, len(txt)/10)
		for {
			var tok interface{}
			txt, tok = scanToken(txt, config)
			arr = append(arr, tok)
			if txt = skipWS(txt, config.ws); len(txt) == 0 {
				panic("gson scanner expectedCloseArray")
			} else if txt[0] == ',' {
				txt = skipWS(txt[1:], config.ws)
			} else if txt[0] == ']' {
				break
			} else {
				panic("gson scanner expectedCloseArray")
			}
		}
		return txt[1:], arr

	case '{':
		if txt = skipWS(txt[1:], config.ws); len(txt) == 0 {
			panic("gson scanner expectedCloseobject")
		} else if txt[0] == '}' {
			return txt[1:], map[string]interface{}{}
		} else if txt[0] != '"' {
			panic("gson scanner expectedKey")
		}
		m := make(map[string]interface{})
		for {
			var tok interface{}
			// NOTE: empty string is also a valid key
			key, remtxt := scanString(str2bytes(txt))
			txt = bytes2str(remtxt)

			if txt = skipWS(txt, config.ws); len(txt) == 0 || txt[0] != ':' {
				panic("gson scanner expectedColon")
			}
			txt, tok = scanToken(skipWS(txt[1:], config.ws), config)
			m[key] = tok
			if txt = skipWS(txt, config.ws); len(txt) == 0 {
				panic("gson scanner expectedCloseobject")
			} else if txt[0] == ',' {
				txt = skipWS(txt[1:], config.ws)
			} else if txt[0] == '}' {
				break
			} else {
				panic("gson scanner expectedCloseobject")
			}
		}
		return txt[1:], m
	}
	panic("gson scanner expectedToken")
}

var spaceCode = [256]byte{ // TODO: size can be optimized
	'\t': 1,
	'\n': 1,
	'\v': 1,
	'\f': 1,
	'\r': 1,
	' ':  1,
}

func skipWS(txt string, ws SpaceKind) string {
	switch ws {
	case UnicodeSpace:
		for i, ch := range txt {
			if unicode.IsSpace(ch) {
				continue
			}
			return txt[i:]
		}
		return ""

	case AnsiSpace:
		i := 0
		for i < len(txt) && spaceCode[txt[i]] == 1 {
			i++
		}
		txt = txt[i:]
	}
	return txt
}

func scanNum(txt string, nk NumberKind) (string, interface{}) {
	s, e, l := 0, 1, len(txt)
	if len(txt) > 1 {
		for ; e < l && intCheck[txt[e]] == 1; e++ {
		}
	}

	switch nk {
	case StringNumber:
		return txt[e:], json.Number(txt[s:e])

	case IntNumber:
		num, err := strconv.Atoi(string(txt[s:e]))
		if err != nil {
			panic("gson scanner expectedJsonInteger")
		}
		return txt[e:], num
	}
	// FloatNumber
	// NOTE: ignore the error because we have only picked
	// valid text to parse.
	num, _ := strconv.ParseFloat(string(txt[s:e]), 64)
	return txt[e:], num
}

var escapeCode = [256]byte{ // TODO: size can be optimized
	'"':  '"',
	'\\': '\\',
	'/':  '/',
	'\'': '\'',
	'b':  '\b',
	'f':  '\f',
	'n':  '\n',
	'r':  '\r',
	't':  '\t',
}

func scanString(txt []byte) (string, []byte) {
	if len(txt) < 2 {
		panic("gson scanner expectedString")
	}

	e := 1
	for txt[e] != '"' {
		c := txt[e]
		if c == '\\' || c == '"' || c < ' ' {
			break
		}
		if c < utf8.RuneSelf {
			e++
			continue
		}
		r, size := utf8.DecodeRune(txt[e:])
		if r == utf8.RuneError && size == 1 {
			break
		}
		e += size
		if e == len(txt) {
			panic("gson scanner expectedString")
		}
	}

	if txt[e] == '"' { // done we have nothing to unquote
		return string(txt[1:e]), txt[e+1:]
	}

	out := make([]byte, (len(txt)+2)*utf8.UTFMax)
	oute := copy(out, txt[:e]) // copy so far

loop:
	for e < len(txt) {
		switch c := txt[e]; {
		case c == '"':
			out[oute] = c
			e++
			break loop

		case c == '\\':
			if txt[e+1] == 'u' {
				r := getu4(txt[e:])
				if r < 0 { // invalid
					panic("gson scanner expectedString")
				}
				e += 6
				if utf16.IsSurrogate(r) {
					nextr := getu4(txt[e:])
					dec := utf16.DecodeRune(r, nextr)
					if dec != unicode.ReplacementChar { // A valid pair consume
						oute += utf8.EncodeRune(out[oute:], dec)
						e += 6
						break loop
					}
					// Invalid surrogate; fall back to replacement rune.
					r = unicode.ReplacementChar
				}
				oute += utf8.EncodeRune(out[oute:], r)

			} else { // escaped with " \ / ' b f n r t
				out[oute] = escapeCode[txt[e+1]]
				e += 2
				oute++
			}

		case c < ' ': // control character is invalid
			panic("gson scanner expectedString")

		case c < utf8.RuneSelf: // ASCII
			out[oute] = c
			oute++
			e++

		default: // coerce to well-formed UTF-8
			r, size := utf8.DecodeRune(txt[e:])
			e += size
			oute += utf8.EncodeRune(out[oute:], r)
		}
	}

	if out[oute] == '"' {
		return bytes2str(out[1:oute]), txt[e:]
	}
	panic("gson scanner expectedString")
}

// getu4 decodes \uXXXX from the beginning of s, returning the hex value,
// or it returns -1.
func getu4(s []byte) rune {
	if len(s) < 6 || s[0] != '\\' || s[1] != 'u' {
		return -1
	}
	r, err := strconv.ParseUint(string(s[2:6]), 16, 64)
	if err != nil {
		return -1
	}
	return rune(r)
}

var intCheck = [256]byte{}
var digitCheck = [256]byte{}

func init() {
	for i := 48; i <= 57; i++ {
		intCheck[i] = 1
	}
	intCheck['-'] = 1
	intCheck['+'] = 1
	intCheck['.'] = 1
	intCheck['e'] = 1
	intCheck['E'] = 1

	for i := 48; i <= 57; i++ {
		digitCheck[i] = 1
	}
	digitCheck['-'] = 1
	digitCheck['+'] = 1
	digitCheck['.'] = 1
}
