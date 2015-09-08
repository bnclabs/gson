// +build ignore

package gson

import "testing"
import "fmt"
import "reflect"

var _ = fmt.Sprintf("dummy")

func TestGson2CollateNil(t *testing.T) {
	obj, ref := interface{}(nil), `\x02\x00`
	code, config := make([]byte, 1024), NewDefaultConfig()
	n := gson2collate(obj, code, config)
	out := fmt.Sprintf("%q", code[:n])
	out = out[1 : len(out)-1]
	if out != ref {
		t.Errorf("expected %v, got %v", ref, out)
	} else if val, _ := collate2gson(code[:n], config); val != nil && n != 2 {
		t.Errorf("expected {%v,%v}, got {%v,%v}", nil, 2, val, n)
	}
}

func TestGson2CollateTrue(t *testing.T) {
	obj, ref := true, `\x04\x00`
	code, config := make([]byte, 1024), NewDefaultConfig()
	n := gson2collate(obj, code, config)
	out := fmt.Sprintf("%q", code[:n])
	out = out[1 : len(out)-1]
	if out != ref {
		t.Errorf("expected %v, got %v", ref, out)
	} else if val, _ := collate2gson(code[:n], config); val != true && n != 2 {
		t.Errorf("expected {%v,%v}, got {%v,%v}", true, 2, val, n)
	}
}

func TestGson2CollateFalse(t *testing.T) {
	obj, ref := false, `\x03\x00`
	code, config := make([]byte, 1024), NewDefaultConfig()
	n := gson2collate(obj, code, config)
	out := fmt.Sprintf("%q", code[:n])
	out = out[1 : len(out)-1]
	if out != ref {
		t.Errorf("expected %v, got %v", ref, out)
	} else if val, _ := collate2gson(code[:n], config); val != false && n != 2 {
		t.Errorf("expected {%v,%v}, got {%v,%v}", false, 2, val, n)
	}
}

func TestGson2CollateNumber(t *testing.T) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	// as float64 using Float64 configuration
	obj, ref := float64(10.2), `\x05>>2102-\x00`
	n := gson2collate(obj, code, config.NumberKind(Float64))
	out := fmt.Sprintf("%q", code[:n])
	out = out[1 : len(out)-1]
	if out != ref {
		t.Errorf("expected %v, got %v", ref, out)
	}
	val, m := collate2gson(code[:n], config)
	if !reflect.DeepEqual(val, obj) || n != m {
		t.Errorf("expected {%v,%v}, got {%v,%v}", obj, n, val, m)
	}
	// as int64 using Float64 configuration
	obj1, ref := int64(10), `\x05>>21-\x00`
	n = gson2collate(obj1, code, config.NumberKind(Float64))
	out = fmt.Sprintf("%q", code[:n])
	out = out[1 : len(out)-1]
	if out != ref {
		t.Errorf("expected %v, got %v", ref, out)
	}
	val, m = collate2gson(code[:n], config)
	if !reflect.DeepEqual(val, float64(10.0)) || n != m {
		t.Errorf("expected {%v,%v}, got {%v,%v}", obj1, n, val, m)
	}

	// as float64 using Int64 configuration
	obj, ref = float64(10.2), `\x05>>210\x00`
	n = gson2collate(obj, code, config.NumberKind(Int64))
	out = fmt.Sprintf("%q", code[:n])
	out = out[1 : len(out)-1]
	if out != ref {
		t.Errorf("expected %v, got %v", ref, out)
	}
	val, m = collate2gson(code[:n], config)
	if !reflect.DeepEqual(val, int64(10)) || n != m {
		t.Errorf("expected {%v,%v}, got {%T,%v}", obj, n, val, m)
	}
	// as int64 using Int64 configuration
	obj1, ref = int64(10), `\x05>>210\x00`
	n = gson2collate(obj1, code, config.NumberKind(Int64))
	out = fmt.Sprintf("%q", code[:n])
	out = out[1 : len(out)-1]
	if out != ref {
		t.Errorf("expected %v, got %v", ref, out)
	}
	val, m = collate2gson(code[:n], config)
	if !reflect.DeepEqual(val, obj1) || n != m {
		t.Errorf("expected {%v,%v}, got {%v,%v}", obj1, n, val, m)
	}

	// as float64 using Decimal configuration
	obj, ref = float64(0.2), `\x05>2-\x00`
	n = gson2collate(obj, code, config.NumberKind(Decimal))
	out = fmt.Sprintf("%q", code[:n])
	out = out[1 : len(out)-1]
	if out != ref {
		t.Errorf("expected %v, got %v", ref, out)
	}
	val, m = collate2gson(code[:n], config)
	if !reflect.DeepEqual(val, obj) || n != m {
		t.Errorf("expected {%v,%v}, got {%v,%v}", obj, n, val, m)
	}
	// as int64 using Decimal configuration, expect a panic
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		gson2collate(int64(10), code, config.NumberKind(Decimal))
	}()
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		gson2collate(float64(10.2), code, config.NumberKind(Decimal))
	}()
}

func TestGson2CollateLength(t *testing.T) {
	obj, ref := Length(10), `\a>>210\x00`
	code, config := make([]byte, 1024), NewDefaultConfig()
	n := gson2collate(obj, code, config)
	out := fmt.Sprintf("%q", code[:n])
	out = out[1 : len(out)-1]
	if out != ref {
		t.Errorf("expected %v, got %v", ref, out)
	}
	val, m := collate2gson(code[:n], config)
	if !reflect.DeepEqual(val, obj) || n != m {
		t.Errorf("expected {%v,%v}, got {%v,%v}", obj, n, val, m)
	}
}

func TestGson2CollateMissing(t *testing.T) {
	obj, ref := MissingLiteral, `\x01\x00`
	code, config := make([]byte, 1024), NewDefaultConfig()
	n := gson2collate(obj, code, config)
	out := fmt.Sprintf("%q", code[:n])
	out = out[1 : len(out)-1]
	if out != ref {
		t.Errorf("expected %v, got %v", ref, out)
	}
	val, m := collate2gson(code[:n], config)
	if !reflect.DeepEqual(val, obj) || n != m {
		t.Errorf("expected {%v,%v}, got {%v,%v}", obj, n, val, m)
	}
	// expect panic when not configured for missing
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		gson2collate(MissingLiteral, code, config.UseMissing(false))
	}()
}

func TestGson2CollateString(t *testing.T) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	// empty string
	obj, ref := "", `\x06\x00\x00`
	n := gson2collate(obj, code, config)
	out := fmt.Sprintf("%q", code[:n])
	out = out[1 : len(out)-1]
	if out != ref {
		t.Errorf("expected %v, got %v", ref, out)
	}
	val, m := collate2gson(code[:n], config)
	if !reflect.DeepEqual(val, obj) || n != m {
		t.Errorf("expected {%v,%v}, got {%v,%v}", obj, n, val, m)
	}
	// normal string
	obj, ref = "hello world", `\x06hello world\x00\x00`
	n = gson2collate(obj, code, config)
	out = fmt.Sprintf("%q", code[:n])
	out = out[1 : len(out)-1]
	if out != ref {
		t.Errorf("expected %v, got %v", ref, out)
	}
	val, m = collate2gson(code[:n], config)
	if !reflect.DeepEqual(val, obj) || n != m {
		t.Errorf("expected {%v,%v}, got {%v,%v}", obj, n, val, m)
	}
	// missing string
	obj, ref = string(MissingLiteral), `\x01\x00`
	n = gson2collate(obj, code, config)
	out = fmt.Sprintf("%q", code[:n])
	out = out[1 : len(out)-1]
	if out != ref {
		t.Errorf("expected %v, got %v", ref, out)
	}
	val, m = collate2gson(code[:n], config)
	if !reflect.DeepEqual(val, Missing(obj)) || n != m {
		t.Errorf("expected {%v,%v}, got {%v,%v}", obj, n, val, m)
	}
	// missing string without doMissing configuration
	obj, ref = string(MissingLiteral), `\x06~[]{}falsenilNA~\x00\x00`
	n = gson2collate(obj, code, config.UseMissing(false))
	out = fmt.Sprintf("%q", code[:n])
	out = out[1 : len(out)-1]
	if out != ref {
		t.Errorf("expected %v, got %v", ref, out)
	}
	val, m = collate2gson(code[:n], config)
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
		n := gson2collate(obj, code, config)
		out := fmt.Sprintf("%q", code[:n])
		out = out[1 : len(out)-1]
		if out != ref {
			t.Errorf("expected %v, got %v", ref, out)
		}
		val, m := collate2gson(code[:n], config)
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
		n := gson2collate(obj, code, config)
		out := fmt.Sprintf("%q", code[:n])
		out = out[1 : len(out)-1]
		if out != ref {
			t.Errorf("expected %v, got %v", ref, out)
		}
		val, m := collate2gson(code[:n], config)
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
		n := gson2collate(obj, code, config)
		out := fmt.Sprintf("%q", code[:n])
		out = out[1 : len(out)-1]
		if out != ref {
			t.Errorf("expected %v, got %v", ref, out)
		}
		val, m := collate2gson(code[:n], config)
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
		n := gson2collate(obj, code, config.SortbyPropertyLen(false))
		out := fmt.Sprintf("%q", code[:n])
		out = out[1 : len(out)-1]
		if out != ref {
			t.Errorf("expected %v, got %v", ref, out)
		}
		val, m := collate2gson(code[:n], config)
		if !reflect.DeepEqual(val, obj) || n != m {
			t.Errorf("expected {%v,%v}, got {%v,%v}", obj, n, val, m)
		}
	}
}

func BenchmarkGsonCollNil(b *testing.B) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	for i := 0; i < b.N; i++ {
		gson2collate(nil, code, config)
	}
}

func BenchmarkCollGsonNil(b *testing.B) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	n := gson2collate(nil, code, config)
	for i := 0; i < b.N; i++ {
		collate2gson(code[:n], config)
	}
}

func BenchmarkGsonCollTrue(b *testing.B) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	val := interface{}(true)
	for i := 0; i < b.N; i++ {
		gson2collate(val, code, config)
	}
}

func BenchmarkCollGsonTrue(b *testing.B) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	n := gson2collate(interface{}(true), code, config)
	for i := 0; i < b.N; i++ {
		collate2gson(code[:n], config)
	}
}

func BenchmarkGsonCollFalse(b *testing.B) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	val := interface{}(false)
	for i := 0; i < b.N; i++ {
		gson2collate(val, code, config)
	}
}

func BenchmarkCollGsonFalse(b *testing.B) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	n := gson2collate(interface{}(false), code, config)
	for i := 0; i < b.N; i++ {
		collate2gson(code[:n], config)
	}
}

func BenchmarkGsonCollF64(b *testing.B) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	val := interface{}(float64(10.121312213123123))
	for i := 0; i < b.N; i++ {
		gson2collate(val, code, config)
	}
}

func BenchmarkCollGsonF64(b *testing.B) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	n := gson2collate(float64(10.121312213123123), code, config)
	for i := 0; i < b.N; i++ {
		collate2gson(code[:n], config)
	}
}

func BenchmarkGsonCollI64(b *testing.B) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	val := interface{}(int64(123456789))
	for i := 0; i < b.N; i++ {
		gson2collate(val, code, config)
	}
}

func BenchmarkCollGsonI64(b *testing.B) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	n := gson2collate(int64(123456789), code, config)
	for i := 0; i < b.N; i++ {
		collate2gson(code[:n], config)
	}
}

func BenchmarkGsonCollMiss(b *testing.B) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	val := interface{}(MissingLiteral)
	for i := 0; i < b.N; i++ {
		gson2collate(val, code, config)
	}
}

func BenchmarkCollGsonMiss(b *testing.B) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	n := gson2collate(MissingLiteral, code, config)
	for i := 0; i < b.N; i++ {
		collate2gson(code[:n], config)
	}
}

func BenchmarkGsonCollStr(b *testing.B) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	val := interface{}("hello world")
	for i := 0; i < b.N; i++ {
		gson2collate(val, code, config)
	}
}

func BenchmarkCollGsonStr(b *testing.B) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	n := gson2collate("hello world", code, config)
	for i := 0; i < b.N; i++ {
		collate2gson(code[:n], config)
	}
}

func BenchmarkGsonCollArr(b *testing.B) {
	arr := []interface{}{nil, true, false, "hello world", 10.23122312}
	code, config := make([]byte, 1024), NewDefaultConfig()
	val := interface{}(arr)
	for i := 0; i < b.N; i++ {
		gson2collate(val, code, config)
	}
}

func BenchmarkCollGsonArr(b *testing.B) {
	arr := []interface{}{nil, true, false, "hello world", 10.23122312}
	code, config := make([]byte, 1024), NewDefaultConfig()
	n := gson2collate(arr, code, config)
	for i := 0; i < b.N; i++ {
		collate2gson(code[:n], config)
	}
}

func BenchmarkGsonCollMap(b *testing.B) {
	obj := map[string]interface{}{
		"key1": nil, "key2": true, "key3": false, "key4": "hello world",
		"key5": 10.23122312,
	}
	code, config := make([]byte, 1024), NewDefaultConfig()
	val := interface{}(obj)
	for i := 0; i < b.N; i++ {
		gson2collate(val, code, config)
	}
}

func BenchmarkCollGsonMap(b *testing.B) {
	obj := map[string]interface{}{
		"key1": nil, "key2": true, "key3": false, "key4": "hello world",
		"key5": 10.23122312,
	}
	code, config := make([]byte, 1024), NewDefaultConfig()
	n := gson2collate(obj, code, config)
	for i := 0; i < b.N; i++ {
		collate2gson(code[:n], config)
	}
}
