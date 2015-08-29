// +build ignore

package collate

////---- CBOR to JSON convertor
//
//func collateCbor(in, out []byte, config *Config) (int, int) {
//    n, m := cborTocollate[in[0]](in, out, config)
//    return n, m
//}
//
//func collateCborNull(buf, out []byte, config *Config) (int, int) {
//    out[n], out[n+1] = TypeNull, Terminator
//    return 1, 2
//}
//
//func collateCborTrue(buf, out []byte, config *Config) (int, int) {
//    out[n], out[n+1] = TypeTrue, Terminator
//    return 1, 2
//}
//
//func collateCborFalse(buf, out []byte, config *Config) (int, int) {
//    out[n], out[n+1] = TypeFalse, Terminator
//    return 1, 2
//}
//
//func collateCborFloat32(buf, out []byte, config *Config) (int, int) {
//    item := uint64(binary.BigEndian.Uint32(buf[1:]))
//    f, n := math.Float32frombits(uint32(item)), 0
//    out[n] = TypeNumber; n++
//    n += normalizeFloat(float64(f), out[n:], config.nt)
//    out[n] = Terminator, n++
//    return 5, n
//}
//
//func collateCborFloat64(buf, out []byte, config *Config) (int, int) {
//    item := uint64(binary.BigEndian.Uint64(buf[1:]))
//    f, n := math.Float64frombits(item), 0
//    out[n] = TypeNumber; n++
//    n += normalizeFloat(float64(f), out[n:], config.nt)
//    out[n] = Terminator; n++
//    return 9, n
//}
//
//func collateCborType0SmallInt(buf, out []byte, config *Config) (int, int) {
//    n := 0
//    out[n] = TypeNumber; n++
//    n += normalizeFloat(int64(info(buf[0])), out[n:], config.nt)
//    out[n] = Terminator; n++
//    return 1, n
//}
//
//func collateCborType1SmallInt(buf, out []byte, config *Config) (int, int) {
//    n := 0
//    out[n] = TypeNumber; n++
//    n += normalizeFloat(-int64(info(buf[0])+1), out[n:], config.nt)
//    out[n] = Terminator; n++
//    return 1, n
//}
//
//func collateCborType0Info24(buf, out []byte, config *Config) (int, int) {
//    n := 0
//    out[n] = TypeNumber; n++
//    n += normalizeFloat(int64(buf[1]), out[n:], config.nt)
//    out[n] = Terminator; n++
//    return 2, n
//}
//
//func collateCborType1Info24(buf, out []byte, config *Config) (int, int) {
//    n := 0
//    out[n] = TypeNumber; n++
//    n += normalizeFloat(-int64(buf[1]+1), out[n:], config.nt)
//    out[n] = Terminator; n++
//    return 2, n
//}
//
//func collateCborType0Info25(buf, out []byte, config *Config) (int, int) {
//    n := 0
//    out[n] = TypeNumber; n++
//    i := int64(binary.BigEndian.Uint16(buf[1:]))
//    n += normalizeFloat(i, out[n:], config.nt)
//    out[n] = Terminator; n++
//    return 3, n
//}
//
//func collateCborType1Info25(buf, out []byte, config *Config) (int, int) {
//    n := 0
//    out[n] = TypeNumber; n++
//    i := -int64(binary.BigEndian.Uint16(buf[1:]) + 1)
//    n += normalizeFloat(i, out[n:], config.nt)
//    out[n] = Terminator; n++
//    return 3, n
//}
//
//func collateCborType0Info26(buf, out []byte, config *Config) (int, int) {
//    n := 0
//    out[n] = TypeNumber; n++
//    i := int64(binary.BigEndian.Uint32(buf[1:]))
//    n += normalizeFloat(i, out[n:], config.nt)
//    out[n] = Terminator; n++
//    return 5, n
//}
//
//func collateCborType1Info26(buf, out []byte, config *Config) (int, int) {
//    n := 0
//    out[n] = TypeNumber; n++
//    i := -int64(binary.BigEndian.Uint32(buf[1:]) + 1)
//    n += normalizeFloat(i, out[n:], config.nt)
//    out[n] = Terminator; n++
//    return 5, n
//}
//
//func collateCborType0Info27(buf, out []byte, config *Config) (int, int) {
//    n := 0
//    out[n] = TypeNumber; n++
//    i := int64(binary.BigEndian.Uint64(buf[1:]))
//    n += normalizeFloat(i, out[n:], config.nt)
//    out[n] = Terminator; n++
//    return 9, n
//}
//
//func collateCborType1Type1Info27(buf, out []byte, config *Config) (int, int) {
//    x := uint64(binary.BigEndian.Uint64(buf[1:]))
//    if x > 9223372036854775807 {
//        panic("cbo->json number exceeds the limit of int64")
//    }
//    val, n := (int64(-x) - 1), 0
//    out[n] = TypeNumber; n++
//    n += normalizeFloat(val, out[n:], config.nt)
//    out[n] = Terminator; n++
//    return 9, len(out)
//}
//
//func collateCborLength(length int, out []byte, config *Config) int {
//    n := 0
//    out[n] = TypeLength; n++
//    n += normalizeFloat(int64(length), out[n:], config.nt)
//    out[n] = Terminator; n++
//    return n
//}
//
//func collateCborType3(buf, out []byte, config *Config) (int, int) {
//    ln, n := decodeLength(buf)
//    if config.doMissing && MissingLiteral.Equal(bytes2str(buf[n:n+ln])) {
//        out[0], out[1] = TypeMissing, Terminator
//        return n+ln, 2
//    }
//    m := 0
//    out[m] = TypeString; m++
//    m += suffixEncodeString(buf[n:ln], out[m:])
//    out[m] = Terminator; m++
//    return n+ln, m
//}
//
//func collateCborType4(buf, out []byte, config *Config) (int, int) {
//    ln, n := decodeLength(buf)
//    m := 0
//    out[m] = TypeArray; m++
//    if config.arrayLenPrefix {
//        m += collateCborLength(ln, out[m:], config)
//    }
//    for ; ln > 0; ln-- {
//        x, y := collateCbor(buf[n:], out[m:], config)
//        m, n = m+y, n+x
//    }
//    out[m] = Terminator; m++
//    return n, m
//}
//
//func collateCborType4Indefinite(buf, out []byte) (n int, m int) {
//    ln := 0
//    out[m] = TypeArray; m++
//    m_, m__ := m, m
//    if config.arrayLenPrefix {
//        m_, m__ = m+32, m+32 // length encoding can go upto max of 32 bytes
//    }
//    n = 1
//    defer func() {
//        if config.arrayLenPrefix {
//            m += collateCborLength(ln, out[m:], config)
//        }
//        copy(out[m:], out[m_:m__])
//        m += (m__ - m_)
//        out[m] = Terminator; m++
//        return
//    }()
//    if buf[n] == brkstp {
//        n++
//        return
//    }
//    for buf[n] != brkstp {
//        x, y := collateCbor(buf[n:], out[m__:], config)
//        m__, n = m__+y, n+x
//    }
//    return
//}
//
//func collateCborType5(buf, out []byte, config *Config) (n int, m int) {
//    ln, n := decodeLength(buf)
//    out[0] = '{'
//    if ln == 0 {
//        out[1] = '}'
//        return n, 2
//    }
//    m := 1
//    for ; ln > 0; ln-- {
//        x, y := cborTojson[buf[n]](buf[n:], out[m:])
//        m, n = m+y, n+x
//        out[m], m = ':', m+1
//
//        x, y = cborTojson[buf[n]](buf[n:], out[m:])
//        m, n = m+y, n+x
//        out[m], m = ',', m+1
//    }
//    out[m-1] = '}'
//    return n, m
//
//    var ln int
//    ln, n = decodeLength(buf)
//    out[m] = TypeObj; m++
//    if config.propertyLenPrefix {
//        m += collateCborLength(ln, out[m:], config)
//    }
//    altcode, p := make([]byte, 10*1024), 0
//    refs, i := make(kvrefs, 10*256), 0
//    for {
//        // NOTE: empty string is also a valid key
//        key, remtxt := scanString(str2bytes(txt))
//        txt = bytes2str(remtxt)
//        if txt = skipWS(txt, config.ws); len(txt) == 0 || txt[0] != ':' {
//            panic("collate scanner expectedColon")
//        }
//        x, txt = scanToken(skipWS(txt[1:], config.ws), altcode[p:], config)
//        refs[i] = kvref{bytes2str(key), altcode[p : p+x]}
//        p += x
//        i++
//
//        if txt = skipWS(txt, config.ws); len(txt) == 0 {
//            panic("collate scanner expectedCloseobject")
//        } else if txt[0] == ',' {
//            txt = skipWS(txt[1:], config.ws)
//        } else if txt[0] == '}' {
//            break
//        } else {
//            panic("collate scanner expectedCloseobject")
//        }
//        ln++
//    }
//            sort.Sort(refs[:i])
//            for _, kv := range refs {
//                n__ += gson2collate(kv.key, code[n__:], config) // encode key
//                copy(code[n__:], kv.code)
//                n__ += len(kv.code)
//            }
//        }
//        if config.propertyLenPrefix {
//            n += gson2collate(Length(ln), code[n:], config)
//            copy(code[n:], code[n_:n__])
//            n += (n__ - n_)
//        } else {
//            n = n__
//        }
//        code[n] = Terminator; n++
//        return n, txt[1:]
//}
//
//func decodeType5IndefiniteTojson(buf, out []byte) (int, int) {
//    out[0] = '{'
//    if buf[1] == brkstp {
//        out[1] = '}'
//        return 2, 2
//    }
//    n, m := 1, 1
//    for buf[n] != brkstp {
//        x, y := cborTojson[buf[n]](buf[n:], out[m:])
//        m, n = m+y, n+x
//        out[m], m = ':', m+1
//
//        x, y = cborTojson[buf[n]](buf[n:], out[m:])
//        m, n = m+y, n+x
//        out[m], m = ',', m+1
//    }
//    out[m-1] = '}'
//    return n + 1, m
//}
//
//func decodeTagTojson(buf, out []byte) (int, int) {
//    byt := (buf[0] & 0x1f) | type0 // fix as positive num
//    item, n := cborDecoders[byt](buf)
//    switch item.(uint64) {
//    case tagJsonString:
//        ln, m := decodeLength(buf[n:])
//        n += m
//        out[0] = '"'
//        copy(out[1:], buf[n:n+ln])
//        out[ln+1] = '"'
//        return n + ln, ln + 2
//    case tagJsonNumber:
//        ln, m := decodeLength(buf[n:])
//        n += m
//        copy(out, buf[n:n+ln])
//        return n + ln, ln
//    }
//    return n, 0 // skip this tag
//}
//
//// ---- decoders
//
//var cborTojson = make(map[byte]func([]byte, []byte) (int, int))
//
//func init() {
//    makePanic := func(msg string) func([]byte, []byte) (int, int) {
//        return func(_, _ []byte) (int, int) { panic(msg) }
//    }
//    //-- type0                  (unsigned integer)
//    // 1st-byte 0..23
//    for i := byte(0); i < info24; i++ {
//        cborTojson[hdr(type0, i)] = decodeType0SmallIntTojson
//    }
//    // 1st-byte 24..27
//    cborTojson[hdr(type0, info24)] = decodeType0Info24Tojson
//    cborTojson[hdr(type0, info25)] = decodeType0Info25Tojson
//    cborTojson[hdr(type0, info26)] = decodeType0Info26Tojson
//    cborTojson[hdr(type0, info27)] = decodeType0Info27Tojson
//    // 1st-byte 28..31
//    msg := "cbor->json decode type0 reserved info"
//    cborTojson[hdr(type0, 28)] = makePanic(msg)
//    cborTojson[hdr(type0, 29)] = makePanic(msg)
//    cborTojson[hdr(type0, 30)] = makePanic(msg)
//    cborTojson[hdr(type0, indefiniteLength)] = makePanic(msg)
//
//    //-- type1                  (signed integer)
//    // 1st-byte 0..23
//    for i := byte(0); i < info24; i++ {
//        cborTojson[hdr(type1, i)] = decodeType1SmallIntTojson
//    }
//    // 1st-byte 24..27
//    cborTojson[hdr(type1, info24)] = decodeType1Info24Tojson
//    cborTojson[hdr(type1, info25)] = decodeType1Info25Tojson
//    cborTojson[hdr(type1, info26)] = decodeType1Info26Tojson
//    cborTojson[hdr(type1, info27)] = decodeType1Info27Tojson
//    // 1st-byte 28..31
//    msg = "cbor->json type1 decode reserved info"
//    cborTojson[hdr(type1, 28)] = makePanic(msg)
//    cborTojson[hdr(type1, 29)] = makePanic(msg)
//    cborTojson[hdr(type1, 30)] = makePanic(msg)
//    cborTojson[hdr(type1, indefiniteLength)] = makePanic(msg)
//
//    //-- type2                  (byte string)
//    // 1st-byte 0..27
//    msg = "cbor->json byte string not supported"
//    for i := 0; i < 28; i++ {
//        cborTojson[hdr(type2, byte(i))] = makePanic(msg)
//    }
//    // 1st-byte 28..31
//    cborTojson[hdr(type2, 28)] = makePanic(msg)
//    cborTojson[hdr(type2, 29)] = makePanic(msg)
//    cborTojson[hdr(type2, 30)] = makePanic(msg)
//    cborTojson[hdr(type2, indefiniteLength)] = makePanic(msg)
//
//    //-- type3                  (string)
//    // 1st-byte 0..27
//    for i := 0; i < 28; i++ {
//        cborTojson[hdr(type3, byte(i))] = decodeType3Tojson
//    }
//    // 1st-byte 28..31
//    cborTojson[hdr(type3, 28)] = decodeType3Tojson
//    cborTojson[hdr(type3, 29)] = decodeType3Tojson
//    cborTojson[hdr(type3, 30)] = decodeType3Tojson
//    msg = "cbor->json indefinite string not supported"
//    cborTojson[hdr(type3, indefiniteLength)] = makePanic(msg)
//
//    //-- type4                  (array)
//    // 1st-byte 0..27
//    for i := 0; i < 28; i++ {
//        cborTojson[hdr(type4, byte(i))] = decodeType4Tojson
//    }
//    // 1st-byte 28..31
//    cborTojson[hdr(type4, 28)] = decodeType4Tojson
//    cborTojson[hdr(type4, 29)] = decodeType4Tojson
//    cborTojson[hdr(type4, 30)] = decodeType4Tojson
//    cborTojson[hdr(type4, indefiniteLength)] = decodeType4IndefiniteTojson
//
//    //-- type5                  (map)
//    // 1st-byte 0..27
//    for i := 0; i < 28; i++ {
//        cborTojson[hdr(type5, byte(i))] = decodeType5Tojson
//    }
//    // 1st-byte 28..31
//    cborTojson[hdr(type5, 28)] = decodeType5Tojson
//    cborTojson[hdr(type5, 29)] = decodeType5Tojson
//    cborTojson[hdr(type5, 30)] = decodeType5Tojson
//    cborTojson[hdr(type5, indefiniteLength)] = decodeType5IndefiniteTojson
//
//    //-- type6
//    // 1st-byte 0..23
//    for i := byte(0); i < info24; i++ {
//        cborTojson[hdr(type6, i)] = decodeTagTojson
//    }
//    // 1st-byte 24..27
//    cborTojson[hdr(type6, info24)] = decodeTagTojson
//    cborTojson[hdr(type6, info25)] = decodeTagTojson
//    cborTojson[hdr(type6, info26)] = decodeTagTojson
//    cborTojson[hdr(type6, info27)] = decodeTagTojson
//    // 1st-byte 28..31
//    msg = "cbor->json type6 decode reserved info"
//    cborTojson[hdr(type6, 28)] = makePanic(msg)
//    cborTojson[hdr(type6, 29)] = makePanic(msg)
//    cborTojson[hdr(type6, 30)] = makePanic(msg)
//    msg = "cbor->json indefinite type6 not supported"
//    cborTojson[hdr(type6, indefiniteLength)] = makePanic(msg)
//
//    //-- type7                  (simple values / floats / break-stop)
//    msg = "cbor->json simple-type < 20 not supported"
//    // 1st-byte 0..19
//    for i := byte(0); i < 20; i++ {
//        cborTojson[hdr(type7, i)] = makePanic(msg)
//    }
//    // 1st-byte 20..23
//    cborTojson[hdr(type7, simpleTypeFalse)] = decodeFalseTojson
//    cborTojson[hdr(type7, simpleTypeTrue)] = decodeTrueTojson
//    cborTojson[hdr(type7, simpleTypeNil)] = decodeNullTojson
//    msg = "cbor->json simple-type-undefined not supported"
//    cborTojson[hdr(type7, simpleUndefined)] = makePanic(msg)
//
//    msg = "cbor->json simple-type > 31 not supported"
//    cborTojson[hdr(type7, simpleTypeByte)] = makePanic(msg)
//    msg = "cbor->json float16 not supported"
//    cborTojson[hdr(type7, flt16)] = makePanic(msg)
//    cborTojson[hdr(type7, flt32)] = decodeFloat32Tojson
//    cborTojson[hdr(type7, flt64)] = decodeFloat64Tojson
//    // 1st-byte 28..31
//    msg = "cbor->json simple-type 28 not supported"
//    cborTojson[hdr(type7, 28)] = makePanic(msg)
//    msg = "cbor->json simple-type 29 not supported"
//    cborTojson[hdr(type7, 29)] = makePanic(msg)
//    msg = "cbor->json simple-type 30 not supported"
//    cborTojson[hdr(type7, 30)] = makePanic(msg)
//    msg = "cbor->json simple-type break-code not supported"
//    cborTojson[hdr(type7, itemBreak)] = makePanic(msg)
//}
//
//var intCheck = [256]byte{}
//var numCheck = [256]byte{}
//var fltCheck = [256]byte{}
//
//func init() {
//    for i := 48; i <= 57; i++ {
//        intCheck[i] = 1
//        numCheck[i] = 1
//    }
//    intCheck['-'] = 1
//    intCheck['+'] = 1
//    intCheck['.'] = 1
//    intCheck['e'] = 1
//    intCheck['E'] = 1
//
//    numCheck['-'] = 1
//    numCheck['+'] = 1
//    numCheck['.'] = 1
//
//    fltCheck['.'] = 1
//    fltCheck['e'] = 1
//    fltCheck['E'] = 1
//}
