//  Copyright (c) 2015 Couchbase, Inc.

// +build ignore

package gson

import "testing"
import "reflect"
import "encoding/json"
import "fmt"

var _ = fmt.Sprintf("dummy")

func TestConfig(t *testing.T) {
	config := NewDefaultConfig()
	ref, buf := 10.2, make([]byte, 16)
	n := config.ValueToCbor(ref, buf)
	val, m := config.CborToValue(buf[:n])
	if n != m {
		t.Errorf("expected %v got %v", n, m)
	} else if !reflect.DeepEqual(ref, val) {
		t.Errorf("expected %v got %v", ref, val)
	}
}

type testLocal byte

func TestUndefined(t *testing.T) {
	config := NewDefaultConfig()
	ref, buf := CborUndefined(cborSimpleUndefined), make([]byte, 16)
	n := config.ValueToCbor(ref, buf)
	val, m := config.CborToValue(buf[:n])
	if n != m {
		t.Errorf("expected %v got %v", n, m)
	} else if !reflect.DeepEqual(ref, val) {
		t.Errorf("expected %v got %v", ref, val)
	}
	// test unknown type.
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		config.ValueToCbor(testLocal(10), buf)
	}()
}

func TestJsonToValue(t *testing.T) {
	config := NewDefaultConfig().SpaceKind(AnsiSpace)
	inp := `"abcd"  "xyz" "10" `
	out := make([]byte, 1024)
	txt, value := config.JsonToValue(inp)
	if ref := `  "xyz" "10" `; txt != ref {
		t.Errorf("expected %q, got %q", ref, txt)
	}
	n := config.ValueToJson(value, out)
	if s := string(out[:n]); s != `"abcd"` {
		t.Errorf("expected %v, got %v", `"abcd"`, s)
	}
}

func TestJsonToValues(t *testing.T) {
	var s string
	config := NewDefaultConfig().SpaceKind(AnsiSpace)
	uni_s := `"汉语 / 漢語; Hàn\b \t\uef24yǔ "`
	inp := `"abcd"  "xyz" "10" ` + uni_s
	if err := json.Unmarshal([]byte(uni_s), &s); err != nil {
		t.Fatal(err)
	}
	ref := []interface{}{"abcd", "xyz", "10", s}
	values := config.JsonToValues(inp)
	if !reflect.DeepEqual(values, ref) {
		t.Errorf("expected %v, got %v", ref, values)
	}
}

func TestParseJsonPointer(t *testing.T) {
	config := NewDefaultConfig()
	segments := config.ParseJsonPointer("/a/b", []string{})
	if ref := []string{"a", "b"}; !reflect.DeepEqual(segments, ref) {
		t.Errorf("expected %v, got %v", ref, segments)
	}
}

func TestToJsonPointer(t *testing.T) {
	config := NewDefaultConfig()
	jptr := "/a/b"
	segments := config.ParseJsonPointer(jptr, []string{})
	pointer := make([]byte, 1024)
	n := config.ToJsonPointer(segments, pointer)
	if s := string(pointer[:n]); jptr != s {
		t.Errorf("expected %v, got %v", jptr, s)
	}
}

func TestGsonToCollate(t *testing.T) {
	config := NewDefaultConfig().NumberKind(IntNumber)
	inp := map[string]interface{}{"a": 10, "b": 20}
	ref := map[string]interface{}{"a": int64(10), "b": int64(20)}
	code := make([]byte, 1024)
	n := config.ValueToCollate(inp, code)
	val, _ := config.CollateToValue(code[:n])
	if !reflect.DeepEqual(ref, val) {
		t.Errorf("expected %v, got %v", inp, val)
	}
}

func TestCborToCollate(t *testing.T) {
	config := NewDefaultConfig().NumberKind(IntNumber)
	ref := [][2]interface{}{
		[2]interface{}{"a", uint64(10)},
		[2]interface{}{"b", uint64(20)},
	}
	refm := CborMap2golangMap(ref)
	code, coll := make([]byte, 1024), make([]byte, 1024)
	out := make([]byte, 1024)
	n := config.ValueToCbor(ref, code)
	_, m := config.CborToCollate(code[:n], coll)
	_, x := config.CollateToCbor(coll[:m], out)
	val, _ := config.CborToValue(out[:x])
	if !reflect.DeepEqual(refm, val) {
		t.Errorf("expected %v, got %v", refm, val)
	}
}

func TestIsBreakCodes(t *testing.T) {
	config := NewDefaultConfig()
	out := make([]byte, 1024)
	value2cbor([]interface{}{}, out, config)
	if config.IsBreakstop(out[1]) == false {
		t.Errorf("expected breakcode-array")
	}
}
