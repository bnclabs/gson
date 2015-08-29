package cbor

import "strconv"
import "unicode"
import "math"
import "encoding/binary"

var nullStr = "null"
var trueStr = "true"
var falseStr = "false"

func scanToken(txt string, out []byte, config *Config) (string, int) {
	txt = skipWS(txt, config.ws)

	if len(txt) < 1 {
		panic("cbor scanner empty json text")
	}

	if numCheck[txt[0]] == 1 {
		return scanNum(txt, config.nk, out)
	}

	switch txt[0] {
	case 'n':
		if len(txt) >= 4 && txt[:4] == nullStr {
			n := encodeNull(out)
			return txt[4:], n
		}
		panic("cbor scanner expected null")

	case 't':
		if len(txt) >= 4 && txt[:4] == trueStr {
			n := encodeTrue(out)
			return txt[4:], n
		}
		panic("cbor scanner expected true")

	case 'f':
		if len(txt) >= 5 && txt[:5] == falseStr {
			n := encodeFalse(out)
			return txt[5:], n
		}
		panic("cbor scanner expected false")

	case '"':
		return scanString(txt, out)

	case '[':
		if config.ct == LengthPrefix {
			panic("cbor scanner LengthPrefix not supported")
		}
		n, m := 0, 0
		n += encodeArrayStart(out[n:])
		if txt = skipWS(txt[1:], config.ws); len(txt) == 0 {
			panic("cbor scanner expected ']'")
		} else if txt[0] == ']' {
			n += encodeBreakStop(out[n:])
			return txt[1:], n
		}
		for {
			txt, m = scanToken(txt, out[n:], config)
			n += m
			if txt = skipWS(txt, config.ws); len(txt) == 0 {
				panic("cbor scanner expected ']'")
			} else if txt[0] == ',' {
				txt = skipWS(txt[1:], config.ws)
			} else if txt[0] == ']' {
				break
			} else {
				panic("cbor scanner expected ']'")
			}
		}
		n += encodeBreakStop(out[n:])
		return txt[1:], n

	case '{':
		if config.ct == LengthPrefix {
			panic("cbor scanner LengthPrefix not supported")
		}
		n, m := 0, 0
		n += encodeMapStart(out[n:])
		txt = skipWS(txt[1:], config.ws)
		if txt[0] == '}' {
			n += encodeBreakStop(out[n:])
			return txt[1:], n
		} else if txt[0] != '"' {
			panic("cbor scanner expected property key")
		}
		for {
			txt, m = scanString(txt, out[n:])
			n += m

			if txt = skipWS(txt, config.ws); len(txt) == 0 || txt[0] != ':' {
				panic("cbor scanner expected property colon")
			}
			txt, m = scanToken(skipWS(txt[1:], config.ws), out[n:], config)
			n += m

			if txt = skipWS(txt, config.ws); len(txt) == 0 {
				panic("cbor scanner expected '}'")
			} else if txt[0] == ',' {
				txt = skipWS(txt[1:], config.ws)
			} else if txt[0] == '}' {
				break
			} else {
				panic("cbor scanner expected '}'")
			}
		}
		n += encodeBreakStop(out[n:])
		return txt[1:], n

	default:
		panic("cbor scanner expected token")
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

func scanNum(txt string, nk NumberKind, out []byte) (string, int) {
	s, e, l, flt := 0, 1, len(txt), false
	if len(txt) > 1 {
		for ; e < l && intCheck[txt[e]] == 1; e++ {
			flt = flt || fltCheck[txt[e]] == 1 // detected as float
		}
	}
	switch nk {
	case JsonNumber:
		n := encodeTag(uint64(tagJsonNumber), out)
		n += encodeText(txt[s:e], out[n:])
		return txt[e:], n

	case FloatNumber:
		num, err := strconv.ParseFloat(string(txt[s:e]), 64)
		if err != nil { // once parsing logic is bullet proof remove this
			panic(err)
		}
		n := encodeFloat64(num, out)
		return txt[e:], n

	case IntNumber:
		if flt {
			panic("cbor scanner expected integer")
		}
		num, err := strconv.Atoi(txt[s:e])
		if err != nil { // once parsing logic is bullet proof remove this
			panic(err)
		}
		n := encodeInt64(int64(num), out)
		return txt[e:], n

	case FloatNumber32:
		num, err := strconv.ParseFloat(string(txt[s:e]), 32)
		if err != nil { // once parsing logic is bullet proof remove this
			panic(err)
		}
		n := encodeFloat32(float32(num), out)
		return txt[e:], n
	}
	// SmartNumber
	if flt && nk == SmartNumber32 {
		f, err := strconv.ParseFloat(string(txt[s:e]), 32)
		if err != nil { // once parsing logic is bullet proof remove this
			panic(err)
		}
		n := encodeFloat32(float32(f), out)
		return txt[e:], n
	} else if flt {
		f, err := strconv.ParseFloat(string(txt[s:e]), 64)
		if err != nil { // once parsing logic is bullet proof remove this
			panic(err)
		}
		n := encodeFloat64(f, out)
		return txt[e:], n
	}
	num, err := strconv.Atoi(txt[s:e])
	if err != nil { // once parsig logic is bullet proof remove this
		panic(err)
	}
	n := encodeInt64(int64(num), out)
	return txt[e:], n
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

//---- CBOR to JSON convertor

var nullBin = []byte("null")
var trueBin = []byte("true")
var falseBin = []byte("false")

func decodeTojson(in, out []byte) (int, int) {
	n, m := cborTojson[in[0]](in, out)
	return n, m
}

func decodeNullTojson(buf, out []byte) (int, int) {
	copy(out, nullBin)
	return 1, 4
}

func decodeTrueTojson(buf, out []byte) (int, int) {
	copy(out, trueBin)
	return 1, 4
}

func decodeFalseTojson(buf, out []byte) (int, int) {
	copy(out, falseBin)
	return 1, 5
}

func decodeFloat32Tojson(buf, out []byte) (int, int) {
	item, n := uint64(binary.BigEndian.Uint32(buf[1:])), 5
	// item, n := decodeType0Info26(buf) => to avoid memory allocation.
	f := math.Float32frombits(uint32(item))
	out = strconv.AppendFloat(out[:0], float64(f), 'f', 6, 32)
	return n, len(out)
}

func decodeFloat64Tojson(buf, out []byte) (int, int) {
	item, n := uint64(binary.BigEndian.Uint64(buf[1:])), 9
	// item, n := decodeType0Info27(buf) => to avoid memory allocation.
	f := math.Float64frombits(item)
	out = strconv.AppendFloat(out[:0], f, 'f', 20, 64)
	return n, len(out)
}

func decodeType0SmallIntTojson(buf, out []byte) (int, int) {
	val, n := uint64(info(buf[0])), 1
	out = strconv.AppendUint(out[:0], val, 10)
	return n, len(out)
}

func decodeType1SmallIntTojson(buf, out []byte) (int, int) {
	val, n := -int64(info(buf[0])+1), 1
	out = strconv.AppendInt(out[:0], val, 10)
	return n, len(out)
}

func decodeType0Info24Tojson(buf, out []byte) (int, int) {
	val, n := uint64(buf[1]), 2
	out = strconv.AppendUint(out[:0], val, 10)
	return n, len(out)
}

func decodeType1Info24Tojson(buf, out []byte) (int, int) {
	val, n := -int64(buf[1]+1), 2
	out = strconv.AppendInt(out[:0], val, 10)
	return n, len(out)
}

func decodeType0Info25Tojson(buf, out []byte) (int, int) {
	val, n := uint64(binary.BigEndian.Uint16(buf[1:])), 3
	out = strconv.AppendUint(out[:0], val, 10)
	return n, len(out)
}

func decodeType1Info25Tojson(buf, out []byte) (int, int) {
	val, n := -int64(binary.BigEndian.Uint16(buf[1:])+1), 3
	out = strconv.AppendInt(out[:0], val, 10)
	return n, len(out)
}

func decodeType0Info26Tojson(buf, out []byte) (int, int) {
	val, n := uint64(binary.BigEndian.Uint32(buf[1:])), 5
	out = strconv.AppendUint(out[:0], val, 10)
	return n, len(out)
}

func decodeType1Info26Tojson(buf, out []byte) (int, int) {
	val, n := -int64(binary.BigEndian.Uint32(buf[1:])+1), 5
	out = strconv.AppendInt(out[:0], val, 10)
	return n, len(out)
}

func decodeType0Info27Tojson(buf, out []byte) (int, int) {
	val, n := uint64(binary.BigEndian.Uint64(buf[1:])), 9
	out = strconv.AppendUint(out[:0], val, 10)
	return n, len(out)
}

func decodeType1Info27Tojson(buf, out []byte) (int, int) {
	x := uint64(binary.BigEndian.Uint64(buf[1:]))
	if x > 9223372036854775807 {
		panic("cbo->json number exceeds the limit of int64")
	}
	val, n := int64(-x)-1, 9
	out = strconv.AppendInt(out[:0], val, 10)
	return n, len(out)
}

// this is to support strings that are encoded via golang,
// but used by cbor->json decoder.
func decodeType3Tojson(buf, out []byte) (int, int) {
	ln, n := decodeLength(buf)
	out[0] = '"'
	copy(out[1:], buf[n:n+ln])
	out[ln+1] = '"'
	return n + ln, ln + 2
}

// this to support arrays thar are encoded via golang,
// but used by cbor->json decoder
func decodeType4Tojson(buf, out []byte) (int, int) {
	ln, n := decodeLength(buf)
	out[0] = '['
	if ln == 0 {
		out[1] = ']'
		return n, 2
	}
	m := 1
	for ; ln > 0; ln-- {
		x, y := cborTojson[buf[n]](buf[n:], out[m:])
		m, n = m+y, n+x
		out[m], m = ',', m+1
	}
	out[m-1] = ']'
	return n, m
}

func decodeType4IndefiniteTojson(buf, out []byte) (int, int) {
	out[0] = '['
	if buf[1] == brkstp {
		out[1] = ']'
		return 2, 2
	}
	n, m := 1, 1
	for buf[n] != brkstp {
		x, y := cborTojson[buf[n]](buf[n:], out[m:])
		m, n = m+y, n+x
		out[m], m = ',', m+1
	}
	out[m-1] = ']'
	return n + 1, m
}

// this to support maps thar are encoded via golang,
// but used by cbor->json decoder
func decodeType5Tojson(buf, out []byte) (int, int) {
	ln, n := decodeLength(buf)
	out[0] = '{'
	if ln == 0 {
		out[1] = '}'
		return n, 2
	}
	m := 1
	for ; ln > 0; ln-- {
		x, y := cborTojson[buf[n]](buf[n:], out[m:])
		m, n = m+y, n+x
		out[m], m = ':', m+1

		x, y = cborTojson[buf[n]](buf[n:], out[m:])
		m, n = m+y, n+x
		out[m], m = ',', m+1
	}
	out[m-1] = '}'
	return n, m
}

func decodeType5IndefiniteTojson(buf, out []byte) (int, int) {
	out[0] = '{'
	if buf[1] == brkstp {
		out[1] = '}'
		return 2, 2
	}
	n, m := 1, 1
	for buf[n] != brkstp {
		x, y := cborTojson[buf[n]](buf[n:], out[m:])
		m, n = m+y, n+x
		out[m], m = ':', m+1

		x, y = cborTojson[buf[n]](buf[n:], out[m:])
		m, n = m+y, n+x
		out[m], m = ',', m+1
	}
	out[m-1] = '}'
	return n + 1, m
}

func decodeTagTojson(buf, out []byte) (int, int) {
	byt := (buf[0] & 0x1f) | type0 // fix as positive num
	item, n := cborDecoders[byt](buf)
	switch item.(uint64) {
	case tagJsonString:
		ln, m := decodeLength(buf[n:])
		n += m
		out[0] = '"'
		copy(out[1:], buf[n:n+ln])
		out[ln+1] = '"'
		return n + ln, ln + 2
	case tagJsonNumber:
		ln, m := decodeLength(buf[n:])
		n += m
		copy(out, buf[n:n+ln])
		return n + ln, ln
	}
	return n, 0 // skip this tag
}

// ---- decoders

var cborTojson = make(map[byte]func([]byte, []byte) (int, int))

func init() {
	makePanic := func(msg string) func([]byte, []byte) (int, int) {
		return func(_, _ []byte) (int, int) { panic(msg) }
	}
	//-- type0                  (unsigned integer)
	// 1st-byte 0..23
	for i := byte(0); i < info24; i++ {
		cborTojson[hdr(type0, i)] = decodeType0SmallIntTojson
	}
	// 1st-byte 24..27
	cborTojson[hdr(type0, info24)] = decodeType0Info24Tojson
	cborTojson[hdr(type0, info25)] = decodeType0Info25Tojson
	cborTojson[hdr(type0, info26)] = decodeType0Info26Tojson
	cborTojson[hdr(type0, info27)] = decodeType0Info27Tojson
	// 1st-byte 28..31
	msg := "cbor->json decode type0 reserved info"
	cborTojson[hdr(type0, 28)] = makePanic(msg)
	cborTojson[hdr(type0, 29)] = makePanic(msg)
	cborTojson[hdr(type0, 30)] = makePanic(msg)
	cborTojson[hdr(type0, indefiniteLength)] = makePanic(msg)

	//-- type1                  (signed integer)
	// 1st-byte 0..23
	for i := byte(0); i < info24; i++ {
		cborTojson[hdr(type1, i)] = decodeType1SmallIntTojson
	}
	// 1st-byte 24..27
	cborTojson[hdr(type1, info24)] = decodeType1Info24Tojson
	cborTojson[hdr(type1, info25)] = decodeType1Info25Tojson
	cborTojson[hdr(type1, info26)] = decodeType1Info26Tojson
	cborTojson[hdr(type1, info27)] = decodeType1Info27Tojson
	// 1st-byte 28..31
	msg = "cbor->json type1 decode reserved info"
	cborTojson[hdr(type1, 28)] = makePanic(msg)
	cborTojson[hdr(type1, 29)] = makePanic(msg)
	cborTojson[hdr(type1, 30)] = makePanic(msg)
	cborTojson[hdr(type1, indefiniteLength)] = makePanic(msg)

	//-- type2                  (byte string)
	// 1st-byte 0..27
	msg = "cbor->json byte string not supported"
	for i := 0; i < 28; i++ {
		cborTojson[hdr(type2, byte(i))] = makePanic(msg)
	}
	// 1st-byte 28..31
	cborTojson[hdr(type2, 28)] = makePanic(msg)
	cborTojson[hdr(type2, 29)] = makePanic(msg)
	cborTojson[hdr(type2, 30)] = makePanic(msg)
	cborTojson[hdr(type2, indefiniteLength)] = makePanic(msg)

	//-- type3                  (string)
	// 1st-byte 0..27
	for i := 0; i < 28; i++ {
		cborTojson[hdr(type3, byte(i))] = decodeType3Tojson
	}
	// 1st-byte 28..31
	cborTojson[hdr(type3, 28)] = decodeType3Tojson
	cborTojson[hdr(type3, 29)] = decodeType3Tojson
	cborTojson[hdr(type3, 30)] = decodeType3Tojson
	msg = "cbor->json indefinite string not supported"
	cborTojson[hdr(type3, indefiniteLength)] = makePanic(msg)

	//-- type4                  (array)
	// 1st-byte 0..27
	for i := 0; i < 28; i++ {
		cborTojson[hdr(type4, byte(i))] = decodeType4Tojson
	}
	// 1st-byte 28..31
	cborTojson[hdr(type4, 28)] = decodeType4Tojson
	cborTojson[hdr(type4, 29)] = decodeType4Tojson
	cborTojson[hdr(type4, 30)] = decodeType4Tojson
	cborTojson[hdr(type4, indefiniteLength)] = decodeType4IndefiniteTojson

	//-- type5                  (map)
	// 1st-byte 0..27
	for i := 0; i < 28; i++ {
		cborTojson[hdr(type5, byte(i))] = decodeType5Tojson
	}
	// 1st-byte 28..31
	cborTojson[hdr(type5, 28)] = decodeType5Tojson
	cborTojson[hdr(type5, 29)] = decodeType5Tojson
	cborTojson[hdr(type5, 30)] = decodeType5Tojson
	cborTojson[hdr(type5, indefiniteLength)] = decodeType5IndefiniteTojson

	//-- type6
	// 1st-byte 0..23
	for i := byte(0); i < info24; i++ {
		cborTojson[hdr(type6, i)] = decodeTagTojson
	}
	// 1st-byte 24..27
	cborTojson[hdr(type6, info24)] = decodeTagTojson
	cborTojson[hdr(type6, info25)] = decodeTagTojson
	cborTojson[hdr(type6, info26)] = decodeTagTojson
	cborTojson[hdr(type6, info27)] = decodeTagTojson
	// 1st-byte 28..31
	msg = "cbor->json type6 decode reserved info"
	cborTojson[hdr(type6, 28)] = makePanic(msg)
	cborTojson[hdr(type6, 29)] = makePanic(msg)
	cborTojson[hdr(type6, 30)] = makePanic(msg)
	msg = "cbor->json indefinite type6 not supported"
	cborTojson[hdr(type6, indefiniteLength)] = makePanic(msg)

	//-- type7                  (simple values / floats / break-stop)
	msg = "cbor->json simple-type < 20 not supported"
	// 1st-byte 0..19
	for i := byte(0); i < 20; i++ {
		cborTojson[hdr(type7, i)] = makePanic(msg)
	}
	// 1st-byte 20..23
	cborTojson[hdr(type7, simpleTypeFalse)] = decodeFalseTojson
	cborTojson[hdr(type7, simpleTypeTrue)] = decodeTrueTojson
	cborTojson[hdr(type7, simpleTypeNil)] = decodeNullTojson
	msg = "cbor->json simple-type-undefined not supported"
	cborTojson[hdr(type7, simpleUndefined)] = makePanic(msg)

	msg = "cbor->json simple-type > 31 not supported"
	cborTojson[hdr(type7, simpleTypeByte)] = makePanic(msg)
	msg = "cbor->json float16 not supported"
	cborTojson[hdr(type7, flt16)] = makePanic(msg)
	cborTojson[hdr(type7, flt32)] = decodeFloat32Tojson
	cborTojson[hdr(type7, flt64)] = decodeFloat64Tojson
	// 1st-byte 28..31
	msg = "cbor->json simple-type 28 not supported"
	cborTojson[hdr(type7, 28)] = makePanic(msg)
	msg = "cbor->json simple-type 29 not supported"
	cborTojson[hdr(type7, 29)] = makePanic(msg)
	msg = "cbor->json simple-type 30 not supported"
	cborTojson[hdr(type7, 30)] = makePanic(msg)
	msg = "cbor->json simple-type break-code not supported"
	cborTojson[hdr(type7, itemBreak)] = makePanic(msg)
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
