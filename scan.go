package json

import (
    "bytes"
    "strconv"
    "unicode"
    "unicode/utf8"
    "unicode/utf16"
)

type Number string

var nullLiteral = []byte("null")
var trueLiteral = []byte("true")
var falseLiteral = []byte("false")

var escapeCode = [256]byte{ // TODO: size can be optimized
    '"': '"',
    '\\': '\\',
    '/':  '/',
    '\'': '\'',
    'b':  '\b',
    'f':  '\f',
    'n':  '\n',
    'r':  '\r',
    't':  '\t',
}

var spaceCode = [256]byte{ // TODO: size can be optimized
    '\t': 1,
    '\n': 1,
    '\v': 1,
    '\f': 1,
    '\r': 1,
    ' ': 1,
}

func scanToken(txt []byte, nk NumberKind, ws SpaceKind, jsonp bool) (interface{}, []byte, []string, error) {
    var err error

    if len(txt) < 1 {
        return nil, txt, nil, nil
    }

    for {
        if digitCheck[txt[0]] == 1 {
            n, remntxt, err := scanNum(txt, nk)
            return n, remntxt, nil, err
        }

        switch txt[0] {
        case 'n':
            if bytes.Compare(txt[:4], nullLiteral) == 0 {
                return nil, txt[4:], nil, nil
            }
            return nil, txt, nil, ErrorScan

        case 't':
            if bytes.Compare(txt[:4], trueLiteral) == 0 {
                return true, txt[4:], nil, nil
            }
            return nil, txt, nil, ErrorScan

        case 'f':
            if bytes.Compare(txt[:5], falseLiteral) == 0 {
                return false, txt[5:], nil, nil
            }
            return nil, txt, nil, ErrorScan

        case '-':
            n, remntxt, err := scanNum(txt, nk)
            return n, remntxt, nil, err

        case '"':
            s, txt, err := scanString(txt)
            return string(s), txt, nil, err

        case '[':
            var jsonpIdx int
            var pointers, childPointers []string
            var tok interface{}

            arr := make([]interface{}, 0)
            txt = skipWS(txt[1:], ws)
            if jsonp {
                pointers = make([]string, 0, 4)
            }
            for {
                if txt[0] == ']' {
                    break
                }
                tok, txt, childPointers, err = scanToken(txt, nk, ws, jsonp)
                if err != nil {
                    return nil, txt, nil, err
                }
                arr = append(arr, tok)
                if jsonp && childPointers != nil {
                    prefix := "/" + strconv.Itoa(jsonpIdx)
                    appendPointers(pointers, prefix, childPointers)
                }
                txt = skipWS(txt, ws)
                if txt[0] == ',' {
                    txt = skipWS(txt[1:], ws)
                }
            }
            return arr, txt[1:], pointers, nil

        case '{':
            var key []byte
            var childPointers, pointers []string
            var tok interface{}

            m := make(map[string]interface{})
            txt = skipWS(txt[1:], ws)
            if jsonp {
                pointers = make([]string, 0, 8)
                pointers = append(pointers, "")
            }
            for {
                if txt[0] == '}' {
                    break
                }
                key, txt, err = scanString(txt)
                if err != nil {
                    return nil, txt, nil, err
                } else if len(key) < 1 {
                    return nil, txt, nil, ErrorScan
                }

                txt = skipWS(txt, ws)
                if txt[0] != ':' {
                    return nil, txt, nil, ErrorScan
                }
                tok, txt, childPointers, err =
                                scanToken(skipWS(txt[1:], ws), nk, ws, jsonp)
                if err != nil {
                    return nil, txt, nil, err
                }
                m[string(key)] = tok
                if jsonp {
                    appendPointers(pointers, "/" + string(key), childPointers)
                }

                txt = skipWS(txt, ws)
                if txt[0] == ',' {
                    txt = skipWS(txt[1:], ws)
                }
            }
            return m, txt[1:], pointers, nil

        default:
            return nil, txt, nil, ErrorScan
        }
    }
}

func scanNum(txt []byte, k NumberKind) (interface{}, []byte, error) {
    s, e, l := 0, 1, len(txt)
    if len(txt) > 1 {
        for ; e < l  &&  intCheck[txt[e]] == 1; e++ {}
    }

    switch k {
    case StringNumber:
        return Number(txt[s:e]), txt[e:], nil

    case IntNumber:
        num, err := strconv.Atoi(string(txt[s:e]))
        return num, txt[e:], err

    case FloatNumber:
        num, err := strconv.ParseFloat(string(txt[s:e]), 64)
        return num, txt[e:], err
    }
    return nil, nil, ErrorScan
}

func scanString(txt []byte) ([]byte, []byte, error) {
    if len(txt) < 2 {
        return nil, nil, ErrorScan
    }

    e := 1
    for ; txt[e] != '"'; {
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
            return nil, nil, ErrorScan
        }
        e += size
    }

    if txt[e] == '"' { // done we have nothing to unquote
        return txt[1:e], txt[e+1:], nil
    }

    out := make([]byte, len(txt) + 2*utf8.UTFMax)
    oute := copy(out, txt[:e]) // copy so far

loop:
    for ; e < len(txt) ; {
        switch c := txt[e]; {
        case c == '"':
            out[oute] = c
            e++
            break loop

        case c == '\\':
            if txt[e+1] == 'u' {
                r := getu4(txt[e:])
                if r < 0 { // invalid
                    return nil, nil, ErrorScan
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
            return nil, nil, ErrorScan

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
        return out[1:oute], txt[e:], nil
    }
    return nil, nil, ErrorScan
}

func skipWS(txt []byte, ws SpaceKind) []byte {
    switch ws {
    case UnicodeSpace:
        for {
            r, i := utf8.DecodeRune(txt)
            if unicode.IsSpace(r) {
                txt = txt[i:]
                continue
            }
            break
        }

    case AnsiSpace:
        for ; spaceCode[txt[0]] == 1; {
            txt = txt[1:]
        }
    }
    return txt
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

func appendPointers(pointers []string, prefix string, childPointers []string) []string {
    pointers = append(pointers, prefix)
    for _, cpointer := range childPointers {
        pointers = append(pointers, prefix + cpointer)
    }
    return pointers
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
