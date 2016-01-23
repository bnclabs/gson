//  Copyright (c) 2015 Couchbase, Inc.

package gson

import "encoding/json"
import "reflect"
import "testing"
import "fmt"

var _ = fmt.Sprintf("dummy")

func TestNil2Json(t *testing.T) {
	config := NewDefaultConfig()
	jsn := config.NewJson(make([]byte, 1024), 0)

	config.NewValue(nil).Tojson(jsn)
	if s := string(jsn.Bytes()); s != "null" {
		t.Errorf("expected %q, got %q", "null", s)
	}
}

func TestBool2Json(t *testing.T) {
	config := NewDefaultConfig()
	jsn := config.NewJson(make([]byte, 1024), 0)

	config.NewValue(true).Tojson(jsn)
	if s := string(jsn.Bytes()); s != "true" {
		t.Errorf("expected %q, got %q", "true", s)
	}

	config.NewValue(false).Tojson(jsn.Reset(nil))
	if s := string(jsn.Bytes()); s != "false" {
		t.Errorf("expected %q, got %q", "false")
	}
}

func TestNumber2Json(t *testing.T) {
	testcases := []interface{}{
		10.0, -10.0, 0.1, -0.1, 10.1, -10.1, -10E-1, -10e+1, 10E-1, 10e+1,
	}
	config := NewDefaultConfig()
	jsn := config.NewJson(make([]byte, 1024), 0)

	for _, tcase := range testcases {
		var out interface{}

		config.NewValue(tcase).Tojson(jsn.Reset(nil))
		if json.Unmarshal(jsn.Bytes(), &out); !reflect.DeepEqual(out, tcase) {
			t.Errorf("expected %v, got %v", tcase, out)
		}
	}
}

func TestValues2Json(t *testing.T) {
	testcases := append(scanvalid, []string{
		string(mapValue),
		string(allValueIndent),
		string(allValueCompact),
		string(pallValueIndent),
		string(pallValueCompact),
	}...)

	config := NewDefaultConfig()
	jsn := config.NewJson(make([]byte, 1024*1024), 0)

	for _, tcase := range testcases {
		var value interface{}

		t.Logf("%v", tcase)
		json.Unmarshal([]byte(tcase), &value)
		config.NewValue(value).Tojson(jsn.Reset(nil))

		_, outval := jsn.Tovalue()
		if reflect.DeepEqual(outval, value) == false {
			t.Errorf("expected %v, got %v", value, outval)
		}
	}
}

func TestCodeVal2Json(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping code.json.gz")
	}

	var value interface{}

	data := testdataFile("testdata/code.json.gz")

	config := NewDefaultConfig()
	jsn := config.NewJson(make([]byte, len(data)*2), 0)
	json.Unmarshal(data, &value)
	jsnrem, outval := config.NewValue(value).Tojson(jsn).Tovalue()

	if jsnrem != nil {
		t.Errorf("remaining text after parsing should be empty, %q", jsnrem)
	} else if reflect.DeepEqual(outval, value) == false {
		t.Errorf("codeJSON expected %v, got %v", value, outval)
	}
}

func BenchmarkVal2JsonNil(b *testing.B) {
	config := NewDefaultConfig()
	config = config.SetNumberKind(FloatNumber)
	jsn := config.NewJson(make([]byte, 1024), 0)
	val := config.NewValue(nil)

	for i := 0; i < b.N; i++ {
		val.Tojson(jsn.Reset(nil))
	}
	b.SetBytes(int64(len(jsn.Bytes())))
}

func BenchmarkVal2JsonBool(b *testing.B) {
	config := NewDefaultConfig()
	config = config.SetNumberKind(FloatNumber)
	jsn := config.NewJson(make([]byte, 1024), 0)
	val := config.NewValue(true)

	for i := 0; i < b.N; i++ {
		val.Tojson(jsn.Reset(nil))
	}
	b.SetBytes(int64(len(jsn.Bytes())))
}

func BenchmarkVal2JsonNum(b *testing.B) {
	config := NewDefaultConfig()
	config = config.SetNumberKind(FloatNumber)
	jsn := config.NewJson(make([]byte, 1024), 0)
	val := config.NewValue(100000.23)

	for i := 0; i < b.N; i++ {
		val.Tojson(jsn.Reset(nil))
	}
	b.SetBytes(int64(len(jsn.Bytes())))
}

func BenchmarkVal2JsonString(b *testing.B) {
	config := NewDefaultConfig()
	jsn := config.NewJson(make([]byte, 1024), 0)
	val := config.NewValue(`"汉语 / 漢語; Hàn\b \tyǔ "`)

	for i := 0; i < b.N; i++ {
		val.Tojson(jsn.Reset(nil))
	}
	b.SetBytes(int64(len(jsn.Bytes())))
}

func BenchmarkVal2JsonArr5(b *testing.B) {
	config := NewDefaultConfig()
	config = config.SetNumberKind(FloatNumber).SetSpaceKind(AnsiSpace)
	jsn := config.NewJson(make([]byte, 1024), 0)
	val := config.NewValue([]interface{}{nil, true, false, 10, "tru\"e"})

	for i := 0; i < b.N; i++ {
		val.Tojson(jsn.Reset(nil))
	}
	b.SetBytes(int64(len(jsn.Bytes())))
}

func BenchmarkVal2JsonMap5(b *testing.B) {
	config := NewDefaultConfig()
	config = config.SetNumberKind(FloatNumber).SetSpaceKind(AnsiSpace)
	jsn := config.NewJson(make([]byte, 1024), 0)
	val := config.NewValue(map[string]interface{}{
		"a": nil, "b": true, "c": false, "d\"": -10E-1, "e": "tru\"e",
	})

	for i := 0; i < b.N; i++ {
		val.Tojson(jsn.Reset(nil))
	}
	b.SetBytes(int64(len(jsn.Bytes())))
}

func BenchmarkVal2JsonTyp(b *testing.B) {
	var ref interface{}

	data := testdataFile("testdata/typical.json")
	if err := json.Unmarshal(data, &ref); err != nil {
		b.Fatal(err)
	}

	config := NewDefaultConfig()
	config = config.SetNumberKind(FloatNumber).SetSpaceKind(AnsiSpace)
	jsn := config.NewJson(make([]byte, len(data)*2), 0)
	val := config.NewValue(ref)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		val.Tojson(jsn.Reset(nil))
	}
	b.SetBytes(int64(len(jsn.Bytes())))
}

func BenchmarkVal2JsonCgz(b *testing.B) {
	var ref interface{}

	if testing.Short() {
		b.Skip("skipping code.json.gz")
	}

	data := testdataFile("testdata/code.json.gz")
	if err := json.Unmarshal(data, &ref); err != nil {
		b.Fatal(err)
	}
	config := NewDefaultConfig()
	config = config.SetNumberKind(FloatNumber).SetSpaceKind(AnsiSpace)
	jsn := config.NewJson(make([]byte, len(data)*2), 0)
	val := config.NewValue(ref)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		val.Tojson(jsn.Reset(nil))
	}
	b.SetBytes(int64(len(jsn.Bytes())))
}
