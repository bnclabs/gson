package collate

import "math"
import "strconv"
import "sort"
import "encoding/binary"

//---- collate to cbor

func collate2cbor(code []byte, out []byte, config *Config) (int, int) {
	if len(code) == 0 {
		return 0, 0
	}
	var scratch [64]byte
	m, n := 1, 0
	switch code[0] {
	case TypeMissing:
		n += encodeText(string(MissingLiteral), out[n:])
		return m + 1, n

	case TypeNull:
		n := encodeNull(out[n:])
		return m + 1, n

	case TypeTrue:
		n += encodeTrue(out[n:])
		return m + 1, n

	case TypeFalse:
		n += encodeFalse(out[n:])
		return m + 1, n

	case TypeNumber:
		x := getDatum(code[m:])
		// -1 is to skip terminator
		num := denormalizeFloat(code[m:m+x-1], config.nt)
		switch v := num.(type) {
		case float64:
			n += encodeFloat64(v, out[n:])
		case int64:
			n += encodeInt64(v, out[n:])
		}
		return m + x, n

	case TypeString:
		scratch := make([]byte, len(code[m:]))
		x, y := suffixDecodeString(code[m:], scratch)
		n := encodeText(bytes2str(scratch[:y]), out[n:])
		return m + x, n

	case TypeArray:
		if config.arrayLenPrefix {
			if code[m] != TypeLength {
				panic("collate decode expected array length prefix")
			}
			x := getDatum(code[m:])
			// -1 skip terminator
			decodeInt(code[m:m+x-1], scratch[:])
			m += x
		}
		n += encodeArrayStart(out[n:])
		for code[m] != Terminator {
			x, y := collate2cbor(code[m:], out[n:], config)
			m, n = m+x, n+y
		}
		n += encodeBreakStop(out[n:])
		return m + 1, n

	case TypeObj:
		if config.propertyLenPrefix {
			if code[m] != TypeLength {
				panic("collate decode expected property length prefix")
			}
			x := getDatum(code[m:])
			// -1 skip terminator
			decodeInt(code[m:m+x-1], scratch[:])
			m += x
		}
		n += encodeMapStart(out[n:])
		for code[m] != Terminator {
			x, y := collate2cbor(code[m:], out[n:], config)
			m, n = m+x, n+y
			x, y = collate2cbor(code[m:], out[n:], config)
			m, n = m+x, n+y
		}
		n += encodeBreakStop(out[n:])
		return m + 1, n
	}
	panic("collate decode to cbor invalid binary")
}

func encodeNull(buf []byte) int {
	buf[0] = hdr(type7, simpleTypeNil)
	return 1
}

func encodeTrue(buf []byte) int {
	buf[0] = hdr(type7, simpleTypeTrue)
	return 1
}

func encodeFalse(buf []byte) int {
	buf[0] = hdr(type7, simpleTypeFalse)
	return 1
}

func encodeInt8(item int8, buf []byte) int {
	if item > maxSmallInt {
		buf[0] = hdr(type0, info24)
		buf[1] = byte(item) // 24..127
		return 2
	} else if item < -maxSmallInt {
		buf[0] = hdr(type1, info24)
		buf[1] = byte(-(item + 1)) // -128..-24
		return 2
	} else if item < 0 {
		buf[0] = hdr(type1, byte(-(item + 1))) // -23..-1
		return 1
	}
	buf[0] = hdr(type0, byte(item)) // 0..23
	return 1
}

func encodeInt16(item int16, buf []byte) int {
	if item > 127 {
		if item < 256 {
			buf[0] = hdr(type0, info24)
			buf[1] = byte(item) // 128..255
			return 2
		}
		buf[0] = hdr(type0, info25)
		binary.BigEndian.PutUint16(buf[1:], uint16(item)) // 256..32767
		return 3

	} else if item < -128 {
		if item > -256 {
			buf[0] = hdr(type1, info24)
			buf[1] = byte(-(item + 1)) // -255..-129
			return 2
		}
		buf[0] = hdr(type1, info25) // -32768..-256
		binary.BigEndian.PutUint16(buf[1:], uint16(-(item + 1)))
		return 3
	}
	return encodeInt8(int8(item), buf)
}

func encodeInt32(item int32, buf []byte) int {
	if item > 32767 {
		if item < 65536 {
			buf[0] = hdr(type0, info25)
			binary.BigEndian.PutUint16(buf[1:], uint16(item)) // 32768..65535
			return 3
		}
		buf[0] = hdr(type0, info26) // 65536 to 2147483647
		binary.BigEndian.PutUint32(buf[1:], uint32(item))
		return 5

	} else if item < -32768 {
		if item > -65536 {
			buf[0] = hdr(type1, info25) // -65535..-32769
			binary.BigEndian.PutUint16(buf[1:], uint16(-(item + 1)))
			return 3
		}
		buf[0] = hdr(type1, info26) // -2147483648..-65536
		binary.BigEndian.PutUint32(buf[1:], uint32(-(item + 1)))
		return 5
	}
	return encodeInt16(int16(item), buf)
}

func encodeInt64(item int64, buf []byte) int {
	if item > 2147483647 {
		if item < 4294967296 {
			buf[0] = hdr(type0, info26) // 2147483647..4294967296
			binary.BigEndian.PutUint32(buf[1:], uint32(item))
			return 5
		}
		buf[0] = hdr(type0, info27) // 4294967296..9223372036854775807
		binary.BigEndian.PutUint64(buf[1:], uint64(item))
		return 9

	} else if item < -2147483648 {
		if item > -4294967296 {
			buf[0] = hdr(type1, info26) // -4294967295..-2147483649
			binary.BigEndian.PutUint32(buf[1:], uint32(-(item + 1)))
			return 5
		}
		buf[0] = hdr(type1, info27) // -9223372036854775808..-4294967296
		binary.BigEndian.PutUint64(buf[1:], uint64(-(item + 1)))
		return 9
	}
	return encodeInt32(int32(item), buf)
}

func encodeUint8(item byte, buf []byte) int {
	if item <= maxSmallInt {
		buf[0] = hdr(type0, item) // 0..23
		return 1
	}
	buf[0] = hdr(type0, info24)
	buf[1] = item // 24..255
	return 2
}

func encodeUint16(item uint16, buf []byte) int {
	if item < 256 {
		return encodeUint8(byte(item), buf)
	}
	buf[0] = hdr(type0, info25)
	binary.BigEndian.PutUint16(buf[1:], item) // 256..65535
	return 3
}

func encodeUint32(item uint32, buf []byte) int {
	if item < 65536 {
		return encodeUint16(uint16(item), buf) // 0..65535
	}
	buf[0] = hdr(type0, info26)
	binary.BigEndian.PutUint32(buf[1:], item) // 65536 to 4294967295
	return 5
}

func encodeUint64(item uint64, buf []byte) int {
	if item < 4294967296 {
		return encodeUint32(uint32(item), buf) // 0..4294967295
	}
	buf[0] = hdr(type0, info27) // 4294967296 to 18446744073709551615
	binary.BigEndian.PutUint64(buf[1:], item)
	return 9
}

func encodeFloat64(item float64, buf []byte) int {
	buf[0] = hdr(type7, flt64)
	binary.BigEndian.PutUint64(buf[1:], math.Float64bits(item))
	return 9
}

func encodeBytes(item []byte, buf []byte) int {
	n := encodeUint64(uint64(len(item)), buf)
	buf[0] = (buf[0] & 0x1f) | type2 // fix the type from type0->type2
	copy(buf[n:], item)
	return n + len(item)
}

func encodeText(item string, buf []byte) int {
	n := encodeBytes(str2bytes(item), buf)
	buf[0] = (buf[0] & 0x1f) | type3 // fix the type from type2->type3
	return n
}

func encodeArrayStart(buf []byte) int {
	buf[0] = hdr(type4, byte(indefiniteLength)) // indefinite length array
	return 1
}

func encodeMapStart(buf []byte) int {
	buf[0] = hdr(type5, byte(indefiniteLength)) // indefinite length map
	return 1
}

func encodeBreakStop(buf []byte) int {
	// break stop for indefinite array or map
	buf[0] = hdr(type7, byte(itemBreak))
	return 1
}

//---- CBOR to Collate

func collateCbor(in, out []byte, config *Config) (int, int) {
	return cborTocollate[in[0]](in, out, config)
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
	n += normalizeFloat(float64(f), out[n:], config.nt)
	out[n] = Terminator
	n++
	return 5, n
}

func collateCborFloat64(buf, out []byte, config *Config) (int, int) {
	item := uint64(binary.BigEndian.Uint64(buf[1:]))
	f, n := math.Float64frombits(item), 0
	out[n] = TypeNumber
	n++
	n += normalizeFloat(float64(f), out[n:], config.nt)
	out[n] = Terminator
	n++
	return 9, n
}

func collateCborType0SmallInt(buf, out []byte, config *Config) (int, int) {
	n := 0
	out[n] = TypeNumber
	n++
	n += normalizeFloat(int64(info(buf[0])), out[n:], config.nt)
	out[n] = Terminator
	n++
	return 1, n
}

func collateCborType1SmallInt(buf, out []byte, config *Config) (int, int) {
	n := 0
	out[n] = TypeNumber
	n++
	n += normalizeFloat(-int64(info(buf[0])+1), out[n:], config.nt)
	out[n] = Terminator
	n++
	return 1, n
}

func collateCborType0Info24(buf, out []byte, config *Config) (int, int) {
	n := 0
	out[n] = TypeNumber
	n++
	n += normalizeFloat(int64(buf[1]), out[n:], config.nt)
	out[n] = Terminator
	n++
	return 2, n
}

func collateCborType1Info24(buf, out []byte, config *Config) (int, int) {
	n := 0
	out[n] = TypeNumber
	n++
	n += normalizeFloat(-int64(buf[1]+1), out[n:], config.nt)
	out[n] = Terminator
	n++
	return 2, n
}

func collateCborType0Info25(buf, out []byte, config *Config) (int, int) {
	n := 0
	out[n] = TypeNumber
	n++
	i := int64(binary.BigEndian.Uint16(buf[1:]))
	n += normalizeFloat(i, out[n:], config.nt)
	out[n] = Terminator
	n++
	return 3, n
}

func collateCborType1Info25(buf, out []byte, config *Config) (int, int) {
	n := 0
	out[n] = TypeNumber
	n++
	i := -int64(binary.BigEndian.Uint16(buf[1:]) + 1)
	n += normalizeFloat(i, out[n:], config.nt)
	out[n] = Terminator
	n++
	return 3, n
}

func collateCborType0Info26(buf, out []byte, config *Config) (int, int) {
	n := 0
	out[n] = TypeNumber
	n++
	i := int64(binary.BigEndian.Uint32(buf[1:]))
	n += normalizeFloat(i, out[n:], config.nt)
	out[n] = Terminator
	n++
	return 5, n
}

func collateCborType1Info26(buf, out []byte, config *Config) (int, int) {
	n := 0
	out[n] = TypeNumber
	n++
	i := -int64(binary.BigEndian.Uint32(buf[1:]) + 1)
	n += normalizeFloat(i, out[n:], config.nt)
	out[n] = Terminator
	n++
	return 5, n
}

func collateCborType0Info27(buf, out []byte, config *Config) (int, int) {
	n := 0
	out[n] = TypeNumber
	n++
	i := int64(binary.BigEndian.Uint64(buf[1:]))
	n += normalizeFloat(i, out[n:], config.nt)
	out[n] = Terminator
	n++
	return 9, n
}

func collateCborType1Info27(buf, out []byte, config *Config) (int, int) {
	x := uint64(binary.BigEndian.Uint64(buf[1:]))
	if x > 9223372036854775807 {
		panic("cbo->collate number exceeds the limit of int64")
	}
	val, n := (int64(-x) - 1), 0
	out[n] = TypeNumber
	n++
	n += normalizeFloat(val, out[n:], config.nt)
	out[n] = Terminator
	n++
	return 9, len(out)
}

func decodeLength(buf []byte) (int, int) {
	if y := info(buf[0]); y < info24 {
		return int(y), 1
	} else if y == info24 {
		return int(buf[1]), 2
	} else if y == info25 {
		return int(binary.BigEndian.Uint16(buf[1:])), 3
	} else if y == info26 {
		return int(binary.BigEndian.Uint32(buf[1:])), 5
	}
	return int(binary.BigEndian.Uint64(buf[1:])), 9 // info27
}

func collateCborType2(buf, out []byte, config *Config) (int, int) {
	ln, m := decodeLength(buf)
	n := 0
	out[n] = TypeBinary
	n++
	copy(out[n:], buf[m:m+ln])
	n += ln
	out[n] = Terminator
	n++
	return m + ln, n
}

func collateCborType3(buf, out []byte, config *Config) (int, int) {
	ln, m := decodeLength(buf)
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
	n += normalizeFloat(int64(length), out[n:], config.nt)
	out[n] = Terminator
	n++
	return n
}

func collateCborType4(buf, out []byte, config *Config) (int, int) {
	ln, m := decodeLength(buf)
	n := 0
	out[n] = TypeArray
	n++
	if config.arrayLenPrefix {
		n += collateCborLength(ln, out[n:], config)
	}
	for ; ln > 0; ln-- {
		x, y := collateCbor(buf[m:], out[n:], config)
		m, n = m+x, n+y
	}
	out[n] = Terminator
	n++
	return m, n
}

func collateCborType4Indefinite(buf, out []byte, config *Config) (m int, n int) {
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

	if buf[1] == brkstp {
		m = 2
		return
	}
	for buf[m] != brkstp {
		x, y := collateCbor(buf[m:], out[n__:], config)
		m, n__ = m+x, n__+y
	}
	return
}

func collateCborType5(buf, out []byte, config *Config) (int, int) {
	ln, m := decodeLength(buf)
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
		x, y := collateCbor(buf[m:], altcode[p:], config) // key
		key := altcode[p : p+y]
		m, p = m+x, p+y
		x, y = collateCbor(buf[m:], altcode[p:], config) // value
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

func collateCborType5Indefinite(buf, out []byte, config *Config) (m int, n int) {
	ln := 0
	out[n] = TypeObj
	n++
	n_, n__ := n, n
	if config.propertyLenPrefix {
		n_, n__ = n+32, n+32 // length encoding can go upto max of 32 bytes
	}

	defer func() {
		if config.propertyLenPrefix {
			n += collateCborLength(ln, out[n:], config)
		}
		copy(out[n:], out[n_:n__])
		n += (n__ - n_)
		out[n] = Terminator
		n++
		return
	}()

	if buf[1] == brkstp {
		m = 2
		return
	}

	pool := getJsonKeyPool(config.maxKeys)
	altcode, p := pool.codepool.Get().([]byte), 0
	defer pool.codepool.Put(altcode)
	refs := pool.keypool.Get().(kvrefs)
	defer pool.keypool.Put(refs)

	for buf[m] != brkstp {
		x, y := collateCbor(buf[m:], altcode[p:], config) // key
		key := altcode[p : p+y]
		m, p = m+x, p+y
		x, y = collateCbor(buf[m:], altcode[p:], config) // value
		refs[ln] = kvref{bytes2str(key), altcode[p : p+y]}
		m, p = m+x, p+y
		ln++
	}
	sort.Sort(refs[:ln])
	for i := 0; i < ln; i++ {
		kv := refs[i]
		copy(out[n:], str2bytes(kv.key))
		n += len(kv.key)
		copy(out[n:], kv.code)
		n += len(kv.code)
	}
	return
}

func collateCborTag(buf, out []byte, config *Config) (int, int) {
	item, m := decodeLength(buf)
	switch uint64(item) {
	case tagJsonString:
		ln, x := decodeLength(buf[m:])
		m += x
		// copy the JSON string into scratch buffer.
		scratch := make([]byte, ln)
		scratch[0] = '"'
		copy(scratch[1:], buf[m:m+ln])
		scratch[1+ln] = '"'
		s, _ := scanString(scratch[:ln+2])
		// collate golang string.
		n := 0
		out[n] = TypeString
		n++
		n += suffixEncodeString(s, out[n:])
		out[n] = Terminator
		n++
		return m + ln, n

	case tagJsonNumber:
		ln, x := decodeLength(buf[m:])
		m += x
		f, err := strconv.ParseFloat(bytes2str(buf[m:m+ln]), 64)
		if err != nil {
			panic(err)
		}
		n := normalizeFloat(f, out, config.nt)
		return m + ln, n
	}
	return m, 0 // skip this tag
}

var cborTocollate = make(map[byte]func([]byte, []byte, *Config) (int, int))

func init() {
	makePanic := func(msg string) func([]byte, []byte, *Config) (int, int) {
		return func(_, _ []byte, _ *Config) (int, int) { panic(msg) }
	}
	//-- type0                  (unsigned integer)
	// 1st-byte 0..23
	for i := byte(0); i < info24; i++ {
		cborTocollate[hdr(type0, i)] = collateCborType0SmallInt
	}
	// 1st-byte 24..27
	cborTocollate[hdr(type0, info24)] = collateCborType0Info24
	cborTocollate[hdr(type0, info25)] = collateCborType0Info25
	cborTocollate[hdr(type0, info26)] = collateCborType0Info26
	cborTocollate[hdr(type0, info27)] = collateCborType0Info27
	// 1st-byte 28..31
	msg := "cbor->collate decode type0 reserved info"
	cborTocollate[hdr(type0, 28)] = makePanic(msg)
	cborTocollate[hdr(type0, 29)] = makePanic(msg)
	cborTocollate[hdr(type0, 30)] = makePanic(msg)
	cborTocollate[hdr(type0, indefiniteLength)] = makePanic(msg)

	//-- type1                  (signed integer)
	// 1st-byte 0..23
	for i := byte(0); i < info24; i++ {
		cborTocollate[hdr(type1, i)] = collateCborType1SmallInt
	}
	// 1st-byte 24..27
	cborTocollate[hdr(type1, info24)] = collateCborType1Info24
	cborTocollate[hdr(type1, info25)] = collateCborType1Info25
	cborTocollate[hdr(type1, info26)] = collateCborType1Info26
	cborTocollate[hdr(type1, info27)] = collateCborType1Info27
	// 1st-byte 28..31
	msg = "cbor->collate type1 decode reserved info"
	cborTocollate[hdr(type1, 28)] = makePanic(msg)
	cborTocollate[hdr(type1, 29)] = makePanic(msg)
	cborTocollate[hdr(type1, 30)] = makePanic(msg)
	cborTocollate[hdr(type1, indefiniteLength)] = makePanic(msg)

	//-- type2                  (byte string)
	// 1st-byte 0..27
	for i := 0; i < 28; i++ {
		cborTocollate[hdr(type2, byte(i))] = collateCborType2
	}
	// 1st-byte 28..31
	cborTocollate[hdr(type2, 28)] = collateCborType2
	cborTocollate[hdr(type2, 29)] = collateCborType2
	cborTocollate[hdr(type2, 30)] = collateCborType2
	cborTocollate[hdr(type2, indefiniteLength)] = makePanic(msg)

	//-- type3                  (string)
	// 1st-byte 0..27
	for i := 0; i < 28; i++ {
		cborTocollate[hdr(type3, byte(i))] = collateCborType3
	}
	// 1st-byte 28..31
	cborTocollate[hdr(type3, 28)] = collateCborType3
	cborTocollate[hdr(type3, 29)] = collateCborType3
	cborTocollate[hdr(type3, 30)] = collateCborType3
	msg = "cbor->collate indefinite string not supported"
	cborTocollate[hdr(type3, indefiniteLength)] = makePanic(msg)

	//-- type4                  (array)
	// 1st-byte 0..27
	for i := 0; i < 28; i++ {
		cborTocollate[hdr(type4, byte(i))] = collateCborType4
	}
	// 1st-byte 28..31
	cborTocollate[hdr(type4, 28)] = collateCborType4
	cborTocollate[hdr(type4, 29)] = collateCborType4
	cborTocollate[hdr(type4, 30)] = collateCborType4
	cborTocollate[hdr(type4, indefiniteLength)] = collateCborType4Indefinite

	//-- type5                  (map)
	// 1st-byte 0..27
	for i := 0; i < 28; i++ {
		cborTocollate[hdr(type5, byte(i))] = collateCborType5
	}
	// 1st-byte 28..31
	cborTocollate[hdr(type5, 28)] = collateCborType5
	cborTocollate[hdr(type5, 29)] = collateCborType5
	cborTocollate[hdr(type5, 30)] = collateCborType5
	cborTocollate[hdr(type5, indefiniteLength)] = collateCborType5Indefinite

	//-- type6
	// 1st-byte 0..23
	for i := byte(0); i < info24; i++ {
		cborTocollate[hdr(type6, i)] = collateCborTag
	}
	// 1st-byte 24..27
	cborTocollate[hdr(type6, info24)] = collateCborTag
	cborTocollate[hdr(type6, info25)] = collateCborTag
	cborTocollate[hdr(type6, info26)] = collateCborTag
	cborTocollate[hdr(type6, info27)] = collateCborTag
	// 1st-byte 28..31
	msg = "cbor->collate type6 decode reserved info"
	cborTocollate[hdr(type6, 28)] = makePanic(msg)
	cborTocollate[hdr(type6, 29)] = makePanic(msg)
	cborTocollate[hdr(type6, 30)] = makePanic(msg)
	msg = "cbor->collate indefinite type6 not supported"
	cborTocollate[hdr(type6, indefiniteLength)] = makePanic(msg)

	//-- type7                  (simple values / floats / break-stop)
	msg = "cbor->collate simple-type < 20 not supported"
	// 1st-byte 0..19
	for i := byte(0); i < 20; i++ {
		cborTocollate[hdr(type7, i)] = makePanic(msg)
	}
	// 1st-byte 20..23
	cborTocollate[hdr(type7, simpleTypeFalse)] = collateCborFalse
	cborTocollate[hdr(type7, simpleTypeTrue)] = collateCborTrue
	cborTocollate[hdr(type7, simpleTypeNil)] = collateCborNull
	msg = "cbor->collate simple-type-undefined not supported"
	cborTocollate[hdr(type7, simpleUndefined)] = makePanic(msg)

	msg = "cbor->collate simple-type > 31 not supported"
	cborTocollate[hdr(type7, simpleTypeByte)] = makePanic(msg)
	msg = "cbor->collate float16 not supported"
	cborTocollate[hdr(type7, flt16)] = makePanic(msg)
	cborTocollate[hdr(type7, flt32)] = collateCborFloat32
	cborTocollate[hdr(type7, flt64)] = collateCborFloat64
	// 1st-byte 28..31
	msg = "cbor->collate simple-type 28 not supported"
	cborTocollate[hdr(type7, 28)] = makePanic(msg)
	msg = "cbor->collate simple-type 29 not supported"
	cborTocollate[hdr(type7, 29)] = makePanic(msg)
	msg = "cbor->collate simple-type 30 not supported"
	cborTocollate[hdr(type7, 30)] = makePanic(msg)
	msg = "cbor->collate simple-type break-code not supported"
	cborTocollate[hdr(type7, itemBreak)] = makePanic(msg)
}
