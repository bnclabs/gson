package cbor

import "unicode"

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

func scanString(txt string, out []byte) (string, int) {
	if len(txt) < 2 {
		panic("cbor scanner expected string")
	}

	skipchar := false
	for off, ch := range txt[1:] {
		if skipchar {
			skipchar = false
			continue
		} else if ch == '\\' {
			skipchar = true
		} else if ch == '"' {
			end := off + 2
			n := encodeTag(uint64(tagJsonString), out)
			n += encodeText(txt[1:end-1], out[n:])
			return txt[end:], n
		}
	}
	panic("cbor scanner expected string")
}

var intCheck = [256]byte{}
var numCheck = [256]byte{}
var fltCheck = [256]byte{}

func init() {
	for i := 48; i <= 57; i++ {
		intCheck[i] = 1
		numCheck[i] = 1
	}
	intCheck['-'] = 1
	intCheck['+'] = 1
	intCheck['.'] = 1
	intCheck['e'] = 1
	intCheck['E'] = 1

	numCheck['-'] = 1
	numCheck['+'] = 1
	numCheck['.'] = 1

	fltCheck['.'] = 1
	fltCheck['e'] = 1
	fltCheck['E'] = 1
}
