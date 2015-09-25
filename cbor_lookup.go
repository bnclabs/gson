//  Copyright (c) 2015 Couchbase, Inc.

package gson

import "strconv"
import "bytes"

//import "fmt"

func cborContainerLen(doc []byte) (mjr byte, n int) {
	n = 0
	mjr, inf := cborMajor(doc[n]), cborInfo(doc[n])
	if mjr == cborType4 || mjr == cborType5 {
		if inf == cborIndefiniteLength {
			return mjr, 1
		}
		panic("cbor pointer len-prefix not supported")
	}
	panic("cbor pointer lookup malformedDocument")
}

func cborPartial(part, doc []byte) (start, end int, key bool) {
	var err error
	var index int
	mjr, n := cborContainerLen(doc)
	//fmt.Println("partial", string(part), part, len(doc), n, doc, mjr)
	if mjr == cborType4 { // array
		if index, err = strconv.Atoi(bytes2str(part)); err != nil {
			panic("cbor pointer segment lookup invalidArrayOffset")
		}
		n += cborArrayIndex(doc[n:], index)
		m := cborItemsEnd(doc[n:])
		//fmt.Println("partial-arr", index, n, n+m, doc[n:n+m], string(part))
		return n, n + m, false

	} else if mjr == cborType5 { // map
		m, found := cborMapIndex(doc[n:], part)
		if !found { // key not found
			return n + m, n + m, found
		}
		n += m
		m = cborItemsEnd(doc[n:])    // key
		p := cborItemsEnd(doc[n+m:]) // value
		//fmt.Println("partial-map", n, n+m, n+m+p, doc[n+m:n+m+p], string(part), found)
		return n, n + m + p, found
	}
	panic("cbor pointer segment lookup invalidPointer")
}

func cborLookup(cborptr, doc []byte) (start, end int, key bool) {
	i, n, m := 1, 0, len(doc)
	start, end = n, m
	if i >= len(cborptr) || cborptr[i] == brkstp { // cborptr is empty ""
		return start, end, false
	}
	var k, keyln int
	byt := cborHdr(cborType6, cborInfo24)
	for {
		doc = doc[n:m]
		if cborptr[i] != byt && cborptr[i+1] != tagJsonString {
			panic("cbor pointer lookup invalidPointer")
		}
		i += 2
		ln, j := cborItemLength(cborptr[i:])
		n, m, key = cborPartial(cborptr[i+j:i+j+ln], doc)
		i += j + ln
		start += n
		end = start + (m - n)
		//fmt.Println("lookup", i, cborptr[i] == brkstp, start, n, m, k, keyln)
		if i >= len(cborptr) || cborptr[i] == brkstp {
			break
		}
		if key {
			keyln, k = cborItemLength(doc[n:])
			n, start = n+k+keyln, start+k+keyln
		}
	}
	return start, end, key
}

func cborArrayIndex(arr []byte, index int) int {
	count, prev, n := 0, 0, 0
	for arr[n] != brkstp {
		if count == index {
			return n
		} else if index >= 0 && arr[n] == brkstp {
			panic("cbor pointer array index invalidArrayOffset")
		}
		prev = n
		n += cborItemsEnd(arr[n:])
		count++
	}
	if index == -1 && arr[n] == brkstp {
		return prev
	}
	panic("cbor pointer array index ivalidArrayOffset")
}

func cborMapIndex(buf []byte, part []byte) (int, bool) {
	n := 0
	for n < len(buf) {
		start := n
		if buf[n] == brkstp { // key-not-found
			return n + 1, false
		}
		// get key
		if cborMajor(buf[n]) != cborType3 {
			panic("cbor pointer map index expectedKey")
		}
		ln, j := cborItemLength(buf[n:])
		n += j
		m := n + ln
		//fmt.Println("mapIndex-", n, m, string(buf[n:m]), start, part)
		if bytes.Compare(part, buf[n:m]) == 0 {
			return start, true
		}
		p := cborItemsEnd(buf[m:]) // value
		//fmt.Println("mapIndex", n, m, p, string(buf[n:m]), start)
		n = m + p
	}
	panic("cbor pointer map index malformedDocument")
}

func cborItemsEnd(buf []byte) int {
	mjr, inf := cborMajor(buf[0]), cborInfo(buf[0])
	if mjr == cborType0 || mjr == cborType1 { // integer item
		if inf < cborInfo24 {
			return 1
		}
		return (1 << (inf - cborInfo24)) + 1

	} else if mjr == cborType3 { // string item
		ln, j := cborItemLength(buf)
		return j + ln

	} else if mjr == cborType4 { // array item
		_, n := cborContainerLen(buf)
		//fmt.Println("itemIndex-arr", n, buf[n] == brkstp)
		if buf[n] == brkstp {
			return n + 1
		}
		n += cborArrayIndex(buf[n:], -1)
		return n + cborItemsEnd(buf[n:]) + 1 // skip brkstp

	} else if mjr == cborType5 { // map item
		_, n := cborContainerLen(buf)
		//fmt.Println("itemIndex-map", n)
		for n < len(buf) {
			if buf[n] == brkstp {
				return n + 1
			}
			n += cborItemsEnd(buf[n:]) // key
			n += cborItemsEnd(buf[n:]) // value
		}

	} else if mjr == cborType7 {
		if inf == cborSimpleTypeNil || inf == cborSimpleTypeFalse ||
			inf == cborSimpleTypeTrue {
			return 1
		} else if inf == cborFlt32 { // item float32
			return 1 + 4
		} else if inf == cborFlt64 { // item float64
			return 1 + 8
		}
		panic("cbor pointer lookup invalidDocument")
	}
	panic("cbor pointer lookup invalidDocument")
}

func cborSkipkey(doc []byte) int {
	ln, j := cborItemLength(doc)
	return j + ln
}

func cborGet(doc, cborptr, item []byte) int {
	n, m, key := cborLookup(cborptr, doc)
	if n == m {
		panic("cbor pointer get noKey")
	} else if key { // if lookup in into a map, skip key
		n += cborSkipkey(doc[n:])
	}
	copy(item, doc[n:m])
	return m - n
}

func cborSet(doc, cborptr, item, newdoc, old []byte) (int, int) {
	n, m, key := cborLookup(cborptr, doc)
	if key {
		n += cborSkipkey(doc[n:])
	}
	ln := len(item)
	copy(newdoc, doc[:n])
	copy(newdoc[n:], item)
	copy(newdoc[n+ln:], doc[m:])
	copy(old, doc[n:m])
	return (n + ln + len(doc[m:])), m - n
}

func cborPrepend(doc, cborptr, item, newdoc []byte, config *Config) int {
	n, _, key := cborLookup(cborptr, doc)
	//fmt.Println(n, key)
	if key { // n points to {key,value} pair
		n += cborSkipkey(doc[n:])
	}
	// n now points to value which can be an array or map.
	mjr := cborMajor(doc[n])
	if mjr != cborType4 && mjr != cborType5 {
		panic("cbor pointer prepend invalidPointer")
	}
	// copy every thing before value
	n++ // including mjr+indefiniteLength
	copy(newdoc, doc[:n])
	ln := len(item)
	copy(newdoc[n:], item)
	copy(newdoc[n+ln:], doc[n:])
	return n + ln + len(doc[n:])
}

func cborDel(doc, cborptr, newdoc, deleted []byte) (int, int) {
	n, m, key := cborLookup(cborptr, doc)
	copy(newdoc, doc[:n])
	copy(newdoc[n:], doc[m:])
	// copy deleted value to o/p buffer.
	p := n
	if key {
		p += cborSkipkey(doc[n:])
	}
	copy(deleted, doc[p:m])
	return n + len(doc[m:]), m - p
}
