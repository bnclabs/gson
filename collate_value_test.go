package gson

import "testing"
import "fmt"
import "reflect"

var _ = fmt.Sprintf("dummy")

func TestVal2CollateNil(t *testing.T) {
	ref := `\f\x00`
	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 1024), 0)

	config.NewValue(interface{}(nil)).Tocollate(clt)

	out := fmt.Sprintf("%q", clt.Bytes())
	if out = out[1 : len(out)-1]; out != ref {
		t.Errorf("expected %v, got %v", ref, out)
	} else if value := clt.Tovalue(); value != nil {
		t.Errorf("expected %v got %v", nil, value)
	}
}

func TestVal2CollateTrue(t *testing.T) {
	ref := `\x0e\x00`
	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 1024), 0)

	config.NewValue(true).Tocollate(clt)
	out := fmt.Sprintf("%q", clt.Bytes())
	if out = out[1 : len(out)-1]; out != ref {
		t.Errorf("expected %v, got %v", ref, out)
	} else if value := clt.Tovalue(); value != true {
		t.Errorf("expected %v got %v", true, value)
	}
}

func TestVal2CollateFalse(t *testing.T) {
	ref := `\r\x00`
	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 1024), 0)

	config.NewValue(false).Tocollate(clt)
	out := fmt.Sprintf("%q", clt.Bytes())
	if out = out[1 : len(out)-1]; out != ref {
		t.Errorf("expected %v, got %v", ref, out)
	} else if value := clt.Tovalue(); value != false {
		t.Errorf("expected %v got %v", false, value)
	}
}

func TestVal2CollateNumber(t *testing.T) {
	// as float64 using FloatNumber configuration
	objf, ref := float64(10.2), `\x0f>>2102-\x00`
	config := NewDefaultConfig().SetNumberKind(FloatNumber)
	clt := config.NewCollate(make([]byte, 1024), 0)

	config.NewValue(objf).Tocollate(clt)
	out := fmt.Sprintf("%q", clt.Bytes())
	if out = out[1 : len(out)-1]; out != ref {
		t.Errorf("expected %v, got %v", ref, out)
	} else if value := clt.Tovalue(); !reflect.DeepEqual(objf, value) {
		t.Errorf("expected %v got %v", objf, value)
	}

	// as int64 using FloatNumber configuration
	obji, ref := int64(10), `\x0f>>21-\x00`
	config = config.SetNumberKind(FloatNumber)
	clt = config.NewCollate(make([]byte, 1024), 0)

	config.NewValue(obji).Tocollate(clt)
	out = fmt.Sprintf("%q", clt.Bytes())
	if out = out[1 : len(out)-1]; out != ref {
		t.Errorf("expected %v, got %v", ref, out)
	} else if value := clt.Tovalue(); !reflect.DeepEqual(value, float64(10.0)) {
		t.Errorf("expected %v, got %v", obji, value)
	}

	// as float32 using FloatNumber configuration: FIXME
	objf32, ref := float32(10.2), `\x0f>>210199999809265137-\x00`
	config = config.SetNumberKind(FloatNumber)
	clt = config.NewCollate(make([]byte, 1024), 0)

	config.NewValue(objf32).Tocollate(clt)
	out = fmt.Sprintf("%q", clt.Bytes())
	if out = out[1 : len(out)-1]; out != ref {
		t.Errorf("expected %v, got %v", ref, out)
	}
	value := clt.Tovalue().(float64)
	if !reflect.DeepEqual(value, 10.199999809265137) {
		t.Errorf("expected %v, got %v", 10.199999809265137, float64(value))
	}

	// as float64 using FloatNumber configuration
	objf, ref = float64(10.2), `\x0f>>2102-\x00`
	objr := float64(10.2)
	config = config.SetNumberKind(FloatNumber)
	clt = config.NewCollate(make([]byte, 1024), 0)

	config.NewValue(objf).Tocollate(clt)
	out = fmt.Sprintf("%q", clt.Bytes())
	if out = out[1 : len(out)-1]; out != ref {
		t.Errorf("expected %v, got %v", ref, out)
	} else if value := clt.Tovalue(); !reflect.DeepEqual(value, objr) {
		t.Errorf("expected %v, got %v", objr, value)
	}

	// as int64 using SmartNumber configuration
	obji, ref = int64(10), `\x0f>>21-\x00`
	config = config.SetNumberKind(SmartNumber)
	clt = config.NewCollate(make([]byte, 1024), 0)

	config.NewValue(obji).Tocollate(clt)
	out = fmt.Sprintf("%q", clt.Bytes())
	if out = out[1 : len(out)-1]; out != ref {
		t.Errorf("expected %v, got %v", ref, out)
	} else if value := clt.Tovalue(); !reflect.DeepEqual(value, float64(obji)) {
		t.Errorf("expected %v, got %v", obji, value)
	}

	// as uint64 using SmartNumber configuration
	obju, ref := uint64(10), `\x0f>>21-\x00`
	config = config.SetNumberKind(SmartNumber)
	clt = config.NewCollate(make([]byte, 1024), 0)

	config.NewValue(obju).Tocollate(clt)
	out = fmt.Sprintf("%q", clt.Bytes())
	if out = out[1 : len(out)-1]; out != ref {
		t.Errorf("expected %v, got %v", ref, out)
	}
	if value := clt.Tovalue().(float64); !reflect.DeepEqual(uint64(value), obju) {
		t.Errorf("expected %v, got %v", obju, value)
	}

	// as float64 using SmartNumber configuration
	objf, ref = float64(0.2), `\x0f>02-\x00`
	config = config.SetNumberKind(SmartNumber)
	clt = config.NewCollate(make([]byte, 1024), 0)

	config.NewValue(objf).Tocollate(clt)
	out = fmt.Sprintf("%q", clt.Bytes())
	if out = out[1 : len(out)-1]; out != ref {
		t.Errorf("expected %v, got %v", ref, out)
	} else if value := clt.Tovalue(); !reflect.DeepEqual(value, objf) {
		t.Errorf("expected %v, got %v", objf, value)
	}
}

func TestVal2CollateMissing(t *testing.T) {
	obj, ref := MissingLiteral, `\v\x00`
	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 1024), 0)

	config.NewValue(obj).Tocollate(clt)
	out := fmt.Sprintf("%q", clt.Bytes())
	if out = out[1 : len(out)-1]; out != ref {
		t.Errorf("expected %v, got %v", ref, out)
	} else if value := clt.Tovalue(); !reflect.DeepEqual(value, obj) {
		t.Errorf("expected %v, got %v", obj, value)
	}

	// expect panic when not configured for missing
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		config = config.UseMissing(false)
		clt := config.NewCollate(make([]byte, 1024), 0)
		config.NewValue(MissingLiteral).Tocollate(clt)
	}()
}

func TestVal2CollateString(t *testing.T) {
	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 1024), 0)

	testcases := [][2]interface{}{
		{"", `\x10\x00\x00`},
		{"hello world", `\x10hello world\x00\x00`},
		{string(MissingLiteral), `\v\x00`},
	}
	for _, tcase := range testcases {
		obj, ref := tcase[0].(string), tcase[1].(string)

		config.NewValue(obj).Tocollate(clt.Reset(nil))
		out := fmt.Sprintf("%q", clt.Bytes())
		if out = out[1 : len(out)-1]; out != ref {
			t.Errorf("expected %v, got %v", ref, out)
		}
		value := clt.Tovalue()
		if s, ok := value.(string); ok {
			if s != obj {
				t.Errorf("expected %v got %v", obj, value)
			}
		} else if s := string(value.(Missing)); s != obj {
			t.Errorf("expected %v, got %v", obj, value)
		}
	}

	// missing string without doMissing configuration
	obj, ref := string(MissingLiteral), `\x10~[]{}falsenilNA~\x00\x00`
	config = config.UseMissing(false)

	config.NewValue(obj).Tocollate(clt.Reset(nil))
	out := fmt.Sprintf("%q", clt.Bytes())
	if out = out[1 : len(out)-1]; out != ref {
		t.Errorf("expected %v, got %v", ref, out)
	} else if value := clt.Tovalue(); !reflect.DeepEqual(value, obj) {
		t.Errorf("expected %v, got %v", obj, value)
	}
}

func TestVal2CollateBytes(t *testing.T) {
	obj, ref := []byte("hello world"), `\x14hello world\x00`
	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 1024), 0)

	config.NewValue(obj).Tocollate(clt)
	out := fmt.Sprintf("%q", clt.Bytes())
	if out = out[1 : len(out)-1]; out != ref {
		t.Errorf("expected %v, got %v", ref, out)
	} else if value := clt.Tovalue(); !reflect.DeepEqual(value, obj) {
		t.Errorf("expected %v, got %v", obj, value)
	}
}

func TestVal2CollateArray(t *testing.T) {
	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 1024), 0)

	// without length prefix
	testcases := [][2]interface{}{
		{[]interface{}{nil, true, false, 10.0, "hello"},
			`\x12\f\x00\x0e\x00\r\x00\x0f>>21-\x00\x10hello\x00\x00\x00`},
		{[]interface{}{},
			`\x12\x00`},
		{[]interface{}{
			nil, true, 10.0, 10.2, []interface{}{},
			map[string]interface{}{"key": map[string]interface{}{}}},
			`\x12\f\x00\x0e\x00\x0f>>21-\x00\x0f>>2102-\x00\x12\x00` +
				`\x13\x11>1\x00\x10key\x00\x00\x13\x110\x00\x00\x00\x00`},
	}
	for _, tcase := range testcases {
		obj, ref := tcase[0], tcase[1].(string)
		config.NewValue(obj).Tocollate(clt.Reset(nil))

		t.Logf("%v %v", tcase[0], clt.Bytes())

		out := fmt.Sprintf("%q", clt.Bytes())
		if out = out[1 : len(out)-1]; out != ref {
			t.Errorf("expected %v, got %v", ref, out)
		} else if value := clt.Tovalue(); !reflect.DeepEqual(value, obj) {
			t.Errorf("expected %v, got %v", obj, value)
		}
	}

	// with length prefix
	config = config.SortbyArrayLen(true)
	clt = config.NewCollate(make([]byte, 1024), 0)
	testcases = [][2]interface{}{
		{[]interface{}{nil, true, false, 10.0, "hello"},
			`\x12\x11>5\x00\f\x00\x0e\x00\r\x00\x0f>>21-\x00\x10` +
				`hello\x00\x00\x00`},
		{[]interface{}{},
			`\x12\x110\x00\x00`},
		{[]interface{}{
			nil, true, 10.0, 10.2, []interface{}{},
			map[string]interface{}{"key": map[string]interface{}{}}},
			`\x12\x11>6\x00\f\x00\x0e\x00\x0f>>21-\x00\x0f>>2102-\x00` +
				`\x12\x110\x00\x00\x13\x11>1\x00\x10key\x00\x00` +
				`\x13\x110\x00\x00\x00\x00`},
	}
	for _, tcase := range testcases {
		obj, ref := tcase[0], tcase[1].(string)
		config.NewValue(obj).Tocollate(clt.Reset(nil))

		t.Logf("%v %v", tcase[0], clt.Bytes())

		out := fmt.Sprintf("%q", clt.Bytes())
		if out = out[1 : len(out)-1]; out != ref {
			t.Errorf("expected %v, got %v", ref, out)
		} else if value := clt.Tovalue(); !reflect.DeepEqual(value, obj) {
			t.Errorf("expected %v, got %v", obj, value)
		}
	}
}

func TestVal2CollateMap(t *testing.T) {
	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 1024), 0)
	// with length prefix
	testcases := [][2]interface{}{
		{
			map[string]interface{}{
				"a": nil, "b": true, "c": false, "d": 10.0, "e": "hello"},
			`\x13\x11>5\x00\x10a\x00\x00\f\x00\x10b\x00\x00\x0e\x00\x10c` +
				`\x00\x00\r\x00\x10d\x00\x00\x0f>>21-\x00\x10e\x00\x00\x10` +
				`hello\x00\x00\x00`},
	}
	for _, tcase := range testcases {
		t.Logf("%v", tcase[0])
		obj, ref := tcase[0], tcase[1].(string)

		config.NewValue(obj).Tocollate(clt.Reset(nil))
		out := fmt.Sprintf("%q", clt.Bytes())
		if out = out[1 : len(out)-1]; out != ref {
			t.Errorf("expected %v, got %v", ref, out)
		} else if value := clt.Tovalue(); !reflect.DeepEqual(value, obj) {
			t.Errorf("expected %v, got %v", obj, value)
		}
	}

	// without length prefix
	config = config.SortbyPropertyLen(false)
	testcases = [][2]interface{}{
		{
			map[string]interface{}{
				"a": nil, "b": true, "c": false, "d": 10.0, "e": "hello"},
			`\x13\x10a\x00\x00\f\x00\x10b\x00\x00\x0e\x00\x10c\x00` +
				`\x00\r\x00\x10d\x00\x00\x0f>>21-\x00\x10e\x00\x00\x10hello` +
				`\x00\x00\x00`},
	}
	for _, tcase := range testcases {
		t.Logf("%v", tcase[0])
		obj, ref := tcase[0], tcase[1].(string)
		clt := config.NewCollate(make([]byte, 1024), 0)

		config.NewValue(obj).Tocollate(clt.Reset(nil))
		out := fmt.Sprintf("%q", clt.Bytes())
		if out = out[1 : len(out)-1]; out != ref {
			t.Errorf("expected %v, got %v", ref, out)
		} else if value := clt.Tovalue(); !reflect.DeepEqual(value, obj) {
			t.Errorf("expected %v, got %v", obj, value)
		}
	}
}

func BenchmarkColl2ValNil(b *testing.B) {
	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 1024), 0)
	config.NewValue(nil).Tocollate(clt)

	for i := 0; i < b.N; i++ {
		clt.Tovalue()
	}
}

func BenchmarkColl2ValTrue(b *testing.B) {
	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 1024), 0)

	config.NewValue(interface{}(true)).Tocollate(clt)
	for i := 0; i < b.N; i++ {
		clt.Tovalue()
	}
}

func BenchmarkColl2ValFalse(b *testing.B) {
	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 1024), 0)

	config.NewValue(interface{}(false)).Tocollate(clt)
	for i := 0; i < b.N; i++ {
		clt.Tovalue()
	}
}

func BenchmarkColl2ValF64(b *testing.B) {
	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 1024), 0)
	config.NewValue(float64(10.121312213123123)).Tocollate(clt)

	for i := 0; i < b.N; i++ {
		clt.Tovalue()
	}
}

func BenchmarkColl2ValI64(b *testing.B) {
	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 1024), 0)
	config.NewValue(int64(123456789)).Tocollate(clt)

	for i := 0; i < b.N; i++ {
		clt.Tovalue()
	}
}

func BenchmarkColl2ValMiss(b *testing.B) {
	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 1024), 0)
	config.NewValue(MissingLiteral).Tocollate(clt)

	for i := 0; i < b.N; i++ {
		clt.Tovalue()
	}
}

func BenchmarkColl2ValStr(b *testing.B) {
	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 1024), 0)
	config.NewValue("hello world").Tocollate(clt)

	for i := 0; i < b.N; i++ {
		clt.Tovalue()
	}
}

func BenchmarkColl2ValArr(b *testing.B) {
	arr := []interface{}{nil, true, false, "hello world", 10.23122312}
	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 1024), 0)
	config.NewValue(arr).Tocollate(clt)

	for i := 0; i < b.N; i++ {
		clt.Tovalue()
	}
}

func BenchmarkColl2ValMap(b *testing.B) {
	obj := map[string]interface{}{
		"key1": nil, "key2": true, "key3": false, "key4": "hello world",
		"key5": 10.23122312,
	}
	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 1024), 0)
	config.NewValue(obj).Tocollate(clt)

	for i := 0; i < b.N; i++ {
		clt.Tovalue()
	}
}

func BenchmarkColl2ValTyp(b *testing.B) {
	config := NewDefaultConfig()
	jsn := config.NewJson(testdataFile("testdata/typical.json"), -1)
	clt := config.NewCollate(make([]byte, 10*1024), 0)
	jsn.Tocollate(clt)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		clt.Tovalue()
	}
}
