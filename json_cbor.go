//  Copyright (c) 2015 Couchbase, Inc.

package gson

// transform json encoded value into cbor encoded value.
// cnf: SpaceKind, NumberKind, ContainerEncoding, strict

import "strconv"

var nullStr = "null"
var trueStr = "true"
var falseStr = "false"

func json2cbor(txt string, out []byte, config *Config) (string, int) {
	txt = skipWS(txt, config.ws)

	if len(txt) < 1 {
		panic("cbor scanner empty json text")
	}

	if numCheck[txt[0]] == 1 {
		return jsonNumToCbor(txt, out, config)
	}

	switch txt[0] {
	case 'n':
		if len(txt) >= 4 && txt[:4] == nullStr {
			n := cborNull(out)
			return txt[4:], n
		}
		panic("cbor scanner expected null")

	case 't':
		if len(txt) >= 4 && txt[:4] == trueStr {
			n := cborTrue(out)
			return txt[4:], n
		}
		panic("cbor scanner expected true")

	case 'f':
		if len(txt) >= 5 && txt[:5] == falseStr {
			n := cborFalse(out)
			return txt[5:], n
		}
		panic("cbor scanner expected false")

	case '"':
		n := 0
		txt, x := scanString(txt, out[n+16:]) // 16 reserved for cbor hdr
		n += valtext2cbor(bytes2str(out[n+16:n+16+x]), out[n:])
		return txt, n

	case '[':
		n, m, n_, n__ := 0, 0, 0, 0
		switch config.ct {
		case LengthPrefix:
			n_, n__ = n+32, n+32
		case Stream:
			n__ += arrayStart(out[n__:])
		}

		var ln int
		if txt = skipWS(txt[1:], config.ws); len(txt) == 0 {
			panic("cbor scanner expected ']'")
		} else if txt[0] != ']' {
			for {
				txt, m = json2cbor(txt, out[n__:], config)
				n__ += m
				ln++
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
		}
		switch config.ct {
		case LengthPrefix:
			x := valuint642cbor(uint64(ln), out[n:])
			out[n] = (out[n] & 0x1f) | cborType4 // fix type from type0->type4
			n += x
			n += copy(out[n:], out[n_:n__])
		case Stream:
			n__ += breakStop(out[n__:])
			n = n__
		}
		return txt[1:], n

	case '{':
		n, m, n_, n__ := 0, 0, 0, 0
		switch config.ct {
		case LengthPrefix:
			n_, n__ = n+32, n+32
		case Stream:
			n__ += mapStart(out[n__:])
		}

		var ln int
		txt = skipWS(txt[1:], config.ws)
		if txt[0] == '}' {
			// pass
		} else if txt[0] != '"' {
			panic("cbor scanner expected property key")
		} else {
			for {
				// 16 reserved for cbor hdr
				txt, m = scanString(txt, out[n__+16:])
				n__ += valtext2cbor(bytes2str(out[n__+16:n__+16+m]), out[n__:])

				if txt = skipWS(txt, config.ws); len(txt) == 0 || txt[0] != ':' {
					panic("cbor scanner expected property colon")
				}
				txt, m = json2cbor(skipWS(txt[1:], config.ws), out[n__:], config)
				n__ += m
				ln++

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
		}
		switch config.ct {
		case LengthPrefix:
			x := valuint642cbor(uint64(ln), out[n:])
			out[n] = (out[n] & 0x1f) | cborType5 // fix type from type0->type5
			n += x
			n += copy(out[n:], out[n_:n__])
		case Stream:
			n__ += breakStop(out[n__:])
			n = n__
		}
		return txt[1:], n

	default:
		panic("cbor scanner expected token")
	}
}

func jsonNumToCbor(txt string, out []byte, config *Config) (string, int) {
	s, e, l, flt := 0, 1, len(txt), false
	if len(txt) > 1 {
		for ; e < l && intCheck[txt[e]] == 1; e++ {
			flt = flt || fltCheck[txt[e]] == 1 // detected as float
		}
	}
	switch config.nk {
	case FloatNumber:
		num, err := strconv.ParseFloat(string(txt[s:e]), 64)
		if err != nil { // once parsing logic is bullet proof remove this
			panic(err)
		}
		n := valfloat642cbor(num, out)
		return txt[e:], n

	case IntNumber:
		if flt && config.strict == false { // try parsing it as float
			num, err := strconv.ParseFloat(string(txt[s:e]), 64)
			if err != nil { // once parsing logic is bullet proof remove this
				panic(err)
			}
			n := valfloat642cbor(num, out)
			return txt[e:], n

		} else if flt {
			panic("cbor scanner expected integer")
		}
		num, err := strconv.Atoi(txt[s:e])
		if err != nil { // once parsing logic is bullet proof remove this
			panic(err)
		}
		n := valint642cbor(int64(num), out)
		return txt[e:], n

	case FloatNumber32:
		f, err := strconv.ParseFloat(string(txt[s:e]), 64)
		if err != nil { // once parsing logic is bullet proof remove this
			panic(err)
		}
		n := valfloat322cbor(float32(f), out)
		return txt[e:], n
	}
	// SmartNumber
	if flt && config.nk == SmartNumber32 {
		f, err := strconv.ParseFloat(string(txt[s:e]), 64)
		if err != nil { // once parsing logic is bullet proof remove this
			panic(err)
		}
		n := valfloat322cbor(float32(f), out)
		return txt[e:], n
	} else if flt {
		f, err := strconv.ParseFloat(string(txt[s:e]), 64)
		if err != nil { // once parsing logic is bullet proof remove this
			panic(err)
		}
		n := valfloat642cbor(f, out)
		return txt[e:], n
	}
	num, err := strconv.Atoi(txt[s:e])
	if err != nil { // once parsig logic is bullet proof remove this
		panic(err)
	}
	n := valint642cbor(int64(num), out)
	return txt[e:], n
}
