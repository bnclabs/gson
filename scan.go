package gson

import "strconv"
import "unicode"
import "unicode/utf8"
import "unicode/utf16"

type Number string

var nullLiteral = "null"
var trueLiteral = "true"
var falseLiteral = "false"

// primary interface to scan JSON text and return,
// a. as go-native value.
// b. text remaining to be parsed.
// c. error in the i/p text.
// calling this function will scan for exactly one JSON value
func scanToken(txt string, config *Config) (interface{}, string, error) {
	txt = skipWS(txt, config.Ws)

	if len(txt) < 1 {
		return nil, txt, ErrorEmptyText
	}

	if digitCheck[txt[0]] == 1 {
		return scanNum(txt, config.Nk)
	}

	var tok interface{}
	var err error

	switch txt[0] {
	case 'n':
		if txt[:4] == nullLiteral {
			return nil, txt[4:], nil
		}
		return nil, txt, ErrorExpectedNil

	case 't':
		if txt[:4] == trueLiteral {
			return true, txt[4:], nil
		}
		return nil, txt, ErrorExpectedTrue

	case 'f':
		if txt[:5] == falseLiteral {
			return false, txt[5:], nil
		}
		return nil, txt, ErrorExpectedFalse

	case '-':
		return scanNum(txt, config.Nk)

	case '"':
		s, remtxt, err := scanString(str2bytes(txt))
		return bytes2str(s), bytes2str(remtxt), err

	case '[':
		txt = skipWS(txt[1:], config.Ws)
		if txt[0] == ']' {
			return []interface{}{}, txt[1:], nil
		}
		arr := make([]interface{}, 0, len(txt)/10)
		for {
			tok, txt, err = scanToken(txt, config)
			if err != nil {
				return nil, txt, err
			}
			arr = append(arr, tok)
			txt = skipWS(txt, config.Ws)
			if txt[0] == ',' {
				txt = skipWS(txt[1:], config.Ws)
			} else if txt[0] == ']' {
				break
			} else {
				return nil, txt, ErrorExpectedClosearray
			}
		}
		return arr, txt[1:], nil

	case '{':
		txt = skipWS(txt[1:], config.Ws)
		if txt[0] == '}' {
			return map[string]interface{}{}, txt[1:], nil
		}
		m := make(map[string]interface{})
		for {
			s, remtxt, err := scanString(str2bytes(txt))
			key := bytes2str(s)
			txt = bytes2str(remtxt)
			if err != nil {
				return nil, txt, err
			} else if len(key) < 1 {
				return nil, txt, ErrorExpectedKey
			}

			txt = skipWS(txt, config.Ws)
			if txt[0] != ':' {
				return nil, txt, ErrorExpectedColon
			}
			tok, txt, err = scanToken(skipWS(txt[1:], config.Ws), config)
			if err != nil {
				return nil, txt, err
			}
			m[key] = tok
			txt = skipWS(txt, config.Ws)
			if txt[0] == ',' {
				txt = skipWS(txt[1:], config.Ws)
			} else if txt[0] == '}' {
				break
			} else {
				return nil, txt, ErrorExpectedCloseobject
			}
		}
		return m, txt[1:], nil

	default:
		return nil, txt, ErrorExpectedToken
	}
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
		for spaceCode[txt[0]] == 1 {
			txt = txt[1:]
		}
	}
	return txt
}

func scanNum(txt string, nk NumberKind) (interface{}, string, error) {
	s, e, l := 0, 1, len(txt)
	if len(txt) > 1 {
		for ; e < l && intCheck[txt[e]] == 1; e++ {
		}
	}

	switch nk {
	case StringNumber:
		return Number(txt[s:e]), txt[e:], nil

	case IntNumber:
		num, err := strconv.Atoi(string(txt[s:e]))
		return num, txt[e:], err

	case FloatNumber:
		num, err := strconv.ParseFloat(string(txt[s:e]), 64)
		return num, txt[e:], err
	}
	return nil, "", ErrorExpectedNum
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

func scanString(txt []byte) ([]byte, []byte, error) {
	if len(txt) < 2 {
		return nil, nil, ErrorExpectedString
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
			return nil, nil, ErrorExpectedString
		}
		e += size
	}

	if txt[e] == '"' { // done we have nothing to unquote
		return txt[1:e], txt[e+1:], nil
	}

	out := make([]byte, len(txt)+2*utf8.UTFMax)
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
					return nil, nil, ErrorExpectedString
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
			return nil, nil, ErrorExpectedString

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
		return out[1:oute], txt[e:], nil
	}
	return nil, nil, ErrorExpectedString
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
