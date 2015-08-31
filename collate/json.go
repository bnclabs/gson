package collate

import "strconv"
import "sort"
import "unicode"
import "unicode/utf8"
import "unicode/utf16"

func scanToken(txt string, code []byte, config *Config) (string, int) {
	txt = skipWS(txt, config.ws)
	if len(txt) < 1 {
		panic("collate scanner jsonEmpty")
	}

	n := 0

	if digitCheck[txt[0]] == 1 {
		code[n] = TypeNumber
		n++
		m, remtxt := scanNum(txt, code[n:], config.nt)
		n += m
		code[n] = Terminator
		n++
		return remtxt, n
	}

	switch txt[0] {
	case 'n':
		if len(txt) >= 4 && txt[:4] == "null" {
			code[n], code[n+1] = TypeNull, Terminator
			return txt[4:], n + 2
		}
		panic("collate scanner expectedNil")

	case 't':
		if len(txt) >= 4 && txt[:4] == "true" {
			code[n], code[n+1] = TypeTrue, Terminator
			return txt[4:], n + 2
		}
		panic("collate scanner expectedTrue")

	case 'f':
		if len(txt) >= 5 && txt[:5] == "false" {
			code[n], code[n+1] = TypeFalse, Terminator
			return txt[5:], n + 2
		}
		panic("collate scanner expectedFalse")

	case '"':
		s, remtxt := scanString(str2bytes(txt))
		if config.doMissing && MissingLiteral.Equal(bytes2str(s)) {
			code[n], code[n+1] = TypeMissing, Terminator
			return bytes2str(remtxt), n + 2
		}
		code[n] = TypeString
		n++
		n += suffixEncodeString(s, code[n:])
		code[n] = Terminator
		n++
		return bytes2str(remtxt), n

	case '[':
		var x int

		code[n] = TypeArray
		n++
		n_, n__, ln := n, n, 0
		if config.arrayLenPrefix {
			n_, n__ = (n_ + 32), (n__ + 32) // prealloc space for Len encoding
		}

		if txt = skipWS(txt[1:], config.ws); len(txt) == 0 {
			panic("collate scanner expectedCloseArray")

		} else if txt[0] != ']' {
			for {
				txt, x = scanToken(txt, code[n__:], config)
				n__ += x
				ln++
				if txt = skipWS(txt, config.ws); len(txt) == 0 {
					panic("gson scanner expectedCloseArray")
				} else if txt[0] == ',' {
					txt = skipWS(txt[1:], config.ws)
				} else if txt[0] == ']' {
					break
				} else {
					panic("collate scanner expectedCloseArray")
				}
			}
		}
		if config.arrayLenPrefix {
			n += gson2collate(Length(ln), code[n:], config)
			copy(code[n:], code[n_:n__])
			n += (n__ - n_)
		} else {
			n = n__
		}
		code[n] = Terminator
		n++
		return txt[1:], n

	case '{':
		var x int
		code[n] = TypeObj
		n++

		altcode, p := make([]byte, 10*1024), 0
		refs, ln := make(kvrefs, 10*256), 0

		if txt = skipWS(txt[1:], config.ws); len(txt) == 0 {
			panic("collate scanner expectedCloseobject")
		} else if txt[0] != '}' && txt[0] != '"' {
			panic("collate scanner expectedKey")
		} else if txt[0] != '}' {
			for {
				// NOTE: empty string is also a valid key
				key, remtxt := scanString(str2bytes(txt))
				txt = bytes2str(remtxt)
				if txt = skipWS(txt, config.ws); len(txt) == 0 || txt[0] != ':' {
					panic("collate scanner expectedColon")
				}
				txt, x = scanToken(skipWS(txt[1:], config.ws), altcode[p:], config)
				refs[ln] = kvref{bytes2str(key), altcode[p : p+x]}
				p += x
				ln++

				if txt = skipWS(txt, config.ws); len(txt) == 0 {
					panic("collate scanner expectedCloseobject")
				} else if txt[0] == ',' {
					txt = skipWS(txt[1:], config.ws)
				} else if txt[0] == '}' {
					break
				} else {
					panic("collate scanner expectedCloseobject")
				}
			}
			sort.Sort(refs[:ln])
		}
		if config.propertyLenPrefix {
			n += gson2collate(Length(ln), code[n:], config)
		}
		for j := 0; j < ln; j++ {
			kv := refs[j]
			n += gson2collate(kv.key, code[n:], config) // encode key
			copy(code[n:], kv.code)
			n += len(kv.code)
		}
		code[n] = Terminator
		n++
		return txt[1:], n
	}
	panic("collate scanner expectedToken")
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

func scanNum(txt string, code []byte, nk NumberKind) (int, string) {
	s, e, l := 0, 1, len(txt)
	if len(txt) > 1 {
		for ; e < l && intCheck[txt[e]] == 1; e++ {
		}
	}
	f, err := strconv.ParseFloat(txt[s:e], 64)
	if err != nil {
		panic(err)
	}
	n := normalizeFloat(f, code, nk)
	return n, txt[e:]
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

func scanString(txt []byte) ([]byte, []byte) {
	if len(txt) < 2 {
		panic("collate scanner expectedString")
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
			panic("collate scanner expectedString")
		}
	}

	if txt[e] == '"' { // done we have nothing to unquote
		return txt[1:e], txt[e+1:]
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
					panic("collate scanner expectedString")
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
			panic("collate scanner expectedString")

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
		return out[1:oute], txt[e:]
	}
	panic("collate scanner expectedString")
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

func collate2json(code []byte, text []byte, config *Config) (int, int) {
	if len(code) == 0 {
		return 0, 0
	}
	var scratch [64]byte
	n, m := 1, 0
	switch code[0] {
	case TypeMissing:
		copy(text, MissingLiteral)
		return n + 1, m + len(MissingLiteral)

	case TypeNull:
		copy(text, "null")
		return n + 1, m + 4

	case TypeTrue:
		copy(text, "true")
		return n + 1, m + 4

	case TypeFalse:
		copy(text, "false")
		return n + 1, m + 5

	case TypeNumber:
		x := getDatum(code[n:])
		y := denormalizeFloatTojson(code[n:n+x-1], text, config.nt)
		return n + x, m + y

	case TypeString:
		text[m] = '"'
		m++
		x, y := suffixDecodeString(code[n:], text[m:])
		m += y
		text[m] = '"'
		m++
		return n + x, m

	case TypeArray:
		if config.arrayLenPrefix {
			x := getDatum(code[n:])
			decodeInt(code[n:n+x-1], scratch[:])
			n += x
		}
		text[m] = '['
		m++
		for code[n] != Terminator {
			x, y := collate2json(code[n:], text[m:], config)
			n += x
			m += y
			text[m] = ','
			m++
		}
		n++ // skip terminator
		if text[m-1] == ',' {
			text[m-1] = ']'
		} else {
			text[m] = ']'
			m++
		}
		return n, m

	case TypeObj:
		if config.propertyLenPrefix {
			x := getDatum(code[n:])
			decodeInt(code[n:n+x-1], scratch[:])
			n += x
		}
		text[m] = '{'
		m++
		for code[n] != Terminator {
			x, y := collate2json(code[n:], text[m:], config)
			n, m = n+x, m+y
			text[m] = ':'
			m++
			x, y = collate2json(code[n:], text[m:], config)
			n, m = n+x, m+y
			text[m] = ','
			m++
		}
		n++ // skip terminator
		if text[m-1] == ',' {
			text[m-1] = '}'
		} else {
			text[m] = '}'
			m++
		}
		return n, m
	}
	panic("collate decode to json invalid binary")
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
