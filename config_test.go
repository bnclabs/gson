//  Copyright (c) 2015 Couchbase, Inc.

package gson

import "testing"
import "bytes"
import "reflect"
import "encoding/json"
import "fmt"

var _ = fmt.Sprintf("dummy")

func TestConfig(t *testing.T) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(make([]byte, 16), 0)
	val := config.NewValue(10.2)
	val.Tocbor(cbr)
	if value := cbr.Tovalue(); !reflect.DeepEqual(val.data, value) {
		t.Errorf("expected %v got %v", val.data, value)
	}
}

type testLocal byte

func TestUndefined(t *testing.T) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(make([]byte, 16), 0)
	val := config.NewValue(CborUndefined(cborSimpleUndefined))
	val.Tocbor(cbr)
	if value := cbr.Tovalue(); !reflect.DeepEqual(val.data, value) {
		t.Errorf("expected %v got %v", val.data, value)
	}
	// test unknown type.
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		config.NewValue(testLocal(10)).Tocbor(cbr.Reset(nil))
	}()
}

func TestJsonToValue(t *testing.T) {
	config := NewDefaultConfig().SpaceKind(AnsiSpace)
	jsn := config.NewJson([]byte(`"abcd"  "xyz" "10" `), -1)
	jsnrmn, value := jsn.Tovalue()
	if string(jsnrmn.Bytes()) != `  "xyz" "10" ` {
		t.Errorf("expected %q, got %q", `  "xyz" "10" `, string(jsnrmn.Bytes()))
	}

	jsnback := config.NewJson(make([]byte, 1024), 0)
	config.NewValue(value).Tojson(jsnback)
	if ref := `"abcd"`; string(jsnback.Bytes()) != ref {
		t.Errorf("expected %v, got %v", ref, string(jsnback.Bytes()))
	}
}

func TestJsonToValues(t *testing.T) {
	var s string
	uni_s := `"汉语 / 漢語; Hàn\b \t\uef24yǔ "`
	json.Unmarshal([]byte(uni_s), &s)
	ref := []interface{}{"abcd", "xyz", "10", s}

	config := NewDefaultConfig().SpaceKind(AnsiSpace)
	jsn := config.NewJson([]byte(`"abcd"  "xyz" "10" `+uni_s), -1)
	if values := jsn.Tovalues(); !reflect.DeepEqual(values, ref) {
		t.Errorf("expected %v, got %v", ref, values)
	}
}

func TestParseJsonPointer(t *testing.T) {
	config := NewDefaultConfig()
	jptr := config.NewJsonpointer("/a/b")
	refsegs := [][]byte{[]byte("a"), []byte("b")}
	if segments := jptr.Segments(); !reflect.DeepEqual(segments, refsegs) {
		t.Errorf("expected %v, got %v", refsegs, segments)
	}
}

func TestToJsonPointer(t *testing.T) {
	config := NewDefaultConfig()
	refptr := config.NewJsonpointer("/a/b")
	jptr := config.NewJsonpointer("").ResetSegments([]string{"a", "b"})

	if bytes.Compare(jptr.path, refptr.path) != 0 {
		t.Errorf("expected %v, got %v", refptr.path, jptr.path)
	}
}

func TestGsonToCollate(t *testing.T) {
	config := NewDefaultConfig().NumberKind(IntNumber)
	clt := config.NewCollate(make([]byte, 1024), 0)
	config.NewValue(map[string]interface{}{"a": 10, "b": 20}).Tocollate(clt)
	ref := map[string]interface{}{"a": int64(10), "b": int64(20)}
	if value := clt.Tovalue(); !reflect.DeepEqual(ref, value) {
		t.Errorf("expected %v, got %v", ref, value)
	}
}

func TestCborToCollate(t *testing.T) {
	config := NewDefaultConfig().NumberKind(IntNumber)
	cbr := config.NewCbor(make([]byte, 1024), 0)
	clt := config.NewCollate(make([]byte, 1024), 0)
	out := config.NewCbor(make([]byte, 1024), 0)

	o := [][2]interface{}{
		[2]interface{}{"a", uint64(10)},
		[2]interface{}{"b", uint64(20)},
	}
	refm := CborMap2golangMap(o)

	value := config.NewValue(o).Tocbor(cbr).Tocollate(clt).Tocbor(out).Tovalue()
	if !reflect.DeepEqual(refm, value) {
		t.Errorf("expected %v, got %v", refm, value)
	}
}

func TestIsBreakCodes(t *testing.T) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(make([]byte, 1024), 0)
	config.NewValue([]interface{}{}).Tocbor(cbr)
	cbr.data = cbr.data[1:]
	if cbr.IsBreakstop() == false {
		t.Errorf("expected breakcode-array")
	}
}
