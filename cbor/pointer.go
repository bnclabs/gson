// Package also implements RFC-6901, JSON pointers.
// Pointers themself can be encoded into cbor format and
// vice-versa.
//
//   cbor-path :    | Text-chunk-start |
//                        | tagJsonString | len | segment1 |
//                        | tagJsonString | len | segment2 |
//                        ...
//                  | Break-stop |
package cbor

import "strconv"
import "bytes"

//import "fmt"
import "errors"

// ErrorInvalidArrayOffset
var ErrorInvalidArrayOffset = errors.New("cbor.invalidArrayOffset")

// ErrorInvalidPointer
var ErrorInvalidPointer = errors.New("cbor.invalidPointer")

// ErrorNoKey
var ErrorNoKey = errors.New("cbor.noKey")

// ErrorExpectedKey
var ErrorExpectedKey = errors.New("cbor.expectedKey")

// ErrorExpectedCborKey
var ErrorExpectedCborKey = errors.New("cbor.expectedCborKey")

// ErrorMalformedDocument
var ErrorMalformedDocument = errors.New("cbor.malformedDocument")

// ErrorInvalidDocument
var ErrorInvalidDocument = errors.New("cbor.invalidDocument")

// ErrorUnknownType to encode
var ErrorUnknownType = errors.New("cbor.unknownType")

const maxPartSize int = 1024

func fromJsonPointer(jsonptr, out []byte) int {
	var part [maxPartSize]byte

	n, off := encodeTextStart(out), 0
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
				n += encodeTag(uint64(tagJsonString), out[n:])
				n += encodeText(bytes2str(part[:off]), out[n:])
				off = 0
			}
			i++

		} else {
			part[off] = jsonptr[i]
			i, off = i+1, off+1
		}
	}
	if off > 0 || (len(jsonptr) > 0 && jsonptr[len(jsonptr)-1] == '/') {
		n += encodeTag(uint64(tagJsonString), out[n:])
		n += encodeText(bytes2str(part[:off]), out[n:])
	}

	n += encodeBreakStop(out[n:])
	return n
}

func toJsonPointer(cborptr, out []byte) int {
	i, n := 1, 0
	for {
		if cborptr[i] == hdr(type6, info24) && cborptr[i+1] == tagJsonString {
			i, out[n] = i+2, '/'
			n += 1
			ln, j := decodeLength(cborptr[i:])
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

func containerLen(doc []byte) (mjr byte, length int, size int, n int) {
	n, size = 0, -1
	if doc[0] == hdr(type6, info24) && doc[1] == tagSizePrefix {
		size, _ = decodeLength(doc[2:])
		n += 7
	}
	mjr, inf := major(doc[n]), info(doc[n])
	if mjr == type4 || mjr == type5 {
		if inf == indefiniteLength {
			return mjr, -1, size, n + 1
		}
		ln, m := decodeLength(doc[n:])
		return mjr, ln, size, n + m
	}
	panic(ErrorMalformedDocument)
}

func partial(part, doc []byte) (start, end int, key bool) {
	var err error
	var index int
	mjr, length, _ /*size*/, n := containerLen(doc)
	//fmt.Println("partial", string(part), len(doc), length, n, doc)
	if mjr == type4 { // array
		if index, err = strconv.Atoi(bytes2str(part)); err != nil {
			panic(ErrorInvalidArrayOffset)
		}
		n += arrayIndex(doc[n:], index, length)
		m := itemsEnd(doc[n:])
		//fmt.Println("partial-arr", index, n, n+m, doc[n:n+m], string(part))
		return n, n + m, false

	} else if mjr == type5 { // map
		m, found := mapIndex(doc[n:], part, length)
		if !found { // key not found
			return n + m, n + m, found
		}
		n += m
		m = itemsEnd(doc[n:])    // key
		p := itemsEnd(doc[n+m:]) // value
		//fmt.Println("partial-map",n,n+m,n+m+p,doc[n+m:n+m+p],string(part),found)
		return n, n + m + p, found
	}
	panic(ErrorInvalidPointer)
}

func lookup(cborptr, doc []byte) (start, end int, key bool) {
	i, n, m := 1, 0, len(doc)
	start, end = n, m
	if cborptr[i] == brkstp { // cborptr is empty ""
		return start, end, false
	}
	var k, keyln int
	for i < len(cborptr) && cborptr[i] != brkstp {
		doc = doc[n:m]
		if cborptr[i] == hdr(type6, info24) && cborptr[i+1] == tagJsonString {
			if key {
				start += 2 + k + keyln
			}
			i += 2
			ln, j := decodeLength(cborptr[i:])
			n, m, key = partial(cborptr[i+j:i+j+ln], doc)
			i += j + ln
			start += n
			end = start + (m - n)
			if key {
				keyln, k = decodeLength(doc[n+2:])
				n += 2 + k + keyln
			}
			continue
		}
		panic(ErrorInvalidPointer)
	}
	return start, end, key
}

func arrayIndex(arr []byte, index, length int) int {
	count, prev, n := 0, 0, 0
	for (length >= 0 && count < length) || arr[n] != brkstp {
		if count == index {
			return n
		} else if index >= 0 && (count == length || arr[n] == brkstp) {
			panic(ErrorInvalidArrayOffset)
		}
		prev = n
		n += itemsEnd(arr[n:])
		count++
	}
	if index == -1 && (count == length || arr[n] == brkstp) {
		return prev
	}
	panic(ErrorInvalidArrayOffset)
}

func mapIndex(buf []byte, part []byte, length int) (int, bool) {
	n := 0
	for n < len(buf) {
		start := n
		if length == 0 { // key not-found
			return n, false
		} else if buf[n] == brkstp { // key-not-found
			return n + 1, false
		}
		// get key
		if major(buf[n]) == type6 && buf[n+1] == tagJsonString {
			n += 2
		}
		if major(buf[n]) != type3 {
			panic(ErrorExpectedKey)
		}
		ln, j := decodeLength(buf[n:])
		n += j
		m := n + ln
		//fmt.Println("mapIndex", n, m, length, string(buf[n:m]), start, part)
		if bytes.Compare(part, buf[n:m]) == 0 {
			return start, true
		}
		p := itemsEnd(buf[m:]) // value
		//fmt.Println("mapIndex", n, m, p, string(buf[n:m]), start)
		n = m + p
		length--
	}
	panic(ErrorMalformedDocument)
}

func itemsEnd(buf []byte) int {
	mjr, inf := major(buf[0]), info(buf[0])
	if mjr == type0 || mjr == type1 { // integer item
		if inf < info24 {
			return 1
		}
		return (1 << (inf - info24)) + 1

	} else if mjr == type3 { // string item
		ln, j := decodeLength(buf)
		return j + ln

	} else if mjr == type6 && inf == info24 && buf[1] == tagSizePrefix {
		_, _, size, m := containerLen(buf)
		//fmt.Println("itemIndex-prefix", m, size)
		return m + size

	} else if mjr == type4 { // array item
		_, length, size, n := containerLen(buf)
		//fmt.Println("itemIndex-arr", length, size, n)
		if size > 0 {
			return n + size
		} else if length == 0 {
			return n
		} else if buf[n] == brkstp {
			return n + 1
		}
		n += arrayIndex(buf[n:], -1, length)
		if length < 0 {
			return n + itemsEnd(buf[n:]) + 1 // skip brkstp
		}
		return n + itemsEnd(buf[n:])

	} else if mjr == type5 { // map item
		_, length, size, n := containerLen(buf)
		//fmt.Println("itemIndex-map", length, size, n)
		if size > 0 {
			return n + size
		}
		for n < len(buf) {
			if length == 0 {
				return n
			} else if buf[n] == brkstp {
				return n + 1
			}
			n += itemsEnd(buf[n:]) // key
			n += itemsEnd(buf[n:]) // value
			length--
		}

	} else if mjr == type7 {
		if inf == simpleTypeNil || inf == simpleTypeFalse ||
			inf == simpleTypeTrue {
			return 1
		} else if inf == flt32 { // item float32
			return 1 + 4
		} else if inf == flt64 { // item float64
			return 1 + 8
		}
		panic(ErrorInvalidDocument)

	} else if mjr == type6 && buf[1] == tagJsonString && major(buf[2]) == type3 {
		ln, j := decodeLength(buf[2:])
		return 2 + j + ln
	}
	//fmt.Println(buf)
	panic(ErrorInvalidDocument)
}

func skipKey(doc []byte) int {
	ln, j := decodeLength(doc[2:])
	return 2 + j + ln
}

func get(doc, cborptr, item []byte) int {
	n, m, key := lookup(cborptr, doc)
	if n == m {
		panic(ErrorNoKey)
	} else if key { // if lookup in into a map, skip key
		n += skipKey(doc[n:])
	}
	copy(item, doc[n:m])
	return m - n
}

func set(doc, cborptr, item, newdoc, old []byte) (int, int) {
	n, m, key := lookup(cborptr, doc)
	if key {
		n += skipKey(doc[n:])
	}
	ln := len(item)
	copy(newdoc, doc[:n])
	copy(newdoc[n:], item)
	copy(newdoc[n+ln:], doc[m:])
	copy(old, doc[n:m])
	return (n + ln + len(doc[m:])), m - n
}

func prepend(doc, cborptr, item, newdoc []byte, config *Config) int {
	n, _, key := lookup(cborptr, doc)
	if key {
		n += skipKey(doc[n:])
	}
	// n now points to value which can be an array or map.
	mjr, length, size, _ := containerLen(doc[n:])
	//fmt.Println("prepend", mjr, length, size, n)
	if mjr != type4 && mjr != type5 {
		panic(ErrorInvalidPointer)
	}
	if size >= 0 {
		size += len(item)
	}
	if length >= 0 {
		length++
	}

	// copy every thing before value
	copy(newdoc, doc[:n])
	x, y := n, n
	if size >= 0 {
		x += encodeTag(tagSizePrefix, newdoc[x:])
		x += encodeLength(uint32(size), newdoc[x:])
		y += 7
	}
	if length >= 0 {
		p := x
		x += encodeLength(uint64(length), newdoc[p:])
		newdoc[p] = hdr(mjr, info(doc[y]))
	} else { // stream encoding
		newdoc[x] = doc[y]
		x++
		y++
	}
	ln := len(item)
	copy(newdoc[x:], item)
	copy(newdoc[x+ln:], doc[y:])
	return x + ln + len(doc[y:])
}

func getContainerPtr(cborptr, ccptr []byte) (prev int) {
	for n := 1; n < len(cborptr); {
		if cborptr[n] == brkstp {
			break
		}
		ln, m := decodeLength(cborptr[n+2:])
		prev, n = n, n+2+m+ln
	}
	copy(ccptr, cborptr[:prev])
	ccptr[prev] = brkstp
	return prev
}

func getLeafPtr(cborptr []byte, off int, ccptr []byte) int {
	a := encodeTextStart(ccptr)
	copy(ccptr[a:], cborptr[off:])
	return len(cborptr[off:]) + a
}

func del(doc, cborptr, newdoc, deleted []byte) (int, int) {
	if len(cborptr) < 2 || cborptr[1] == brkstp {
		panic(ErrorInvalidPointer)
	}

	var ccptr []byte
	if len(cborptr) < len(newdoc) {
		ccptr = newdoc
	} else {
		ccptr = make([]byte, len(cborptr))
	}

	// get container pointer
	off := getContainerPtr(cborptr, ccptr)
	n, m, key := lookup(ccptr, doc)
	if key {
		n += skipKey(doc[n:])
	}

	// get leaf pointer
	getLeafPtr(cborptr, off, ccptr)
	x, y, key := lookup(ccptr, doc[n:m])

	copy(newdoc, doc[:n]) // copy every thing till start of contianer

	// adjust size and length prefix
	mjr, length, size, v := containerLen(doc[n:m])
	if size >= 0 {
		size -= (y - x)
	}
	if length >= 0 {
		length--
	}

	// fix the size, if present, for the new container.
	p, q := n, n
	if size >= 0 {
		p += encodeTag(tagSizePrefix, newdoc[p:])
		p += encodeLength(uint32(size), newdoc[p:])
		q += 7 // p and q are same here.
	}
	newdoc[p] = doc[q] // assume stream encoding
	if length >= 0 {
		// may be not, but its okay.
		p += encodeLength(uint64(length), newdoc[p:])
		newdoc[q] = hdr(mjr, info(doc[q]))
		q += v - 7
	}

	copy(newdoc[p:], doc[q:n+x]) // copy siblings uptil deleted item
	p += len(doc[q : n+x])

	// skip the deleted value and copy the remaining.
	copy(newdoc[p:], doc[n+y:])
	docsz := p + len(doc[n+y:])

	// copy the deleted value
	if key {
		x += skipKey(doc[n+x:])
	}
	copy(deleted, doc[n+x:n+y])
	valsz := len(doc[n+x : n+y])
	return docsz, valsz
}
