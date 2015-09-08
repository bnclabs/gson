package gson

const maxPartSize int = 1024

func jptrToCbor(jsonptr, out []byte) int {
	var part [maxPartSize]byte

	n, off := textStart(out), 0
	for i := 0; i < len(jsonptr); {
		if jsonptr[i] == '~' {
			if jsonptr[i+1] == '1' {
				part[off] = '/'
				off, i = off+1, i+2

			} else if jsonptr[i+1] == '0' {
				part[off] = '~'
				off, i = off+1, i+2
			}

		} else if jsonptr[i] == '/' {
			if off > 0 {
				n += tag2cbor(uint64(tagJsonString), out[n:])
				n += valtext2cbor(bytes2str(part[:off]), out[n:])
				off = 0
			}
			i++

		} else {
			part[off] = jsonptr[i]
			i, off = i+1, off+1
		}
	}
	if off > 0 || (len(jsonptr) > 0 && jsonptr[len(jsonptr)-1] == '/') {
		n += tag2cbor(uint64(tagJsonString), out[n:])
		n += valtext2cbor(bytes2str(part[:off]), out[n:])
	}

	n += breakStop(out[n:])
	return n
}

func cborToJptr(cborptr, out []byte) int {
	i, n := 1, 0
	for {
		byt := cborHdr(cborType6, cborInfo24)
		if cborptr[i] == byt && cborptr[i+1] == tagJsonString {
			i, out[n] = i+2, '/'
			n += 1
			ln, j := cborItemLength(cborptr[i:])
			ln, i = ln+i+j, i+j
			for i < ln {
				switch cborptr[i] {
				case '/':
					out[n], out[n+1] = '~', '1'
					n += 2
				case '~':
					out[n], out[n+1] = '~', '0'
					n += 2
				default:
					out[n] = cborptr[i]
					n += 1
				}
				i++
			}
		}
		if cborptr[i] == brkstp {
			break
		}
	}
	return n
}
