//  Copyright (c) 2015 Couchbase, Inc.

package gson

import "strconv"
import "bytes"

import "fmt"

func cborGet(doc []byte, segments [][]byte, item []byte) int {
	_, _, start, end := cborLookup(doc, segments)
	if start < 0 {
		key := bytes2str(segments[len(segments)-1])
		panic(fmt.Sprintf("key %v not found", key))
	}
	return copy(item, doc[start:end])
}

func cborSet(doc []byte, segments [][]byte, item, newdoc, olditem []byte) (int, int) {
	var x int

	cont, _, start, end := cborLookup(doc, segments)

	m := copy(newdoc, doc[:cont])

	if start < 0 {
		x, m = addlength(m, cont, doc, newdoc)
		m += copy(newdoc[m:], segments[len(segments)-1])
		m += copy(newdoc[m:], item)
		m += copy(newdoc[m:], doc[x:])
		return m, 0
	}
	m += copy(newdoc[m:], doc[cont:start])
	m += copy(newdoc[m:], item)
	m += copy(newdoc[m:], doc[end:])
	n := copy(olditem, doc[start:end])
	return m, n
}

func cborDel(doc []byte, segments [][]byte, newdoc, deleted []byte) (int, int) {
	var x int

	cont, keyn, start, end := cborLookup(doc, segments)

	m := copy(newdoc, doc[:cont])

	major, _ := cborMajor(doc[cont]), cborInfo(doc[cont])
	switch major {
	case cborType4:
		if keyn >= 0 {
			panic("cborType4 expected keyn to be -1")
		}
		x, m = addlength(m, cont, doc, newdoc)
		m += copy(newdoc[m:], doc[x:start])
		m += copy(newdoc[m:], doc[end:])
		n := copy(deleted, doc[start:end])
		return m, n

	case cborType5:
		if keyn < 0 {
			panic("cborType5 expected keyn to be > 0")
		}
		x, m = addlength(m, cont, doc, newdoc)
		m += copy(newdoc[m:], doc[x:keyn])
		m += copy(newdoc[m:], doc[end:])
		n := copy(deleted, doc[start:end])
		return m, n
	}
	panic("unreachable code")
}

func cborPrepend(doc []byte, segments [][]byte, item, newdoc []byte) int {
	var x int

	_, _, start, _ := cborLookup(doc, segments)
	major, _ := cborMajor(doc[start]), cborInfo(doc[start])
	if major != cborType4 {
		panic("cannot prepend to non array containers")
	}
	m := copy(newdoc, doc[:start])
	x, m = addlength(m, start, doc, newdoc)
	m += copy(newdoc[m:], item)
	m += copy(newdoc[m:], doc[x:])
	return m
}

func cborAppend(doc []byte, segments [][]byte, item, newdoc []byte) int {
	var x int

	_, _, start, end := cborLookup(doc, segments)
	major, info := cborMajor(doc[start]), cborInfo(doc[start])
	if major != cborType4 {
		panic("cannot append to non array containers")
	}
	m := copy(newdoc, doc[:start])
	x, m = addlength(m, start, doc, newdoc)
	if info == cborIndefiniteLength {
		m += copy(newdoc[m:], doc[x:end-1])
		m += copy(newdoc[m:], item)
		newdoc[m] = doc[end]
		m++
	} else {
		m += copy(newdoc[m:], doc[x:end])
		m += copy(newdoc[m:], item)
	}
	return m
}

func cborLookup(doc []byte, segments [][]byte) (cont, keyn, start, end int) {
	var ln int

nextseg:
	for i, segment := range segments {
		major, info := cborMajor(doc[start]), cborInfo(doc[start])
		switch major {
		case cborType4:
			idx, count := segment2idx(segment), 0
			cont, keyn = start, -1
			if info == cborIndefiniteLength {
				for end = start; doc[end] != brkstp; count++ {
					_, n := cborItem(doc[end:])
					start, end = end, end+n
					if count == idx {
						continue nextseg
					}
				}
				panic(fmt.Sprintf("index %v overflow", idx))
			}
			ln, end = cborItemLength(doc)
			for ; count < ln; count++ {
				_, n := cborItem(doc[end:])
				start, end = end, end+n
				if count == idx {
					continue nextseg
				}
			}
			panic(fmt.Sprintf("index %v overflow", idx))

		case cborType5:
			cont = start
			if info == cborIndefiniteLength {
				for end = start; doc[end] != brkstp; {
					_, m := cborItem(doc[end:])
					_, n := cborItem(doc[end+m:])
					keyn, start, end = end, end+m, end+m+n
					if bytes.Compare(doc[keyn:start], segment) == 0 {
						continue nextseg
					}
				}
				if i == (len(segments) - 1) { // leaf
					return cont, -1, -1, -1
				}
				panic(fmt.Sprintf("key %v not found", bytes2str(segment)))
			}
			ln, end = cborItemLength(doc)
			for i := 0; i < ln; i++ {
				_, m := cborItem(doc[end:])
				_, n := cborItem(doc[end+m:])
				keyn, start, end = end, end+m, end+m+n
				if bytes.Compare(doc[keyn:start], segment) == 0 {
					continue nextseg
				}
			}
			if i == (len(segments) - 1) { // leaf
				return cont, -1, -1, -1
			}
			panic(fmt.Sprintf("key %v not found", bytes2str(segment)))
		}
	}
	return
}

func segment2idx(segment []byte) int {
	idx, err := strconv.Atoi(bytes2str(segment))
	if err != nil {
		fmsg := "pointer %v expected to be array index"
		panic(fmt.Sprintf(fmsg, bytes2str(segment)))
	} else if idx < 0 {
		panic(fmt.Sprintf("array index %v can be < 0", idx))
	}
	return idx
}

func addlength(m, cont int, doc, newdoc []byte) (int, int) {
	var x, ln int

	major, info := cborMajor(doc[cont]), cborInfo(doc[cont])

	if info == cborIndefiniteLength {
		newdoc[m] = doc[cont]
		m += 1
	} else {
		ln, x = cborItemLength(doc[cont:])
		y := valuint642cbor(uint64(ln-1), newdoc[m:])
		newdoc[m] = (newdoc[m] & 0x1f) | major // fix the type from type0->type4
		m += y
	}
	return x, m
}
