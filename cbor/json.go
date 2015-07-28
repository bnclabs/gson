package cbor

import "strconv"
import "unicode"
import "math"
import "encoding/json"
import "encoding/binary"

//---- JSON to CBOR
// a. directly uses cbor.codec functions to encode and decode stuff
// b. number can be encoded as integer or float.
// c. string is parsed as per encoding/json reference.
// d. arrays and maps are encoded using indefinite encoding.
// e. byte-string encoding is not used.

var nullStr = "null"
var trueStr = "true"
var falseStr = "false"

func scanToken(txt string, out []byte, config *Config) (string, int) {
	txt = skipWS(txt, config.Ws)

	if len(txt) < 1 {
		panic(ErrorEmptyText)
	}

	if numCheck[txt[0]] == 1 {
		return scanNum(txt, config.Nk, out)
	}

	switch txt[0] {
	case 'n':
		if len(txt) >= 4 && txt[:4] == nullStr {
			n := encodeNull(out)
			return txt[4:], n
		}
		panic(ErrorExpectedNil)

	case 't':
		if len(txt) >= 4 && txt[:4] == trueStr {
			n := encodeTrue(out)
			return txt[4:], n
		}
		panic(ErrorExpectedTrue)

	case 'f':
		if len(txt) >= 5 && txt[:5] == falseStr {
			n := encodeFalse(out)
			return txt[5:], n
		}
		panic(ErrorExpectedFalse)

	case '"':
		return scanString(txt, out)

	case '[':
		n, m := encodeArrayStart(out), 0
		if txt = skipWS(txt[1:], config.Ws); len(txt) == 0 {
			panic(ErrorExpectedClosearray)
		} else if txt[0] == ']' {
			n += encodeBreakStop(out[n:])
			return txt[1:], n
		}
		for {
			txt, m = scanToken(txt, out[n:], config)
			n += m
			if txt = skipWS(txt, config.Ws); len(txt) == 0 {
				panic(ErrorExpectedClosearray)
			} else if txt[0] == ',' {
				txt = skipWS(txt[1:], config.Ws)
			} else if txt[0] == ']' {
				break
			} else {
				panic(ErrorExpectedClosearray)
			}
		}
		n += encodeBreakStop(out[n:])
		return txt[1:], n

	case '{':
		n := encodeMapStart(out)
		txt = skipWS(txt[1:], config.Ws)
		if txt[0] == '}' {
			n += encodeBreakStop(out[n:])
			return txt[1:], n
		} else if txt[0] != '"' {
			panic(ErrorExpectedKey)
		}
		var m int
		for {
			txt, m = scanString(txt, out[n:])
			n += m

			if txt = skipWS(txt, config.Ws); len(txt) == 0 || txt[0] != ':' {
				panic(ErrorExpectedColon)
			}
			txt, m = scanToken(skipWS(txt[1:], config.Ws), out[n:], config)
			n += m

			if txt = skipWS(txt, config.Ws); len(txt) == 0 {
				panic(ErrorExpectedCloseobject)
			} else if txt[0] == ',' {
				txt = skipWS(txt[1:], config.Ws)
			} else if txt[0] == '}' {
				break
			} else {
				panic(ErrorExpectedCloseobject)
			}
		}
		n += encodeBreakStop(out[n:])
		return txt[1:], n

	default:
		panic(ErrorExpectedToken)
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
	if nk == IntNumber && flt {
		panic(ErrorExpectedInteger)

	} else if nk == IntNumber {
		num, err := strconv.Atoi(txt[s:e])
		if err == nil {
			n := encodeInt64(int64(num), out)
			return txt[e:], n
		}

	} else if nk == FloatNumber {
		num, err := strconv.ParseFloat(string(txt[s:e]), 64)
		if err == nil {
			n := encodeFloat64(num, out)
			return txt[e:], n
		}
	}
	// SmartNumber
	if flt {
		f, _ := strconv.ParseFloat(string(txt[s:e]), 64)
		n := encodeFloat64(f, out)
		return txt[e:], n
	}
	num, _ := strconv.Atoi(txt[s:e])
	n := encodeInt64(int64(num), out)
	return txt[e:], n
}

func scanString(txt string, out []byte) (string, int) {
	if len(txt) < 2 {
		panic(ErrorExpectedString)
	}

	skipchar := false
	for off, ch := range txt[1:] {
		if skipchar {
			skipchar = false
			continue
		} else if ch == '\\' {
			skipchar = true
		} else if ch == '"' {
			str := bytes2str(out[10:])
			end := off + 2
			// use encoding/json for unmarshaling string.
			if err := json.Unmarshal(str2bytes(txt[:end]), &str); err != nil {
				panic(err)
			}
			return txt[end:], encodeText(str, out)
		}
	}
	panic(ErrorExpectedString)
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

//---- CBOR to JSON convertor

var nullBin = []byte("null")
var trueBin = []byte("true")
var falseBin = []byte("false")

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
	item, n := decodeType0Info26(buf)
	f := math.Float32frombits(uint32(item.(uint64)))
	out = strconv.AppendFloat(out[:0], float64(f), 'f', 6, 32)
	return n, len(out)
}

func decodeFloat64Tojson(buf, out []byte) (int, int) {
	item, n := decodeType0Info27(buf)
	f := math.Float64frombits(item.(uint64))
	out = strconv.AppendFloat(out[:0], f, 'f', 6, 64)
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
	//if x > 9223372036854775807 {
	//    panic("number exceeds the limit of int64")
	//}
	val, n := int64(-x)-1, 9
	out = strconv.AppendInt(out[:0], val, 10)
	return n, len(out)
}

func decodeType3Tojson(buf, out []byte) (int, int) {
	ln, n := decodeLength(buf)
	data, _ := json.Marshal(bytes2str(buf[n : n+ln]))
	copy(out, data)
	return n + ln, len(data)
}

func decodeType4IndefiniteTojson(buf, out []byte) (int, int) {
	brkstp := hdr(type7, itemBreak)
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

func decodeType5IndefiniteTojson(buf, out []byte) (int, int) {
	brkstp := hdr(type7, itemBreak)
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

// ---- decoders

var cborTojson = make(map[byte]func([]byte, []byte) (int, int))

func init() {
	makePanic := func(msg error) func([]byte, []byte) (int, int) {
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
	cborTojson[hdr(type0, 28)] = makePanic(ErrorInfoReserved)
	cborTojson[hdr(type0, 29)] = makePanic(ErrorInfoReserved)
	cborTojson[hdr(type0, 30)] = makePanic(ErrorInfoReserved)
	cborTojson[hdr(type0, indefiniteLength)] = makePanic(ErrorInfoIndefinite)

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
	cborTojson[hdr(type1, 28)] = makePanic(ErrorInfoReserved)
	cborTojson[hdr(type1, 29)] = makePanic(ErrorInfoReserved)
	cborTojson[hdr(type1, 30)] = makePanic(ErrorInfoReserved)
	cborTojson[hdr(type1, indefiniteLength)] = makePanic(ErrorInfoIndefinite)

	//-- type2                  (byte string)
	// 1st-byte 0..27
	for i := 0; i < 28; i++ {
		cborTojson[hdr(type2, byte(i))] = makePanic(ErrorByteString)
	}
	// 1st-byte 28..31
	cborTojson[hdr(type2, 28)] = makePanic(ErrorByteString)
	cborTojson[hdr(type2, 29)] = makePanic(ErrorByteString)
	cborTojson[hdr(type2, 30)] = makePanic(ErrorByteString)
	cborTojson[hdr(type2, indefiniteLength)] = makePanic(ErrorByteString)

	//-- type3                  (string)
	// 1st-byte 0..27
	for i := 0; i < 28; i++ {
		cborTojson[hdr(type3, byte(i))] = decodeType3Tojson
	}
	// 1st-byte 28..31
	cborTojson[hdr(type3, 28)] = makePanic(ErrorInfoReserved)
	cborTojson[hdr(type3, 29)] = makePanic(ErrorInfoReserved)
	cborTojson[hdr(type3, 30)] = makePanic(ErrorInfoReserved)
	cborTojson[hdr(type3, indefiniteLength)] = makePanic(ErrorInfoIndefinite)

	//-- type4                  (array)
	// 1st-byte 0..27
	for i := 0; i < 28; i++ {
		cborTojson[hdr(type4, byte(i))] = makePanic(ErrorExpectedIndefinite)
	}
	// 1st-byte 28..31
	cborTojson[hdr(type4, 28)] = makePanic(ErrorInfoReserved)
	cborTojson[hdr(type4, 29)] = makePanic(ErrorInfoReserved)
	cborTojson[hdr(type4, 30)] = makePanic(ErrorInfoReserved)
	cborTojson[hdr(type4, indefiniteLength)] = decodeType4IndefiniteTojson

	//-- type5                  (map)
	// 1st-byte 0..27
	for i := 0; i < 28; i++ {
		cborTojson[hdr(type5, byte(i))] = makePanic(ErrorExpectedIndefinite)
	}
	// 1st-byte 28..31
	cborTojson[hdr(type5, 28)] = makePanic(ErrorInfoReserved)
	cborTojson[hdr(type5, 29)] = makePanic(ErrorInfoReserved)
	cborTojson[hdr(type5, 30)] = makePanic(ErrorInfoReserved)
	cborTojson[hdr(type5, indefiniteLength)] = decodeType5IndefiniteTojson

	//-- type6
	// 1st-byte 0..23
	for i := byte(0); i < info24; i++ {
		cborTojson[hdr(type6, i)] = makePanic(ErrorTagNotSupported)
	}
	// 1st-byte 24..27
	cborTojson[hdr(type6, info24)] = makePanic(ErrorTagNotSupported)
	cborTojson[hdr(type6, info25)] = makePanic(ErrorTagNotSupported)
	cborTojson[hdr(type6, info26)] = makePanic(ErrorTagNotSupported)
	cborTojson[hdr(type6, info27)] = makePanic(ErrorTagNotSupported)
	// 1st-byte 28..31
	cborTojson[hdr(type6, 28)] = makePanic(ErrorInfoReserved)
	cborTojson[hdr(type6, 29)] = makePanic(ErrorInfoReserved)
	cborTojson[hdr(type6, 30)] = makePanic(ErrorInfoReserved)
	cborTojson[hdr(type6, indefiniteLength)] = makePanic(ErrorInfoIndefinite)

	//-- type7                  (simple values / floats / break-stop)
	// 1st-byte 0..19
	for i := byte(0); i < 20; i++ {
		cborTojson[hdr(type7, i)] = makePanic(ErrorUnassigned)
	}
	// 1st-byte 20..23
	cborTojson[hdr(type7, simpleTypeFalse)] = decodeFalseTojson
	cborTojson[hdr(type7, simpleTypeTrue)] = decodeTrueTojson
	cborTojson[hdr(type7, simpleTypeNil)] = decodeNullTojson
	cborTojson[hdr(type7, simpleUndefined)] = makePanic(ErrorUndefined)

	cborTojson[hdr(type7, simpleTypeByte)] = makePanic(ErrorSimpleType)
	cborTojson[hdr(type7, flt16)] = makePanic(ErrorFloat16)
	cborTojson[hdr(type7, flt32)] = decodeFloat32Tojson
	cborTojson[hdr(type7, flt64)] = decodeFloat64Tojson
	// 1st-byte 28..31
	cborTojson[hdr(type7, 28)] = makePanic(ErrorUnassigned)
	cborTojson[hdr(type7, 29)] = makePanic(ErrorUnassigned)
	cborTojson[hdr(type7, 30)] = makePanic(ErrorUnassigned)
	cborTojson[hdr(type7, itemBreak)] = makePanic(ErrorBreakcode)
}
