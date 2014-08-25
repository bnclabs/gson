package json

import (
    "unicode/utf8"
)

func parsePointer(s []byte) [][]byte {
    if len(s) == 0 {
        return [][]byte{}
    }

    u := [6]byte{}

    parts := make([][]byte, 0, 3)
    part := make([]byte, 0, len(s))
    for i := 0; i < len(s); {
        if s[i] == '~' {
            if s[i+1] == '1' {
                part = append(part, '/')
                i += 2

            } else if s[i+1] == '0' {
                part = append(part, '~')
                i += 2
            }

        } else if  s[i] == '/' {
            if len(part) > 0 {
                parts = append(parts, part)
                part = make([]byte, 0, len(s))
            }
            i++

        } else if s[i] < utf8.RuneSelf {
            part = append(part, s[i])
            i++

        } else {
            r, size := utf8.DecodeRune(s[i:])
            i += size
            parti := utf8.EncodeRune(u[:], r)
            part = append(part, u[:parti]...)
        }
    }

    return append(parts, part)
}

func encodePointer(p []string, out []byte) string {
    for _, s := range p {
        out = append(out, '/')
        for _, c := range []byte(s) {
            switch c {
            case '/':
                out = append(out, '~', '1')
            case '~':
                out = append(out, '~', '0')
            default:
                out = append(out, c)
            }
        }
    }
    return string(out)
}
