//  Copyright (c) 2015 Couchbase, Inc.

package gson

// Value abstractions for golang-native value.
type Value struct {
	config *Config
	data   interface{}
}

// Tojson encode golang native value to json text.
func (val *Value) Tojson(jsn *Json) *Json {
	jsn.n += value2json(val.data, jsn.data[jsn.n:], val.config)
	return jsn
}

// Tocbor encode golang native into cbor binary.
func (val *Value) Tocbor(cbr *Cbor) *Cbor {
	cbr.n += value2cbor(val.data, cbr.data[cbr.n:], val.config)
	return cbr
}

// Tocollate encode golang native into binary-collation.
func (val *Value) Tocollate(clt *Collate) *Collate {
	clt.n += gson2collate(val.data, clt.data[clt.n:], val.config)
	return clt
}

// ListPointers all possible pointers in value.
func (val *Value) ListPointers(ptrs []string) []string {
	prefix := val.config.pools.prefixPool.Get().([]byte)
	defer val.config.pools.prefixPool.Put(prefix[:0])
	ptrs = allpaths(val.data, ptrs, prefix)
	ptrs = append(ptrs, "")
	return ptrs
}

// Get field or nested field specified by json pointer.
func (val *Value) Get(jptr *Jsonpointer) (item interface{}) {
	return valGet(jptr.Segments(), val.data)
}

// Set field or nested field specified by json pointer. While
// `newval` is gauranteed to contain the `item`, `val` _may_ not be.
// Suggested usage,
//      val := config.NewValue([]interface{}{"hello"})
//      newval, _ = val.Set("/-", "world")
func (val *Value) Set(jptr *Jsonpointer, item interface{}) (newval, oldval interface{}) {
	return valSet(jptr.Segments(), val.data, item)
}

// Delete field or nested field specified by json pointer. While
// `newval` is gauranteed to be updated, `val` _may_ not be.
// Suggested usage,
//      val := NewValue([]interface{}{"hello", "world"})
//      newval, _ = val.Delete("/1")
func (val *Value) Delete(jptr *Jsonpointer) (newval, deleted interface{}) {
	return valDel(jptr.Segments(), val.data)
}

// Append item to end of an array pointed by json-pointer.
// returns `newval`, is gauranteed to be updated,
//      val := NewValue([]interface{}{"hello", "world"})
//      newval, _ = val.Append("", "welcome")
func (val *Value) Append(jptr *Jsonpointer, item interface{}) interface{} {
	return valAppend(jptr.Segments(), val.data, item)
}

// Prepend an item to the beginning of an array.
// returns `newval`, is gauranteed to be updated,
//      val := NewValue([]interface{}{"hello", "world"})
//      newval, _ = val.Append("", "welcome")
func (val *Value) Prepend(jptr *Jsonpointer, item interface{}) interface{} {
	return valPrepend(jptr.Segments(), val.data, item)
}
