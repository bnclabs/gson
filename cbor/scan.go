package cbor

import "strconv"
import "unicode"
import "unicode/utf8"
import "unicode/utf16"

var nullLiteral = "null"
var trueLiteral = "true"
var falseLiteral = "false"

func scanToken(txt string, out []byte, config *Config) (int, string, error) {
	txt = skipWS(txt, config.Ws)

	if len(txt) < 1 {
		return 0, txt, ErrorEmptyText
	}

	if digitCheck[txt[0]] == 1 {
		return scanNum(txt, config.Nk, out)
	}

	var err error

	switch txt[0] {
	case 'n':
		if len(txt) >= 4 && txt[:4] == nullLiteral {
			n := encodeNull(out)
			return n, txt[4:], nil
		}
		return 0, txt, ErrorExpectedNil

	case 't':
		if len(txt) >= 4 && txt[:4] == trueLiteral {
			n := encodeTrue(out)
			return n, txt[4:], nil
		}
		return 0, txt, ErrorExpectedTrue

	case 'f':
		if len(txt) >= 5 && txt[:5] == falseLiteral {
			n := encodeFalse(out)
			return n, txt[5:], nil
		}
		return 0, txt, ErrorExpectedFalse

	case '"':
		n, remtxt, err := scanString(str2bytes(txt), out)
		return n, bytes2str(remtxt), err

	case '[':
		if txt = skipWS(txt[1:], config.Ws); len(txt) == 0 {
			return 0, txt, ErrorExpectedClosearray
		} else if txt[0] == ']' {
			n := encodeArray([]interface{}{}, out)
			return n, txt[1:], nil
		}
		n, m := encodeArrayStart(out), 0
		for {
			m, txt, err = scanToken(txt, out[n:], config)
			if err != nil {
				return 0, txt, err
			}
			n += m
			if txt = skipWS(txt, config.Ws); len(txt) == 0 {
				return 0, txt, ErrorExpectedClosearray
			} else if txt[0] == ',' {
				txt = skipWS(txt[1:], config.Ws)
			} else if txt[0] == ']' {
				break
			} else {
				return 0, txt, ErrorExpectedClosearray
			}
		}
		n += encodeBreakStop(out[n:])
		return n, txt[1:], nil

	case '{':
		txt = skipWS(txt[1:], config.Ws)
		if txt[0] == '}' {
			n := encodeMap([][2]interface{}{}, out)
			return n, txt[1:], nil
		} else if txt[0] != '"' {
			return 0, txt, ErrorExpectedKey
		}
		n, m := encodeMapStart(out), 0
		for {
			m, remtxt, err := scanString(str2bytes(txt), out[n:])
			key := bytes2str(out[n : n+m])
			txt = bytes2str(remtxt)
			if err != nil {
				return 0, txt, err
			} else if len(key) < 1 {
				return 0, txt, ErrorExpectedKey
			}
			n += m

			if txt = skipWS(txt, config.Ws); len(txt) == 0 {
				return 0, txt, ErrorExpectedColon
			} else if txt[0] != ':' {
				return 0, txt, ErrorExpectedColon
			}
			m, txt, err = scanToken(skipWS(txt[1:], config.Ws), out[n:], config)
			if err != nil {
				return 0, txt, err
			}
			n += m

			if txt = skipWS(txt, config.Ws); len(txt) == 0 {
				return 0, txt, ErrorExpectedCloseobject
			} else if txt[0] == ',' {
				txt = skipWS(txt[1:], config.Ws)
			} else if txt[0] == '}' {
				break
			} else {
				return 0, txt, ErrorExpectedCloseobject
			}
		}
		return m, txt[1:], nil

	default:
		return 0, txt, ErrorExpectedToken
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

func scanNum(txt string, nk NumberKind, out []byte) (int, string, error) {
	s, e, l := 0, 1, len(txt)
	if len(txt) > 1 {
		for ; e < l && intCheck[txt[e]] == 1; e++ {
		}
	}
	if nk == IntNumber {
		num, err := strconv.Atoi(txt[s:e])
		if err != nil {
			n := encodeInt64(int64(num), out)
			return n, txt[e:], nil
		}
		return 0, txt, err
	}
	num, err := strconv.ParseFloat(string(txt[s:e]), 64)
	if err != nil {
		n := encodeFloat64(num, out)
		return n, txt[e:], nil
	}
	return 0, txt, err
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

func scanString(txt, out []byte) (int, []byte, error) {
	if len(txt) < 2 {
		return 0, txt, ErrorExpectedString
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
			return 0, txt, ErrorExpectedString
		}
		e += size
		if e == len(txt) {
			return 0, txt, ErrorExpectedString
		}
	}

	if txt[e] == '"' { // done we have nothing to unquote
		n := encodeText(bytes2str(txt[1:e]), out)
		return n, txt[e+1:], nil
	}

	outt := make([]byte, len(txt)+2*utf8.UTFMax)
	oute := copy(outt, txt[:e]) // copy so far

loop:
	for e < len(txt) {
		switch c := txt[e]; {
		case c == '"':
			outt[oute] = c
			e++
			break loop

		case c == '\\':
			if txt[e+1] == 'u' {
				r := getu4(txt[e:])
				if r < 0 { // invalid
					return 0, txt, ErrorExpectedString
				}
				e += 6
				if utf16.IsSurrogate(r) {
					nextr := getu4(txt[e:])
					dec := utf16.DecodeRune(r, nextr)
					if dec != unicode.ReplacementChar { // A valid pair consume
						oute += utf8.EncodeRune(outt[oute:], dec)
						e += 6
						break loop
					}
					// Invalid surrogate; fall back to replacement rune.
					r = unicode.ReplacementChar
				}
				oute += utf8.EncodeRune(outt[oute:], r)

			} else { // escaped with " \ / ' b f n r t
				outt[oute] = escapeCode[txt[e+1]]
				e += 2
				oute++
			}

		case c < ' ': // control character is invalid
			return 0, txt, ErrorExpectedString

		case c < utf8.RuneSelf: // ASCII
			outt[oute] = c
			oute++
			e++

		default: // coerce to well-formed UTF-8
			r, size := utf8.DecodeRune(txt[e:])
			e += size
			oute += utf8.EncodeRune(outt[oute:], r)
		}
	}

	if outt[oute] == '"' {
		n := encodeText(bytes2str(outt[1:oute]), out)
		return n, txt[e:], nil
	}
	return 0, txt, ErrorExpectedString
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
