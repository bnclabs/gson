//  Copyright (c) 2015 Couchbase, Inc.

package gson

import "strconv"
import "sort"

func json2collate(txt string, code []byte, config *Config) (string, int) {
	txt = skipWS(txt, config.ws)
	if len(txt) < 1 {
		panic("collate scanner jsonEmpty")
	}

	n := 0

	if digitCheck[txt[0]] == 1 {
		code[n] = TypeNumber
		n++
		m, remtxt := jsonnum2collate(txt, code[n:], config.nk)
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
		scratch := config.pools.stringPool.Get().([]byte)
		defer config.pools.stringPool.Put(scratch)

		txt, p := scanString(txt, scratch)
		if config.doMissing && MissingLiteral.Equal(bytes2str(scratch[:p])) {
			code[n], code[n+1] = TypeMissing, Terminator
			return txt, n + 2
		}
		code[n] = TypeString
		n++
		n += suffixEncodeString(scratch[:p], code[n:])
		code[n] = Terminator
		n++
		return txt, n

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
				txt, x = json2collate(txt, code[n__:], config)
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
			n += collateLength(ln, code[n:])
			n += copy(code[n:], code[n_:n__])
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

		altcode, p := config.pools.codepool.Get().([]byte), 0
		defer config.pools.codepool.Put(altcode)
		refs, ln := config.pools.keypool.Get().(kvrefs), 0
		defer config.pools.keypool.Put(refs)

		if txt = skipWS(txt[1:], config.ws); len(txt) == 0 {
			panic("collate scanner expectedCloseobject")
		} else if txt[0] != '}' && txt[0] != '"' {
			panic("collate scanner expectedKey")
		} else if txt[0] != '}' {
			for {
				// NOTE: empty string is also a valid key
				txt, x = scanString(txt, altcode[p:])
				if txt = skipWS(txt, config.ws); len(txt) == 0 || txt[0] != ':' {
					panic("collate scanner expectedColon")
				}
				key := bytes2str(altcode[p : p+x])
				p += x

				txt = skipWS(txt[1:], config.ws)
				txt, x = json2collate(txt, altcode[p:], config)
				refs[ln] = kvref{key, altcode[p : p+x]}
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
			n += collateLength(ln, code[n:])
		}
		for j := 0; j < ln; j++ {
			kv := refs[j]
			n += gson2collate(kv.key, code[n:], config) // encode key
			n += copy(code[n:], kv.code)
		}
		code[n] = Terminator
		n++
		return txt[1:], n
	}
	panic("collate scanner expectedToken")
}

func jsonnum2collate(txt string, code []byte, nk NumberKind) (int, string) {
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
		y := denormalizeFloatTojson(code[n:n+x-1], text, config.nk)
		return n + x, m + y

	case TypeString:
		block := config.pools.stringPool.Get()
		scratch := block.([]byte)
		defer config.pools.stringPool.Put(block)
		x, y := suffixDecodeString(code[n:], scratch[:])
		config.buf.Reset()
		if err := config.enc.Encode(bytes2str(scratch[:y])); err != nil {
			panic(err)
		}
		s := config.buf.Bytes()
		m += copy(text[m:], s[:len(s)-1]) // -1 to strip \n
		return n + x, m

	case TypeArray:
		if config.arrayLenPrefix {
			x := getDatum(code[n:])
			collated2Int(code[n:n+x-1], scratch[:])
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
			collated2Int(code[n:n+x-1], scratch[:])
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
