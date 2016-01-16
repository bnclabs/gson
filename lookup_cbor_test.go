//  Copyright (c) 2015 Couchbase, Inc.

// +build ignore

package gson

import "testing"
import "fmt"
import "bytes"
import "encoding/json"
import "reflect"

var _ = fmt.Sprintf("dummy")

func TestCborTypicalPointers(t *testing.T) {
	config := NewDefaultConfig()
	cborptr := make([]byte, 1024)
	cbordoc := make([]byte, 1024*1024)
	item := make([]byte, 10*1024)
	pointers := make([]string, 0, 1024)

	txt := string(testdataFile("testdata/typical.json"))
	_, doc := config.JsonToValue(txt)
	pointers = config.ListPointers(doc, pointers)
	_, n := config.JsonToCbor(txt, cbordoc)
	cbordoc = cbordoc[:n]
	for _, ptr := range pointers {
		if ln := len(ptr); ln > 0 && ptr[ln-1] == '-' {
			continue
		}
		t.Logf("pointer %v", ptr)
		ref := config.DocGet(ptr, doc)
		n := config.JsonPointerToCbor([]byte(ptr), cborptr)
		t.Logf("%v", cbordoc)
		n = config.CborGet(cbordoc, cborptr[:n], item)
		val, _ := config.CborToValue(item[:n])
		if !reflect.DeepEqual(CborMap2golangMap(val), ref) {
			fmsg := "expected {%T,%v} for ptr %q, got {%T,%v}"
			t.Fatalf(fmsg, ref, ref, ptr, val, val)
		}
	}
}

func TestCborGet(t *testing.T) {
	txt := `{"a": 10, "arr": [1,2], "dict": {"a":10, "b":20}}`
	testcases := [][2]interface{}{
		//[2]interface{}{"/a", 10.0},
		[2]interface{}{"/arr/0", 1.0},
		[2]interface{}{"/arr/1", 2.0},
		[2]interface{}{"/dict/a", 10.0},
		[2]interface{}{"/dict/b", 20.0},
	}

	cbordoc, cborptr := make([]byte, 1024), make([]byte, 1024)
	item := make([]byte, 1024)
	config := NewDefaultConfig()
	_, n := config.JsonToCbor(txt, cbordoc)
	cbordoc = cbordoc[:n]
	for _, tcase := range testcases {
		ptr := tcase[0].(string)
		ref := tcase[1]
		t.Logf("%v", ptr)
		n := config.JsonPointerToCbor([]byte(ptr), cborptr)
		n = config.CborGet(cbordoc, cborptr[:n], item)
		val, _ := config.CborToValue(item[:n]) // cbor->value
		if !reflect.DeepEqual(CborMap2golangMap(val), ref) {
			fmsg := "expected {%T,%v}, for ptr %v, got {%T,%v}"
			t.Fatalf(fmsg, ref, ref, ptr, val, val)
		}
	}
}

func TestCborSet(t *testing.T) {
	txt := `{"a": 10, "arr": [1,2], "dict": {"a":10, "b":20}}`
	testcases := [][4]interface{}{
		[4]interface{}{
			"/a", 11.0, `10`, `{"a":11,"arr":[1,2],"dict":{"a":10,"b":20}}`},
		[4]interface{}{
			"/arr/0", 10.0, `1`, `{"a":11,"arr":[10,2],"dict":{"a":10,"b":20}}`},
		[4]interface{}{
			"/arr/1", 20.0, `2`, `{"a":11,"arr":[10,20],"dict":{"a":10,"b":20}}`},
		[4]interface{}{
			"/dict/a", 1.0, `10`, `{"a":11,"arr":[10,20],"dict":{"a":1,"b":20}}`},
		[4]interface{}{
			"/dict/b", 2.0, `20`, `{"a":11,"arr":[10,20],"dict":{"a":1,"b":2}}`},
	}

	// cbor initialization
	cbordoc, cbordocnew := make([]byte, 1024), make([]byte, 1024)
	cborptr := make([]byte, 1024)
	item, itemold := make([]byte, 1024), make([]byte, 1024)
	config := NewDefaultConfig()
	_, n := config.JsonToCbor(txt, cbordoc)
	cbordoc = cbordoc[:n]

	for _, tcase := range testcases {
		ptr := tcase[0].(string)
		t.Logf("%v", ptr)
		n := config.ValueToCbor(tcase[1], item)
		item = item[:n]

		config.JsonPointerToCbor([]byte(ptr), cborptr)
		n, m := config.CborSet(
			cbordoc, cborptr /*[:n]*/, item, cbordocnew, itemold)
		copy(cbordoc, cbordocnew[:n])
		cbordoc = cbordoc[:n]

		val, _ := config.CborToValue(cbordocnew[:n])
		var vref, iref interface{}
		if err := json.Unmarshal([]byte(tcase[3].(string)), &vref); err != nil {
			t.Fatalf("parsing json: %v", err)
		} else if v := CborMap2golangMap(val); !reflect.DeepEqual(v, vref) {
			t.Fatalf("for %v expected %v, got %v", ptr, vref, v)
		}
		oval, _ := config.CborToValue(itemold[:m])
		if err := json.Unmarshal([]byte(tcase[2].(string)), &iref); err != nil {
			t.Fatalf("parsing json: %v", err)
		} else if ov := CborMap2golangMap(oval); !reflect.DeepEqual(ov, iref) {
			t.Fatalf("for %v item expected %v, got %v", ptr, iref, ov)
		}
	}

	// cbor set with empty pointer
	cbordoc, cbordocnew = make([]byte, 1024), make([]byte, 1024)
	cborptr = make([]byte, 1024)
	item, itemold = make([]byte, 1024), make([]byte, 1024)
	n = config.JsonPointerToCbor([]byte(""), cborptr)
	cborptr = cborptr[:n]
	_, n = config.JsonToCbor(txt, cbordoc)
	cbordoc = cbordoc[:n]
	txt = `{"a": 10}`
	_, n = config.JsonToCbor(txt, item)
	item = item[:n]
	n, m := config.CborSet(
		cbordoc, cborptr /*[:n]*/, item, cbordocnew, itemold)
	if bytes.Compare(cbordocnew[:n], item) != 0 {
		t.Errorf("expected %v, got %v", item, cbordocnew[:n])
	} else if bytes.Compare(itemold[:m], cbordoc) != 0 {
		t.Errorf("expected %v, got %v", cbordoc, itemold[:m])
	}
}

func TestCborPrepend(t *testing.T) {
	txt := `{"a": 10, "arr": [1,2], "dict": {"a":10, "b":20}}`
	ftxt := `{"a": 10, "b": 20, "arr": [1,2,3],"dict":{"a":10,"b":20,"c":30}}`

	// cbor initialization
	cbordoc, cbordocnew := make([]byte, 1024), make([]byte, 1024)
	cborptr, item := make([]byte, 1024), make([]byte, 1024)

	config := NewDefaultConfig()
	_, n := config.JsonToCbor(txt, cbordoc)

	// prepend "/", {"b": 20}
	t.Logf(`prepend "/", {"b": 20}`)
	i := config.JsonPointerToCbor([]byte(""), cborptr)
	m := config.MapsliceToCbor(
		[][2]interface{}{[2]interface{}{"b", 20}}, item)
	n = config.CborPrepend(cbordoc[:n], cborptr[:i], item[:m], cbordocnew)
	copy(cbordoc, cbordocnew[:n])

	// prepend "/arr" 3.0
	t.Logf(`prepend "/arr" 3.0`)
	config.JsonPointerToCbor([]byte("/arr"), cborptr)
	m = config.ValueToCbor(float64(3.0), item)
	n = config.CborPrepend(cbordoc[:n], cborptr, item[:m], cbordocnew)
	copy(cbordoc, cbordocnew[:n])

	// prepend "/dict/c" 30.0
	t.Logf(`prepend "/dict/c" 30`)
	config.JsonPointerToCbor([]byte("/dict"), cborptr)
	m = config.MapsliceToCbor(
		[][2]interface{}{[2]interface{}{"c", 30}}, item)
	n = config.CborPrepend(cbordoc[:n], cborptr, item[:m], cbordocnew)
	copy(cbordoc, cbordocnew[:n])

	val, _ := config.CborToValue(cbordoc[:n])
	if err := json.Unmarshal([]byte(ftxt), &val); err != nil {
		t.Fatalf("parsing json: %v", err)
	} else if v := CborMap2golangMap(val); !reflect.DeepEqual(v, val) {
		t.Fatalf("finally exptected %v, got %v", val, v)
	}

	// parent doc as an array
	t.Logf(`prepend "" to [1,2]`)
	txt = `[1,2]`
	ftxt = `[1,2,3]`
	_, n = config.JsonToCbor(txt, cbordoc)
	config.JsonPointerToCbor([]byte(""), cborptr)
	m = config.ValueToCbor(float64(3.0), item)
	n = config.CborPrepend(cbordoc[:n], cborptr, item[:m], cbordocnew)
	copy(cbordoc, cbordocnew[:n])

	val, _ = config.CborToValue(cbordoc[:n])
	if err := json.Unmarshal([]byte(ftxt), &val); err != nil {
		t.Fatalf("parsing json: %v", err)
	} else if v := CborMap2golangMap(val); !reflect.DeepEqual(v, val) {
		t.Fatalf("finally exptected %v, got %v", val, v)
	}
}

func TestCborDel(t *testing.T) {
	txt := `{"a": "10", "arr": [1,2], "dict": {"a":10, "b":20}}`
	testcases := [][3]interface{}{
		[3]interface{}{
			"/a", `"10"`, `{"arr": [1,2], "dict": {"a":10, "b":20}}`},
		[3]interface{}{
			"/arr/1", `2`, `{"arr": [1], "dict": {"a":10, "b":20}}`},
		[3]interface{}{
			"/dict/a", `10`, `{"arr": [1], "dict": {"b":20}}`},
		[3]interface{}{
			"/dict", `{"b":20}`, `{"arr": [1]}`},
		[3]interface{}{
			"/arr", `[1]`, `{}`},
	}

	// cbor initialization
	cbordoc, cbordocnew := make([]byte, 1024), make([]byte, 1024)
	cborptr, itemold := make([]byte, 1024), make([]byte, 1024)
	config := NewDefaultConfig()
	_, n := config.JsonToCbor(txt, cbordoc)
	cbordoc = cbordoc[:n]

	for _, tcase := range testcases {
		ptr := tcase[0].(string)
		t.Logf("%v", ptr)

		x := config.JsonPointerToCbor([]byte(ptr), cborptr)
		n, m := config.CborDelete(cbordoc, cborptr[:x], cbordocnew, itemold)
		copy(cbordoc, cbordocnew[:n])
		cbordoc = cbordoc[:n]

		val, _ := config.CborToValue(cbordocnew[:n])
		var vref, iref interface{}
		if err := json.Unmarshal([]byte(tcase[2].(string)), &vref); err != nil {
			t.Fatalf("parsing json: %v", err)
		} else if v := CborMap2golangMap(val); !reflect.DeepEqual(v, vref) {
			t.Fatalf("for %v expected %v, got %v", ptr, vref, v)
		}
		oval, _ := config.CborToValue(itemold[:m])
		if err := json.Unmarshal([]byte(tcase[1].(string)), &iref); err != nil {
			t.Fatalf("parsing json: %v", err)
		} else if ov := CborMap2golangMap(oval); !reflect.DeepEqual(ov, iref) {
			t.Fatalf("for %v item expected %v, got %v", ptr, iref, ov)
		}
	}
}

func TestCborLookups(t *testing.T) {
	cbordoc, item := make([]byte, 1024), make([]byte, 1024)
	cborptr := make([]byte, 1024)
	// panic on invalid document
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		config := NewDefaultConfig()
		_, n := config.JsonToCbor("10", cbordoc)
		p := config.JsonPointerToCbor([]byte("/a"), cborptr)
		config.CborGet(cbordoc[:n], cborptr[:p], item)
	}()
	// panic on cbor pointer len-prefix
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		config := NewDefaultConfig().ContainerEncoding(LengthPrefix)
		_, n := config.JsonToCbor("[1,2]", cbordoc)
		config.JsonPointerToCbor([]byte("/0"), cborptr)
		config.CborGet(cbordoc, cborptr[:n], item)
	}()
	// panic on invalid array offset
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		config := NewDefaultConfig()
		_, n := config.JsonToCbor("[1,2]", cbordoc)
		config.JsonPointerToCbor([]byte("/2"), cborptr)
		config.CborGet(cbordoc, cborptr[:n], item)
	}()
	// panic key not found
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		config := NewDefaultConfig()
		_, n := config.JsonToCbor(`{"1": 10, "2": 20}`, cbordoc)
		m := config.JsonPointerToCbor([]byte("/3"), cborptr)
		config.CborGet(cbordoc[:n], cborptr[:m], item)
	}()
	// panic invalid pointer
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		config := NewDefaultConfig()
		_, n := config.JsonToCbor(`{"1": 10, "2": 20}`, cbordoc)
		m := config.JsonPointerToCbor([]byte("/1/2"), cborptr)
		config.CborGet(cbordoc[:n], cborptr[:m], item)
	}()
}

func BenchmarkCborGet(b *testing.B) {
	config := NewDefaultConfig()
	txt := string(testdataFile("testdata/typical.json"))

	cbordoc := make([]byte, 10*1024)
	_, n := config.JsonToCbor(txt, cbordoc)
	cborptr := make([]byte, 10*1024)
	m := config.JsonPointerToCbor([]byte("/projects/Sherri/members/0"), cborptr)
	item := make([]byte, 10*1024)

	var p int
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p = config.CborGet(cbordoc[:n], cborptr[:m], item)
	}
	b.SetBytes(int64(p))
}

func BenchmarkCborSet(b *testing.B) {
	config := NewDefaultConfig().NumberKind(FloatNumber).SpaceKind(UnicodeSpace)
	txt := string(testdataFile("testdata/typical.json"))

	cbordoc := make([]byte, 10*1024)
	_, n := config.JsonToCbor(txt, cbordoc)

	cborptr := make([]byte, 10*1024)
	m := config.JsonPointerToCbor([]byte("/projects/Sherri/members/0"), cborptr)

	item := make([]byte, 10*1024)
	itemref := 10
	p := config.ValueToCbor(itemref, item)

	newdoc := make([]byte, 10*1024)
	old := make([]byte, 10*1024)

	var x, y int
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x, y = config.CborSet(cbordoc[:n], cborptr[:m], item[:p], newdoc, old)
	}
	b.SetBytes(int64(x + y))
}

func BenchmarkCborPrepend(b *testing.B) {
	config := NewDefaultConfig()
	txt := string(testdataFile("testdata/typical.json"))

	cbordoc := make([]byte, 10*1024)
	_, n := config.JsonToCbor(txt, cbordoc)
	cborptr1, cborptr2 := make([]byte, 10*1024), make([]byte, 10*1024)
	p := config.JsonPointerToCbor([]byte("/projects/Sherri/members"), cborptr1)
	q := config.JsonPointerToCbor([]byte("/projects/Sherri/members/0"), cborptr2)
	item := make([]byte, 10*1024)
	refitem := 10.0
	m := config.ValueToCbor(refitem, item)

	newdoc := make([]byte, 10*1024)

	var x, z int
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x = config.CborPrepend(cbordoc[:n], cborptr1[:p], item[:m], newdoc)
		n, m = config.CborDelete(newdoc[:x], cborptr2[:q], cbordoc, item)
	}
	if val, _ := config.CborToValue(item); !reflect.DeepEqual(val, refitem) {
		b.Fatalf("exptected %v, got %v", refitem, val)
	}
	b.SetBytes(int64(x + z))
}
