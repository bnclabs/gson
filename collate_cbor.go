//  Copyright (c) 2015 Couchbase, Inc.

package gson

import "math"
import "strconv"
import "sort"
import "encoding/binary"

//---- collate to cbor

func collate2cbor(code, out []byte, config *Config) (int, int) {
	if len(code) == 0 {
		return 0, 0
	}
	var scratch [64]byte
	m, n := 1, 0
	switch code[0] {
	case TypeMissing:
		n += valtext2cbor(string(MissingLiteral), out[n:])
		return m + 1, n

	case TypeNull:
		n += cborNull(out[n:])
		return m + 1, n

	case TypeTrue:
		n += cborTrue(out[n:])
		return m + 1, n

	case TypeFalse:
		n += cborFalse(out[n:])
		return m + 1, n

	case TypeNumber:
		x := getDatum(code[m:])
		// -1 is to skip terminator
		num := denormalizeFloat(code[m:m+x-1], config.nk)
		switch v := num.(type) {
		case float32:
			n += valfloat322cbor(v, out[n:])
		case float64:
			n += valfloat642cbor(v, out[n:])
		case int64:
			n += valint642cbor(v, out[n:])
		}
		return m + x, n

	case TypeString:
		scratch := stringPool.Get().([]byte)
		defer stringPool.Put(scratch)
		x, y := suffixDecodeString(code[m:], scratch)
		n += valtext2cbor(bytes2str(scratch[:y]), out[n:])
		return m + x, n

	case TypeBinary:
		x := getDatum(code[m:])
		n += valbytes2cbor(code[m:m+x-1], out[n:])
		return m + x, n

	case TypeArray:
		if config.arrayLenPrefix {
			if code[m] != TypeLength {
				panic("collate decode expected array length prefix")
			}
			x := getDatum(code[m:])
			// -1 skip terminator
			collated2Int(code[m:m+x-1], scratch[:]) // skip length
			m += x
		}
		n_, n__ := n, n
		if config.ct == LengthPrefix {
			n_, n__ = n+32, n+32
		} else if config.ct == Stream {
			n__ += arrayStart(out[n__:])
		}
		ln := 0
		for code[m] != Terminator {
			x, y := collate2cbor(code[m:], out[n__:], config)
			m, n__ = m+x, n__+y
			ln++
		}
		if config.ct == LengthPrefix {
			x := valuint642cbor(uint64(ln), out[n:])
			out[n] = (out[n] & 0x1f) | cborType4 // fix type from type0->type4
			n += x
			n += copy(out[n:], out[n_:n__])
		} else if config.ct == Stream {
			n__ += breakStop(out[n__:])
			n = n__
		}
		return m + 1, n

	case TypeObj:
		if config.propertyLenPrefix {
			if code[m] != TypeLength {
				panic("collate decode expected property length prefix")
			}
			x := getDatum(code[m:])
			// -1 skip terminator
			collated2Int(code[m:m+x-1], scratch[:]) // skip length
			m += x
		}
		n_, n__ := n, n
		if config.ct == LengthPrefix {
			n_, n__ = n+32, n+32
		} else if config.ct == Stream {
			n__ += mapStart(out[n__:])
		}

		ln := 0
		for code[m] != Terminator {
			x, y := collate2cbor(code[m:], out[n__:], config)
			m, n__ = m+x, n__+y
			x, y = collate2cbor(code[m:], out[n__:], config)
			m, n__ = m+x, n__+y
			ln++
		}
		if config.ct == LengthPrefix {
			x := valuint642cbor(uint64(ln), out[n:])
			out[n] = (out[n] & 0x1f) | cborType5 // fix type from type0->type5
			n += x
			n += copy(out[n:], out[n_:n__])
		} else if config.ct == Stream {
			n__ += breakStop(out[n__:])
			n = n__
		}
		return m + 1, n
	}
	panic("collate decode to cbor invalid binary")
}

//---- CBOR to Collate

func cbor2collate(in, out []byte, config *Config) (int, int) {
	return cbor2collateM[in[0]](in, out, config)
}

func collateCborNull(buf, out []byte, config *Config) (int, int) {
	out[0], out[1] = TypeNull, Terminator
	return 1, 2
}

func collateCborTrue(buf, out []byte, config *Config) (int, int) {
	out[0], out[1] = TypeTrue, Terminator
	return 1, 2
}

func collateCborFalse(buf, out []byte, config *Config) (int, int) {
	out[0], out[1] = TypeFalse, Terminator
	return 1, 2
}

func collateCborFloat32(buf, out []byte, config *Config) (int, int) {
	item := uint64(binary.BigEndian.Uint32(buf[1:]))
	f, n := math.Float32frombits(uint32(item)), 0
	out[n] = TypeNumber
	n++
	n += normalizeFloat(float64(f), out[n:], config.nk)
	out[n] = Terminator
	n++
	return 5, n
}

func collateCborFloat64(buf, out []byte, config *Config) (int, int) {
	item := uint64(binary.BigEndian.Uint64(buf[1:]))
	f, n := math.Float64frombits(item), 0
	out[n] = TypeNumber
	n++
	n += normalizeFloat(float64(f), out[n:], config.nk)
	out[n] = Terminator
	n++
	return 9, n
}

func collateCborT0SmallInt(buf, out []byte, config *Config) (int, int) {
	n := 0
	out[n] = TypeNumber
	n++
	n += normalizeFloat(int64(cborInfo(buf[0])), out[n:], config.nk)
	out[n] = Terminator
	n++
	return 1, n
}

func collateCborT1SmallInt(buf, out []byte, config *Config) (int, int) {
	n := 0
	out[n] = TypeNumber
	n++
	n += normalizeFloat(-int64(cborInfo(buf[0])+1), out[n:], config.nk)
	out[n] = Terminator
	n++
	return 1, n
}

func collateCborT0Info24(buf, out []byte, config *Config) (int, int) {
	n := 0
	out[n] = TypeNumber
	n++
	n += normalizeFloat(int64(buf[1]), out[n:], config.nk)
	out[n] = Terminator
	n++
	return 2, n
}

func collateCborT1Info24(buf, out []byte, config *Config) (int, int) {
	n := 0
	out[n] = TypeNumber
	n++
	n += normalizeFloat(-int64(buf[1]+1), out[n:], config.nk)
	out[n] = Terminator
	n++
	return 2, n
}

func collateCborT0Info25(buf, out []byte, config *Config) (int, int) {
	n := 0
	out[n] = TypeNumber
	n++
	i := int64(binary.BigEndian.Uint16(buf[1:]))
	n += normalizeFloat(i, out[n:], config.nk)
	out[n] = Terminator
	n++
	return 3, n
}

func collateCborT1Info25(buf, out []byte, config *Config) (int, int) {
	n := 0
	out[n] = TypeNumber
	n++
	i := -int64(binary.BigEndian.Uint16(buf[1:]) + 1)
	n += normalizeFloat(i, out[n:], config.nk)
	out[n] = Terminator
	n++
	return 3, n
}

func collateCborT0Info26(buf, out []byte, config *Config) (int, int) {
	n := 0
	out[n] = TypeNumber
	n++
	i := int64(binary.BigEndian.Uint32(buf[1:]))
	n += normalizeFloat(i, out[n:], config.nk)
	out[n] = Terminator
	n++
	return 5, n
}

func collateCborT1Info26(buf, out []byte, config *Config) (int, int) {
	n := 0
	out[n] = TypeNumber
	n++
	i := -int64(binary.BigEndian.Uint32(buf[1:]) + 1)
	n += normalizeFloat(i, out[n:], config.nk)
	out[n] = Terminator
	n++
	return 5, n
}

func collateCborT0Info27(buf, out []byte, config *Config) (int, int) {
	n := 0
	out[n] = TypeNumber
	n++
	i := int64(binary.BigEndian.Uint64(buf[1:]))
	n += normalizeFloat(i, out[n:], config.nk)
	out[n] = Terminator
	n++
	return 9, n
}

func collateCborT1Info27(buf, out []byte, config *Config) (int, int) {
	x := uint64(binary.BigEndian.Uint64(buf[1:]))
	if x > 9223372036854775807 {
		panic("cbo->collate number exceeds the limit of int64")
	}
	val, n := (int64(-x) - 1), 0
	out[n] = TypeNumber
	n++
	n += normalizeFloat(val, out[n:], config.nk)
	out[n] = Terminator
	n++
	return 9, n
}

func collateCborT2(buf, out []byte, config *Config) (int, int) {
	ln, m := cborItemLength(buf)
	n := 0
	out[n] = TypeBinary
	n++
	copy(out[n:], buf[m:m+ln])
	n += ln
	out[n] = Terminator
	n++
	return m + ln, n
}

func collateCborT3(buf, out []byte, config *Config) (int, int) {
	ln, m := cborItemLength(buf)
	if config.doMissing && MissingLiteral.Equal(bytes2str(buf[m:m+ln])) {
		out[0], out[1] = TypeMissing, Terminator
		return m + ln, 2
	}
	n := 0
	out[n] = TypeString
	n++
	n += suffixEncodeString(buf[m:m+ln], out[n:])
	out[n] = Terminator
	n++
	return m + ln, n
}

func collateCborLength(length int, out []byte, config *Config) int {
	n := 0
	out[n] = TypeLength
	n++
	n += normalizeFloat(int64(length), out[n:], IntNumber)
	out[n] = Terminator
	n++
	return n
}

func collateCborT4(buf, out []byte, config *Config) (int, int) {
	ln, m := cborItemLength(buf)
	n := 0
	out[n] = TypeArray
	n++
	if config.arrayLenPrefix {
		n += collateCborLength(ln, out[n:], config)
	}
	for ; ln > 0; ln-- {
		x, y := cbor2collate(buf[m:], out[n:], config)
		m, n = m+x, n+y
	}
	out[n] = Terminator
	n++
	return m, n
}

func collateCborT4Indef(buf, out []byte, config *Config) (m int, n int) {
	ln := 0
	out[n] = TypeArray
	n++
	n_, n__ := n, n
	if config.arrayLenPrefix {
		n_, n__ = n+32, n+32 // length encoding can go upto max of 32 bytes
	}

	defer func() {
		if config.arrayLenPrefix {
			n += collateCborLength(ln, out[n:], config)
		}
		copy(out[n:], out[n_:n__])
		n += (n__ - n_)
		out[n] = Terminator
		n++
		return
	}()

	m = 1
	if buf[1] == brkstp {
		m = 2
		return
	}
	for buf[m] != brkstp {
		x, y := cbor2collate(buf[m:], out[n__:], config)
		m, n__ = m+x, n__+y
		ln++
	}
	m++
	return
}

func collateCborT5(buf, out []byte, config *Config) (int, int) {
	ln, m := cborItemLength(buf)
	n := 0
	out[n] = TypeObj
	n++
	if config.propertyLenPrefix {
		n += collateCborLength(ln, out[n:], config)
	}

	pool := getJsonKeyPool(config.maxKeys)
	altcode, p := pool.codepool.Get().([]byte), 0
	defer pool.codepool.Put(altcode)
	refs := pool.keypool.Get().(kvrefs)
	defer pool.keypool.Put(refs)

	for i := 0; i < ln; i++ {
		x, y := cbor2collate(buf[m:], altcode[p:], config) // key
		key := altcode[p : p+y]
		m, p = m+x, p+y
		x, y = cbor2collate(buf[m:], altcode[p:], config) // value
		refs[i] = kvref{bytes2str(key), altcode[p : p+y]}
		m, p = m+x, p+y
	}
	sort.Sort(refs[:ln])
	for i := 0; i < ln; i++ {
		kv := refs[i]
		copy(out[n:], str2bytes(kv.key))
		n += len(kv.key)
		copy(out[n:], kv.code)
		n += len(kv.code)
	}

	out[n] = Terminator
	n++
	return m, n
}

func collateCborT5Indef(buf, out []byte, config *Config) (m int, n int) {
	ln := 0
	out[n] = TypeObj
	n++

	pool := getJsonKeyPool(config.maxKeys)
	altcode, p := pool.codepool.Get().([]byte), 0
	defer pool.codepool.Put(altcode)
	refs := pool.keypool.Get().(kvrefs)
	defer pool.keypool.Put(refs)

	m = 1
	for buf[m] != brkstp {
		x, y := cbor2collate(buf[m:], altcode[p:], config) // key
		key := altcode[p : p+y]
		m, p = m+x, p+y
		x, y = cbor2collate(buf[m:], altcode[p:], config) // value
		refs[ln] = kvref{bytes2str(key), altcode[p : p+y]}
		m, p = m+x, p+y
		ln++
	}
	m++
	sort.Sort(refs[:ln])

	if config.propertyLenPrefix {
		n += collateCborLength(ln, out[n:], config)
	}
	for i := 0; i < ln; i++ {
		kv := refs[i]
		copy(out[n:], str2bytes(kv.key))
		n += len(kv.key)
		copy(out[n:], kv.code)
		n += len(kv.code)
	}
	out[n] = Terminator
	n++
	return
}

func collateCborTag(buf, out []byte, config *Config) (int, int) {
	item, m := cborItemLength(buf)
	switch uint64(item) {
	case tagJsonString:
		ln, x := cborItemLength(buf[m:])
		m += x
		// copy the JSON string into scratch buffer.
		scratch := stringPool.Get().([]byte)
		utf8str := stringPool.Get().([]byte)
		defer stringPool.Put(scratch)
		defer stringPool.Put(utf8str)
		scratch[0] = '"'
		copy(scratch[1:], buf[m:m+ln])
		scratch[1+ln] = '"'
		_, y := scanString(bytes2str(scratch[:ln+2]), utf8str)
		// collate golang string.
		n := 0
		out[n] = TypeString
		n++
		n += suffixEncodeString(utf8str[:y], out[n:])
		out[n] = Terminator
		n++
		return m + ln, n

	case tagJsonNumber:
		ln, x := cborItemLength(buf[m:])
		m += x
		f, err := strconv.ParseFloat(bytes2str(buf[m:m+ln]), 64)
		if err != nil {
			panic(err)
		}
		n := 0
		out[n] = TypeNumber
		n++
		n += normalizeFloat(f, out[n:], config.nk)
		out[n] = Terminator
		return m + ln, n + 1
	}
	return m, 0 // skip this tag
}

var cbor2collateM = make(map[byte]func([]byte, []byte, *Config) (int, int))

func init() {
	makePanic := func(msg string) func([]byte, []byte, *Config) (int, int) {
		return func(_, _ []byte, _ *Config) (int, int) { panic(msg) }
	}
	//-- type0                  (unsigned integer)
	// 1st-byte 0..23
	for i := byte(0); i < cborInfo24; i++ {
		cbor2collateM[cborHdr(cborType0, i)] = collateCborT0SmallInt
	}
	// 1st-byte 24..27
	cbor2collateM[cborHdr(cborType0, cborInfo24)] = collateCborT0Info24
	cbor2collateM[cborHdr(cborType0, cborInfo25)] = collateCborT0Info25
	cbor2collateM[cborHdr(cborType0, cborInfo26)] = collateCborT0Info26
	cbor2collateM[cborHdr(cborType0, cborInfo27)] = collateCborT0Info27
	// 1st-byte 28..31
	msg := "cbor->collate decode type0 reserved info"
	cbor2collateM[cborHdr(cborType0, 28)] = makePanic(msg)
	cbor2collateM[cborHdr(cborType0, 29)] = makePanic(msg)
	cbor2collateM[cborHdr(cborType0, 30)] = makePanic(msg)
	cbor2collateM[cborHdr(cborType0, cborIndefiniteLength)] = makePanic(msg)

	//-- type1                  (signed integer)
	// 1st-byte 0..23
	for i := byte(0); i < cborInfo24; i++ {
		cbor2collateM[cborHdr(cborType1, i)] = collateCborT1SmallInt
	}
	// 1st-byte 24..27
	cbor2collateM[cborHdr(cborType1, cborInfo24)] = collateCborT1Info24
	cbor2collateM[cborHdr(cborType1, cborInfo25)] = collateCborT1Info25
	cbor2collateM[cborHdr(cborType1, cborInfo26)] = collateCborT1Info26
	cbor2collateM[cborHdr(cborType1, cborInfo27)] = collateCborT1Info27
	// 1st-byte 28..31
	msg = "cbor->collate cborType1 decode reserved info"
	cbor2collateM[cborHdr(cborType1, 28)] = makePanic(msg)
	cbor2collateM[cborHdr(cborType1, 29)] = makePanic(msg)
	cbor2collateM[cborHdr(cborType1, 30)] = makePanic(msg)
	cbor2collateM[cborHdr(cborType1, cborIndefiniteLength)] = makePanic(msg)

	//-- type2                  (byte string)
	// 1st-byte 0..27
	for i := 0; i < 28; i++ {
		cbor2collateM[cborHdr(cborType2, byte(i))] = collateCborT2
	}
	// 1st-byte 28..31
	cbor2collateM[cborHdr(cborType2, 28)] = collateCborT2
	cbor2collateM[cborHdr(cborType2, 29)] = collateCborT2
	cbor2collateM[cborHdr(cborType2, 30)] = collateCborT2
	cbor2collateM[cborHdr(cborType2, cborIndefiniteLength)] = makePanic(msg)

	//-- type3                  (string)
	// 1st-byte 0..27
	for i := 0; i < 28; i++ {
		cbor2collateM[cborHdr(cborType3, byte(i))] = collateCborT3
	}

	// 1st-byte 28..31
	cbor2collateM[cborHdr(cborType3, 28)] = collateCborT3
	cbor2collateM[cborHdr(cborType3, 29)] = collateCborT3
	cbor2collateM[cborHdr(cborType3, 30)] = collateCborT3
	msg = "cbor->collate indefinite string not supported"
	cbor2collateM[cborHdr(cborType3, cborIndefiniteLength)] = makePanic(msg)

	//-- type4                  (array)
	// 1st-byte 0..27
	for i := 0; i < 28; i++ {
		cbor2collateM[cborHdr(cborType4, byte(i))] = collateCborT4
	}
	// 1st-byte 28..31
	cbor2collateM[cborHdr(cborType4, 28)] = collateCborT4
	cbor2collateM[cborHdr(cborType4, 29)] = collateCborT4
	cbor2collateM[cborHdr(cborType4, 30)] = collateCborT4
	cbor2collateM[cborHdr(cborType4, cborIndefiniteLength)] = collateCborT4Indef

	//-- type5                  (map)
	// 1st-byte 0..27
	for i := 0; i < 28; i++ {
		cbor2collateM[cborHdr(cborType5, byte(i))] = collateCborT5
	}
	// 1st-byte 28..31
	cbor2collateM[cborHdr(cborType5, 28)] = collateCborT5
	cbor2collateM[cborHdr(cborType5, 29)] = collateCborT5
	cbor2collateM[cborHdr(cborType5, 30)] = collateCborT5
	cbor2collateM[cborHdr(cborType5, cborIndefiniteLength)] = collateCborT5Indef

	//-- type6
	// 1st-byte 0..23
	for i := byte(0); i < cborInfo24; i++ {
		cbor2collateM[cborHdr(cborType6, i)] = collateCborTag
	}
	// 1st-byte 24..27
	cbor2collateM[cborHdr(cborType6, cborInfo24)] = collateCborTag
	cbor2collateM[cborHdr(cborType6, cborInfo25)] = collateCborTag
	cbor2collateM[cborHdr(cborType6, cborInfo26)] = collateCborTag
	cbor2collateM[cborHdr(cborType6, cborInfo27)] = collateCborTag
	// 1st-byte 28..31
	msg = "cbor->collate type6 decode reserved info"
	cbor2collateM[cborHdr(cborType6, 28)] = makePanic(msg)
	cbor2collateM[cborHdr(cborType6, 29)] = makePanic(msg)
	cbor2collateM[cborHdr(cborType6, 30)] = makePanic(msg)
	msg = "cbor->collate indefinite type6 not supported"
	cbor2collateM[cborHdr(cborType6, cborIndefiniteLength)] = makePanic(msg)

	//-- type7                  (simple values / floats / break-stop)
	msg = "cbor->collate simple-type < 20 not supported"
	// 1st-byte 0..19
	for i := byte(0); i < 20; i++ {
		cbor2collateM[cborHdr(cborType7, i)] = makePanic(msg)
	}
	// 1st-byte 20..23
	cbor2collateM[cborHdr(cborType7, cborSimpleTypeFalse)] = collateCborFalse
	cbor2collateM[cborHdr(cborType7, cborSimpleTypeTrue)] = collateCborTrue
	cbor2collateM[cborHdr(cborType7, cborSimpleTypeNil)] = collateCborNull
	msg = "cbor->collate simple-type-undefined not supported"
	cbor2collateM[cborHdr(cborType7, cborSimpleUndefined)] = makePanic(msg)

	msg = "cbor->collate simple-type > 31 not supported"
	cbor2collateM[cborHdr(cborType7, cborSimpleTypeByte)] = makePanic(msg)
	msg = "cbor->collate float16 not supported"
	cbor2collateM[cborHdr(cborType7, cborFlt16)] = makePanic(msg)
	cbor2collateM[cborHdr(cborType7, cborFlt32)] = collateCborFloat32
	cbor2collateM[cborHdr(cborType7, cborFlt64)] = collateCborFloat64
	// 1st-byte 28..31
	msg = "cbor->collate simple-type 28 not supported"
	cbor2collateM[cborHdr(cborType7, 28)] = makePanic(msg)
	msg = "cbor->collate simple-type 29 not supported"
	cbor2collateM[cborHdr(cborType7, 29)] = makePanic(msg)
	msg = "cbor->collate simple-type 30 not supported"
	cbor2collateM[cborHdr(cborType7, 30)] = makePanic(msg)
	msg = "cbor->collate simple-type break-code not supported"
	cbor2collateM[cborHdr(cborType7, cborItemBreak)] = makePanic(msg)
}
