package cbor

const maxPartSize int = 1024

// FromJsonPointer converts json path in RFC-6901 into cbor format,
// cbor-path :    | Text-chunk-start |
//                      | tagJsonString | text1 |
//                      | tagJsonString | text2 |
//                      ...
//                | Break-stop |
func FromJsonPointer(path []byte, out []byte) int {
	var part [maxPartSize]byte

	if len(path) > 0 && path[0] != '/' {
		panic(ErrorExpectedJsonPointer)
	}

	n, off := encodeTextStart(out), 0
	for i := 0; i < len(path); {
		if path[i] == '~' {
			if path[i+1] == '1' {
				part[off] = '/'
				off, i = off+1, i+2

			} else if path[i+1] == '0' {
				part[off] = '~'
				off, i = off+1, i+2
			}

		} else if path[i] == '/' {
			if off > 0 {
				n += encodeTag(uint64(tagJsonString), out[n:])
				n += encodeText(bytes2str(part[:off]), out[n:])
				off = 0
			}
			i++

		} else {
			part[off] = path[i]
			i, off = i+1, off+1
		}
	}
	if off > 0 || (len(path) > 0 && path[len(path)-1] == '/') {
		n += encodeTag(uint64(tagJsonString), out[n:])
		n += encodeText(bytes2str(part[:off]), out[n:])
	}

	n += encodeBreakStop(out[n:])
	return n
}

// ToJsonPointer coverts cbor encoded path into json path RFC-6901
func ToJsonPointer(bin []byte, out []byte) int {
	if !IsIndefiniteText(Indefinite(bin[0])) {
		panic(ErrorExpectedCborPointer)
	}

	i, n, brkstp := 1, 0, hdr(type7, itemBreak)
	for {
		if bin[i] == hdr(type6, info24) && bin[i+1] == tagJsonString {
			i, out[n] = i+2, '/'
			n += 1
			ln, j := decodeLength(bin[i:])
			ln, i = ln+i+j, i+j
			for i < ln {
				switch bin[i] {
				case '/':
					out[n], out[n+1] = '~', '1'
					n += 2
				case '~':
					out[n], out[n+1] = '~', '0'
					n += 2
				default:
					out[n] = bin[i]
					n += 1
				}
				i++
			}
		}
		if bin[i] == brkstp {
			break
		}
	}
	return n
}
