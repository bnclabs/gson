//  Copyright (c) 2015 Couchbase, Inc.

package gson

import "testing"
import "fmt"
import "reflect"

var _ = fmt.Sprintf("dummy")

func TestGson2CollateNil(t *testing.T) {
	obj, ref := interface{}(nil), `\x02\x00`
	code, config := make([]byte, 1024), NewDefaultConfig()
	n := config.ValueToCollate(obj, code)
	out := fmt.Sprintf("%q", code[:n])
	out = out[1 : len(out)-1]
	if out != ref {
		t.Errorf("expected %v, got %v", ref, out)
	}
	if val, _ := config.CollateToValue(code[:n]); val != nil && n != 2 {
		t.Errorf("expected {%v,%v}, got {%v,%v}", nil, 2, val, n)
	}
}

func TestGson2CollateTrue(t *testing.T) {
	obj, ref := true, `\x04\x00`
	code, config := make([]byte, 1024), NewDefaultConfig()
	n := config.ValueToCollate(obj, code)
	out := fmt.Sprintf("%q", code[:n])
	out = out[1 : len(out)-1]
	if out != ref {
		t.Errorf("expected %v, got %v", ref, out)
	}
	if val, _ := config.CollateToValue(code[:n]); val != true && n != 2 {
		t.Errorf("expected {%v,%v}, got {%v,%v}", true, 2, val, n)
	}
}

func TestGson2CollateFalse(t *testing.T) {
	obj, ref := false, `\x03\x00`
	code, config := make([]byte, 1024), NewDefaultConfig()
	n := config.ValueToCollate(obj, code)
	out := fmt.Sprintf("%q", code[:n])
	out = out[1 : len(out)-1]
	if out != ref {
		t.Errorf("expected %v, got %v", ref, out)
	}
	if val, _ := config.CollateToValue(code[:n]); val != false && n != 2 {
		t.Errorf("expected {%v,%v}, got {%v,%v}", false, 2, val, n)
	}
}

func TestGson2CollateNumber(t *testing.T) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	// as float64 using FloatNumber configuration
	obj, ref := float64(10.2), `\x05>>2102-\x00`
	config = config.NumberKind(FloatNumber)
	n := config.ValueToCollate(obj, code)
	out := fmt.Sprintf("%q", code[:n])
	out = out[1 : len(out)-1]
	if out != ref {
		t.Errorf("expected %v, got %v", ref, out)
	}
	val, m := config.CollateToValue(code[:n])
	if !reflect.DeepEqual(val, obj) || n != m {
		t.Errorf("expected {%v,%v}, got {%v,%v}", obj, n, val, m)
	}
	// as int64 using FloatNumber configuration
	obj1, ref := int64(10), `\x05>>21-\x00`
	config = config.NumberKind(FloatNumber)
	n = config.ValueToCollate(obj1, code)
	out = fmt.Sprintf("%q", code[:n])
	out = out[1 : len(out)-1]
	if out != ref {
		t.Errorf("expected %v, got %v", ref, out)
	}
	val, m = config.CollateToValue(code[:n])
	if !reflect.DeepEqual(val, float64(10.0)) || n != m {
		t.Errorf("expected {%v,%v}, got {%v,%v}", obj1, n, val, m)
	}

	// as float64 using IntNumber configuration
	obj, ref = float64(10.2), `\x05>>210\x00`
	config = config.NumberKind(IntNumber)
	n = config.ValueToCollate(obj, code)
	out = fmt.Sprintf("%q", code[:n])
	out = out[1 : len(out)-1]
	if out != ref {
		t.Errorf("expected %v, got %v", ref, out)
	}
	val, m = config.CollateToValue(code[:n])
	if !reflect.DeepEqual(val, int64(10)) || n != m {
		t.Errorf("expected {%v,%v}, got {%T,%v}", obj, n, val, m)
	}
	// as int64 using IntNumber configuration
	obj1, ref = int64(10), `\x05>>210\x00`
	config = config.NumberKind(IntNumber)
	n = config.ValueToCollate(obj1, code)
	out = fmt.Sprintf("%q", code[:n])
	out = out[1 : len(out)-1]
	if out != ref {
		t.Errorf("expected %v, got %v", ref, out)
	}
	val, m = config.CollateToValue(code[:n])
	if !reflect.DeepEqual(val, obj1) || n != m {
		t.Errorf("expected {%v,%v}, got {%v,%v}", obj1, n, val, m)
	}

	// as float64 using Decimal configuration
	obj, ref = float64(0.2), `\x05>2-\x00`
	config = config.NumberKind(Decimal)
	n = config.ValueToCollate(obj, code)
	out = fmt.Sprintf("%q", code[:n])
	out = out[1 : len(out)-1]
	if out != ref {
		t.Errorf("expected %v, got %v", ref, out)
	}
	val, m = config.CollateToValue(code[:n])
	if !reflect.DeepEqual(val, obj) || n != m {
		t.Errorf("expected {%v,%v}, got {%v,%v}", obj, n, val, m)
	}
	// as int64 using Decimal configuration, expect a panic
	config = config.NumberKind(Decimal)
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		config.ValueToCollate(int64(10), code)
	}()
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		config.ValueToCollate(float64(10.2), code)
	}()
}

func TestGson2CollateLength(t *testing.T) {
	obj, ref := Length(10), `\a>>210\x00`
	code, config := make([]byte, 1024), NewDefaultConfig()
	n := config.ValueToCollate(obj, code)
	out := fmt.Sprintf("%q", code[:n])
	out = out[1 : len(out)-1]
	if out != ref {
		t.Errorf("expected %v, got %v", ref, out)
	}
	val, m := config.CollateToValue(code[:n])
	if !reflect.DeepEqual(val, obj) || n != m {
		t.Errorf("expected {%v,%v}, got {%v,%v}", obj, n, val, m)
	}
}

func TestGson2CollateMissing(t *testing.T) {
	obj, ref := MissingLiteral, `\x01\x00`
	code, config := make([]byte, 1024), NewDefaultConfig()
	n := config.ValueToCollate(obj, code)
	out := fmt.Sprintf("%q", code[:n])
	out = out[1 : len(out)-1]
	if out != ref {
		t.Errorf("expected %v, got %v", ref, out)
	}
	val, m := config.CollateToValue(code[:n])
	if !reflect.DeepEqual(val, obj) || n != m {
		t.Errorf("expected {%v,%v}, got {%v,%v}", obj, n, val, m)
	}
	// expect panic when not configured for missing
	config = config.UseMissing(false)
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		config.ValueToCollate(MissingLiteral, code)
	}()
}

func TestGson2CollateString(t *testing.T) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	testcases := [][2]interface{}{
		[2]interface{}{"", `\x06\x00\x00`},
		[2]interface{}{"hello world", `\x06hello world\x00\x00`},
		[2]interface{}{string(MissingLiteral), `\x01\x00`},
	}
	for _, tcase := range testcases {
		obj, ref := tcase[0].(string), tcase[1].(string)
		n := config.ValueToCollate(obj, code)
		out := fmt.Sprintf("%q", code[:n])
		out = out[1 : len(out)-1]
		if out != ref {
			t.Errorf("expected %v, got %v", ref, out)
		}
		val, m := config.CollateToValue(code[:n])
		if s, ok := val.(string); ok {
			if s != obj || n != m {
				t.Errorf("expected {%v,%v}, got {%v,%v}", obj, n, val, m)
			}
		} else if s := string(val.(Missing)); s != obj || n != m {
			t.Errorf("expected {%v,%v}, got {%v,%v}", obj, n, val, m)
		}
	}
	// missing string without doMissing configuration
	obj, ref := string(MissingLiteral), `\x06~[]{}falsenilNA~\x00\x00`
	config = config.UseMissing(false)
	n := config.ValueToCollate(obj, code)
	out := fmt.Sprintf("%q", code[:n])
	out = out[1 : len(out)-1]
	if out != ref {
		t.Errorf("expected %v, got %v", ref, out)
	}
	val, m := config.CollateToValue(code[:n])
	if !reflect.DeepEqual(val, obj) || n != m {
		t.Errorf("expected {%v,%v}, got {%v,%v}", obj, n, val, m)
	}
}

func TestGson2CollateBytes(t *testing.T) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	obj, ref := []byte("hello world"), `\nhello world\x00`
	n := config.ValueToCollate(obj, code)
	out := fmt.Sprintf("%q", code[:n])
	out = out[1 : len(out)-1]
	if out != ref {
		t.Errorf("expected %v, got %v", ref, out)
	}
	val, m := config.CollateToValue(code[:n])
	if !reflect.DeepEqual(val, obj) || n != m {
		t.Errorf("expected {%v,%v}, got {%v,%v}", obj, n, val, m)
	}
}

func TestGson2CollateArray(t *testing.T) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	// without length prefix
	testcases := [][2]interface{}{
		[2]interface{}{[]interface{}{nil, true, false, 10.0, "hello"},
			`\b\x02\x00\x04\x00\x03\x00\x05>>21-\x00\x06hello\x00\x00\x00`},
		[2]interface{}{[]interface{}{},
			`\b\x00`},
		[2]interface{}{[]interface{}{
			nil, true, 10.0, 10.2, []interface{}{},
			map[string]interface{}{"key": map[string]interface{}{}}},
			`\b\x02\x00\x04\x00\x05>>21-\x00\x05>>2102-\x00\b\x00` +
				`\t\a>1\x00\x06key\x00\x00\t\a0\x00\x00\x00\x00`},
	}
	for _, tcase := range testcases {
		t.Logf("%v", tcase[0])
		obj, ref := tcase[0], tcase[1].(string)
		n := config.ValueToCollate(obj, code)
		out := fmt.Sprintf("%q", code[:n])
		out = out[1 : len(out)-1]
		if out != ref {
			t.Errorf("expected %v, got %v", ref, out)
		}
		val, m := config.CollateToValue(code[:n])
		if !reflect.DeepEqual(val, obj) || n != m {
			t.Errorf("expected {%v,%v}, got {%v,%v}", obj, n, val, m)
		}
	}
	// with length prefix
	config = config.SortbyArrayLen(true)
	testcases = [][2]interface{}{
		[2]interface{}{[]interface{}{nil, true, false, 10.0, "hello"},
			`\b\a>5\x00\x02\x00\x04\x00\x03\x00\x05>>21-\x00` +
				`\x06hello\x00\x00\x00`},
		[2]interface{}{[]interface{}{},
			`\b\a0\x00\x00`},
		[2]interface{}{[]interface{}{
			nil, true, 10.0, 10.2, []interface{}{},
			map[string]interface{}{"key": map[string]interface{}{}}},
			`\b\a>6\x00\x02\x00\x04\x00\x05>>21-\x00\x05>>2102-\x00` +
				`\b\a0\x00\x00\t\a>1\x00\x06key\x00\x00\t\a0\x00\x00\x00\x00`},
	}
	for _, tcase := range testcases {
		t.Logf("%v", tcase[0])
		obj, ref := tcase[0], tcase[1].(string)
		n := config.ValueToCollate(obj, code)
		out := fmt.Sprintf("%q", code[:n])
		out = out[1 : len(out)-1]
		if out != ref {
			t.Errorf("expected %v, got %v", ref, out)
		}
		val, m := config.CollateToValue(code[:n])
		if !reflect.DeepEqual(val, obj) || n != m {
			t.Errorf("expected {%v,%v}, got {%v,%v}", obj, n, val, m)
		}
	}
}

func TestGson2CollateMap(t *testing.T) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	// with length prefix
	testcases := [][2]interface{}{
		[2]interface{}{
			map[string]interface{}{
				"a": nil, "b": true, "c": false, "d": 10.0, "e": "hello"},
			`\t\a>5\x00\x06a\x00\x00\x02\x00\x06b\x00\x00\x04\x00\x06c` +
				`\x00\x00\x03\x00\x06d\x00\x00\x05>>21-\x00\x06e\x00\x00` +
				`\x06hello\x00\x00\x00`},
	}
	for _, tcase := range testcases {
		t.Logf("%v", tcase[0])
		obj, ref := tcase[0], tcase[1].(string)
		n := config.ValueToCollate(obj, code)
		out := fmt.Sprintf("%q", code[:n])
		out = out[1 : len(out)-1]
		if out != ref {
			t.Errorf("expected %v, got %v", ref, out)
		}
		val, m := config.CollateToValue(code[:n])
		if !reflect.DeepEqual(val, obj) || n != m {
			t.Errorf("expected {%v,%v}, got {%v,%v}", obj, n, val, m)
		}
	}
	// without length prefix
	testcases = [][2]interface{}{
		[2]interface{}{
			map[string]interface{}{
				"a": nil, "b": true, "c": false, "d": 10.0, "e": "hello"},
			`\t\x06a\x00\x00\x02\x00\x06b\x00\x00\x04\x00\x06c\x00\x00` +
				`\x03\x00\x06d\x00\x00\x05>>21-\x00\x06e\x00\x00` +
				`\x06hello\x00\x00\x00`},
	}
	for _, tcase := range testcases {
		t.Logf("%v", tcase[0])
		obj, ref := tcase[0], tcase[1].(string)
		config = config.SortbyPropertyLen(false)
		n := config.ValueToCollate(obj, code)
		out := fmt.Sprintf("%q", code[:n])
		out = out[1 : len(out)-1]
		if out != ref {
			t.Errorf("expected %v, got %v", ref, out)
		}
		val, m := config.CollateToValue(code[:n])
		if !reflect.DeepEqual(val, obj) || n != m {
			t.Errorf("expected {%v,%v}, got {%v,%v}", obj, n, val, m)
		}
	}
}

func BenchmarkVal2CollNil(b *testing.B) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	for i := 0; i < b.N; i++ {
		config.ValueToCollate(nil, code)
	}
}

func BenchmarkColl2ValNil(b *testing.B) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	n := config.ValueToCollate(nil, code)
	for i := 0; i < b.N; i++ {
		config.CollateToValue(code[:n])
	}
}

func BenchmarkVal2CollTrue(b *testing.B) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	val := interface{}(true)
	for i := 0; i < b.N; i++ {
		config.ValueToCollate(val, code)
	}
}

func BenchmarkColl2ValTrue(b *testing.B) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	n := config.ValueToCollate(interface{}(true), code)
	for i := 0; i < b.N; i++ {
		config.CollateToValue(code[:n])
	}
}

func BenchmarkVal2CollFalse(b *testing.B) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	val := interface{}(false)
	for i := 0; i < b.N; i++ {
		config.ValueToCollate(val, code)
	}
}

func BenchmarkColl2ValFalse(b *testing.B) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	n := config.ValueToCollate(interface{}(false), code)
	for i := 0; i < b.N; i++ {
		config.CollateToValue(code[:n])
	}
}

func BenchmarkVal2CollF64(b *testing.B) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	val := interface{}(float64(10.121312213123123))
	for i := 0; i < b.N; i++ {
		config.ValueToCollate(val, code)
	}
}

func BenchmarkColl2ValF64(b *testing.B) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	n := config.ValueToCollate(float64(10.121312213123123), code)
	for i := 0; i < b.N; i++ {
		config.CollateToValue(code[:n])
	}
}

func BenchmarkVal2CollI64(b *testing.B) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	val := interface{}(int64(123456789))
	for i := 0; i < b.N; i++ {
		config.ValueToCollate(val, code)
	}
}

func BenchmarkColl2ValI64(b *testing.B) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	n := config.ValueToCollate(int64(123456789), code)
	for i := 0; i < b.N; i++ {
		config.CollateToValue(code[:n])
	}
}

func BenchmarkVal2CollMiss(b *testing.B) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	val := interface{}(MissingLiteral)
	for i := 0; i < b.N; i++ {
		config.ValueToCollate(val, code)
	}
}

func BenchmarkColl2ValMiss(b *testing.B) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	n := config.ValueToCollate(MissingLiteral, code)
	for i := 0; i < b.N; i++ {
		config.CollateToValue(code[:n])
	}
}

func BenchmarkVal2CollStr(b *testing.B) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	val := interface{}("hello world")
	for i := 0; i < b.N; i++ {
		config.ValueToCollate(val, code)
	}
}

func BenchmarkColl2ValStr(b *testing.B) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	n := config.ValueToCollate("hello world", code)
	for i := 0; i < b.N; i++ {
		config.CollateToValue(code[:n])
	}
}

func BenchmarkVal2CollArr(b *testing.B) {
	arr := []interface{}{nil, true, false, "hello world", 10.23122312}
	code, config := make([]byte, 1024), NewDefaultConfig()
	val := interface{}(arr)
	for i := 0; i < b.N; i++ {
		config.ValueToCollate(val, code)
	}
}

func BenchmarkColl2ValArr(b *testing.B) {
	arr := []interface{}{nil, true, false, "hello world", 10.23122312}
	code, config := make([]byte, 1024), NewDefaultConfig()
	n := config.ValueToCollate(arr, code)
	for i := 0; i < b.N; i++ {
		config.CollateToValue(code[:n])
	}
}

func BenchmarkVal2CollMap(b *testing.B) {
	obj := map[string]interface{}{
		"key1": nil, "key2": true, "key3": false, "key4": "hello world",
		"key5": 10.23122312,
	}
	code, config := make([]byte, 1024), NewDefaultConfig()
	val := interface{}(obj)
	for i := 0; i < b.N; i++ {
		config.ValueToCollate(val, code)
	}
}

func BenchmarkColl2ValMap(b *testing.B) {
	obj := map[string]interface{}{
		"key1": nil, "key2": true, "key3": false, "key4": "hello world",
		"key5": 10.23122312,
	}
	code, config := make([]byte, 1024), NewDefaultConfig()
	n := config.ValueToCollate(obj, code)
	for i := 0; i < b.N; i++ {
		config.CollateToValue(code[:n])
	}
}
