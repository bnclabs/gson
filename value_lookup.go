//  Copyright (c) 2015 Couchbase, Inc.

package gson

import "strconv"

func docGet(segments []string, doc interface{}) interface{} {
	if len(segments) == 0 { // exit recursion.
		return doc
	}

	switch val := doc.(type) {
	case []interface{}:
		if segments[0] == "-" { // not required as per rfc-6901
			return docGet(segments[1:], val[len(val)-1])
		} else if idx, err := strconv.Atoi(segments[0]); err != nil {
			panic("docGet() gson pointer-invalidIndex")
		} else if idx >= len(val) {
			panic("docGet() gson pointer-index-outofRange")
		} else {
			return docGet(segments[1:], val[idx])
		}

	case map[string]interface{}:
		if doc, ok := val[segments[0]]; !ok {
			panic("docGet() gson pointer-invalidKey")
		} else {
			return docGet(segments[1:], doc)
		}
	}
	panic("docGet() gson invalidPointer")
}

func docSet(segments []string, doc, item interface{}) (newdoc, old interface{}) {
	ln, container := len(segments), doc
	if ln == 0 {
		panic("docSet() document is not a container")
	} else if ln > 1 {
		container = docGet(segments[:ln-1], doc)
	} // else if ln == 1, container _is_ doc

	key := segments[ln-1]

	var ok bool
	switch cont := container.(type) {
	case []interface{}:
		if key == "-" {
			old = item
			cont = append(cont, item)
			if ln > 1 {
				docSet(segments[:ln-1], doc, cont)
			} else { // edge case !
				return cont, item
			}
		} else if idx, err := strconv.Atoi(key); err != nil {
			panic("docSet() gson pointer-invalidIndex")
		} else if idx >= len(cont) {
			panic("docSet() gson pointer-outofRange")
		} else {
			old, cont[idx] = cont[idx], item
		}

	case map[string]interface{}:
		if old, ok = cont[key]; !ok {
			old = item
		}
		cont[key] = item
	default:
		panic("docSet() gson invalidPointer")
	}
	return doc, old
}

func docDel(segments []string, doc interface{}) (newdoc, old interface{}) {
	ln, container := len(segments), doc
	if ln == 0 {
		panic("docDel() document is not a container")
	} else if ln > 1 {
		container = docGet(segments[:ln-1], doc)
	} // else if ln == 1, container _is_ doc

	key := segments[ln-1]

	switch cont := container.(type) {
	case []interface{}:
		if idx, err := strconv.Atoi(key); err != nil {
			panic("docDel() gson pointer-invalidIndex")
		} else if idx >= len(cont) {
			panic("docDel() gson pointer-outofRange")
		} else {
			old = cont[idx]
			copy(cont[idx:], cont[idx+1:])
			cont = cont[:len(cont)-1]
			if ln > 1 {
				docSet(segments[:ln-1], doc, cont)
			} else { // edge case !!
				return cont, old
			}
		}

	case map[string]interface{}:
		old, _ = cont[key]
		delete(cont, key)

	default:
		panic("docDel() gson invalidPointer")
	}
	return doc, old
}
