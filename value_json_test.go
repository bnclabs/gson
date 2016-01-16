//  Copyright (c) 2015 Couchbase, Inc.

package gson

import "compress/gzip"
import "encoding/json"
import "strings"
import "io/ioutil"
import "os"
import "reflect"
import "testing"
import "fmt"

var _ = fmt.Sprintf("dummy")

func TestJsonEmpty2Value(t *testing.T) {
	config := NewDefaultConfig()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic")
		}
	}()
	json2value("", config)
}

func TestScanNull(t *testing.T) {
	config := NewDefaultConfig()

	jsn := config.NewJson(make([]byte, 1024), 0)

	if jsnrem, val := jsn.Reset([]byte("null")).Tovalue(); jsnrem != nil {
		t.Errorf("remaining text after parsing should be empty, %q", jsnrem)
	} else if val != nil {
		t.Errorf("`null` should be parsed to nil")
	}

	// test bad input
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		jsontxt := []byte("nil")
		jsn.Reset(jsontxt)
		jsn.Tovalue()
	}()
}

func TestScanBool(t *testing.T) {
	testcases := []string{"true", "false"}

	var refval interface{}

	config := NewDefaultConfig()
	jsn := config.NewJson(make([]byte, 1024), 0)

	for _, tcase := range testcases {
		json.Unmarshal([]byte(tcase), &refval)
		jsn.Reset([]byte(tcase))

		if jsnrem, val := jsn.Tovalue(); jsnrem != nil {
			t.Errorf("remaining text after parsing should be empty, %q", jsnrem)
		} else if v, ok := val.(bool); !ok || v != refval.(bool) {
			t.Errorf("%q should be parsed to %v", tcase, refval)
		}
	}

	// test bad input
	fn := func(input string) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		jsn.Reset([]byte(input))
		jsn.Tovalue()
	}
	fn("trrr")
	fn("flse")
}

func TestScanIntegers(t *testing.T) {
	testcases := []string{"10", "-10"}

	var ref interface{}

	for _, tcase := range testcases {
		json.Unmarshal([]byte(tcase), &ref)

		config := NewDefaultConfig().NumberKind(IntNumber).SpaceKind(AnsiSpace)
		jsn := config.NewJson(make([]byte, 1024), 0)
		if jsnrem, val := jsn.Reset([]byte(tcase)).Tovalue(); jsnrem != nil {
			t.Errorf("remaining text after parsing should be empty, %q", jsnrem)
		} else if v, ok := val.(int64); !ok || v != int64(ref.(float64)) {
			t.Errorf("%q int should be parsed to %T %v", tcase, val, ref)
		}

		config = NewDefaultConfig().NumberKind(JSONNumber).SpaceKind(AnsiSpace)
		jsn = config.NewJson(make([]byte, 1024), 0)
		if jsnrem, val := jsn.Reset([]byte(tcase)).Tovalue(); jsnrem != nil {
			t.Errorf("remaining text after parsing should be empty, %q", jsnrem)
		} else if v, ok := val.(json.Number); !ok || string(v) != tcase {
			t.Errorf("expected {%T,%v}, got {%T,%v} %v", v, v, tcase, tcase, ok)
		}
	}

	testcases = []string{
		"0.1", "-0.1", "10.1", "-10.1", "-10E-1", "-10e+1", "10E-1", "10e+1",
	}
	for _, tcase := range testcases {
		json.Unmarshal([]byte(tcase), &ref)

		config := NewDefaultConfig().NumberKind(FloatNumber).SpaceKind(AnsiSpace)
		jsn := config.NewJson(make([]byte, 1024), 0)
		_, val := jsn.Reset([]byte(tcase)).Tovalue()
		if v, ok := val.(float64); !ok || v != ref.(float64) {
			t.Errorf("%q int should be parsed to %v", tcase, ref)
		}

		func() {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("expected panic")
				}
			}()
			json.Unmarshal([]byte(tcase), &ref)
			config := NewDefaultConfig().NumberKind(IntNumber).SpaceKind(AnsiSpace)
			jsn := config.NewJson(make([]byte, 1024), 0)
			jsnrem, _ := jsn.Reset([]byte(tcase)).Tovalue()
			if remtxt := string(jsnrem.Bytes()); remtxt == tcase {
				t.Errorf("expected %v got %v", tcase, remtxt)
			}
		}()

		config = NewDefaultConfig().NumberKind(JSONNumber).SpaceKind(AnsiSpace)
		jsn = config.NewJson(make([]byte, 1024), 0)
		if jsnrem, val := jsn.Reset([]byte(tcase)).Tovalue(); jsnrem != nil {
			t.Errorf("remaining text after parsing should be empty, %q", jsnrem)
		} else if v, ok := val.(json.Number); !ok || string(v) != tcase {
			t.Errorf("should be parsed as String-number")
		}
	}
}

func TestScanMalformed(t *testing.T) {
	config := NewDefaultConfig().NumberKind(IntNumber).SpaceKind(AnsiSpace)

	for _, tcase := range scaninvalid {
		func() {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("expected panic")
				}
			}()
			t.Logf("%v", tcase)
			json2value(tcase, config)
		}()
	}
}

func TestScanValues(t *testing.T) {
	testcases := append(scanvalid, []string{
		string(mapValue),
		string(allValueIndent),
		string(allValueCompact),
		string(pallValueIndent),
		string(pallValueCompact),
	}...)

	var ref interface{}
	config := NewDefaultConfig()
	jsn := config.NewJson(make([]byte, 1024*1024), 0)

	for _, tcase := range testcases {
		t.Logf("%v", tcase)
		json.Unmarshal([]byte(tcase), &ref)

		jsn.Reset([]byte(tcase))
		_, val := jsn.Tovalue()
		if reflect.DeepEqual(val, ref) == false {
			t.Errorf("%q should be parsed as: %v, got %v", tcase, ref, val)
		}
	}
}

func TestCodeJSON(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping code.json.gz")
	}

	var ref interface{}
	data := testdataFile("testdata/code.json.gz")
	json.Unmarshal(data, &ref)

	config := NewDefaultConfig()
	jsn := config.NewJson(data, len(data))
	if jsnrem, val := jsn.Tovalue(); jsnrem != nil {
		t.Errorf("remaining text after parsing should be empty, %q", jsnrem)
	} else if reflect.DeepEqual(val, ref) == false {
		t.Errorf("codeJSON parsing failed with reference: %v", ref)
	}
}

func BenchmarkJson2ValFlt(b *testing.B) {
	config := NewDefaultConfig().NumberKind(FloatNumber)
	in := "100000.23"
	jsn := config.NewJson([]byte(in), len(in))

	b.SetBytes(int64(len(in)))
	for i := 0; i < b.N; i++ {
		jsn.Tovalue()
	}
}

func BenchmarkUnmarshalFlt(b *testing.B) {
	var val interface{}

	in := []byte("100000.23")
	b.SetBytes(int64(len(in)))
	for i := 0; i < b.N; i++ {
		json.Unmarshal(in, &val)
	}
}

func BenchmarkJson2ValJsn(b *testing.B) {
	config := NewDefaultConfig().NumberKind(JSONNumber)
	in := "100000.23"
	jsn := config.NewJson([]byte(in), len(in))
	b.SetBytes(int64(len(in)))
	for i := 0; i < b.N; i++ {
		jsn.Tovalue()
	}
}

func BenchmarkUnmarshalNum(b *testing.B) {
	var val interface{}

	in := []byte("100000.23")
	b.SetBytes(int64(len(in)))
	for i := 0; i < b.N; i++ {
		json.Unmarshal(in, &val)
	}
}

func BenchmarkJson2ValString(b *testing.B) {
	config := NewDefaultConfig()
	in := `"汉语 / 漢語; Hàn\b \tyǔ "`
	jsn := config.NewJson([]byte(in), len(in))
	b.SetBytes(int64(len(in)))
	for i := 0; i < b.N; i++ {
		jsn.Tovalue()
	}
}

func BenchmarkUnmarshalStr(b *testing.B) {
	var val interface{}

	in := []byte(`"汉语 / 漢語; Hàn\b \tyǔ "`)
	b.SetBytes(int64(len(in)))
	for i := 0; i < b.N; i++ {
		json.Unmarshal(in, &val)
	}
}

func BenchmarkJson2ValArr5(b *testing.B) {
	in := ` [null,true,false,10,"tru\"e"]`
	config := NewDefaultConfig().NumberKind(FloatNumber).SpaceKind(AnsiSpace)
	jsn := config.NewJson([]byte(in), len(in))
	b.SetBytes(int64(len(in)))
	for i := 0; i < b.N; i++ {
		jsn.Tovalue()
	}
}

func BenchmarkUnmarshalArr5(b *testing.B) {
	var a []interface{}

	in := []byte(` [null,true,false,10,"tru\"e"]`)
	b.SetBytes(int64(len(in)))
	for i := 0; i < b.N; i++ {
		json.Unmarshal(in, &a)
	}
}

func BenchmarkJson2ValMap5(b *testing.B) {
	in := `{"a": null, "b" : true,"c":false, "d\"":-10E-1, "e":"tru\"e" }`
	config := NewDefaultConfig().NumberKind(FloatNumber).SpaceKind(AnsiSpace)
	jsn := config.NewJson([]byte(in), len(in))
	b.SetBytes(int64(len(in)))
	for i := 0; i < b.N; i++ {
		jsn.Tovalue()
	}
}

func BenchmarkUnmarshalMap5(b *testing.B) {
	var m map[string]interface{}

	in := []byte(`{"a": null, "b" : true,"c":false, "d\"":-10E-1, "e":"tru\"e" }`)
	b.SetBytes(int64(len(in)))
	for i := 0; i < b.N; i++ {
		json.Unmarshal(in, &m)
	}
}

func BenchmarkJson2ValTyp(b *testing.B) {
	in := string(testdataFile("testdata/typical.json"))

	config := NewDefaultConfig().NumberKind(FloatNumber).SpaceKind(AnsiSpace)
	jsn := config.NewJson([]byte(in), len(in))
	b.SetBytes(int64(len(in)))
	for i := 0; i < b.N; i++ {
		jsn.Tovalue()
	}
}

func BenchmarkUnmarshalTyp(b *testing.B) {
	var m map[string]interface{}

	data := testdataFile("testdata/typical.json")
	b.SetBytes(int64(len(data)))
	for i := 0; i < b.N; i++ {
		json.Unmarshal(data, &m)
	}
}

func BenchmarkJson2ValCgz(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping code.json.gz")
	}

	data := testdataFile("testdata/code.json.gz")
	config := NewDefaultConfig().NumberKind(FloatNumber).SpaceKind(AnsiSpace)
	jsn := config.NewJson(data, len(data))
	b.SetBytes(int64(len(data)))
	for i := 0; i < b.N; i++ {
		jsn.Tovalue()
	}
}

func BenchmarkUnmarshalCgz(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping code.json.gz")
	}

	var m map[string]interface{}

	data := testdataFile("testdata/code.json.gz")
	b.SetBytes(int64(len(data)))
	for i := 0; i < b.N; i++ {
		json.Unmarshal(data, &m)
	}
}

func BenchmarkVal2JsonFlt(b *testing.B) {
	config := NewDefaultConfig().NumberKind(FloatNumber)
	jsn := config.NewJson(make([]byte, 1024), 0)
	val := config.NewValue(100000.23)
	for i := 0; i < b.N; i++ {
		jsn.Reset(nil)
		val.Tojson(jsn)
	}
	b.SetBytes(int64(len(jsn.Bytes())))
}

func BenchmarkVal2JsonString(b *testing.B) {
	config := NewDefaultConfig()
	jsn := config.NewJson(make([]byte, 1024), 0)
	val := config.NewValue(`"汉语 / 漢語; Hàn\b \tyǔ "`)
	for i := 0; i < b.N; i++ {
		jsn.Reset(nil)
		val.Tojson(jsn)
	}
	b.SetBytes(int64(len(jsn.Bytes())))
}

func BenchmarkVal2JsonArr5(b *testing.B) {
	config := NewDefaultConfig().NumberKind(FloatNumber).SpaceKind(AnsiSpace)
	jsn := config.NewJson(make([]byte, 1024), 0)
	val := config.NewValue([]interface{}{nil, true, false, 10, "tru\"e"})
	for i := 0; i < b.N; i++ {
		jsn.Reset(nil)
		val.Tojson(jsn)
	}
	b.SetBytes(int64(len(jsn.Bytes())))
}

func BenchmarkVal2JsonMap5(b *testing.B) {
	config := NewDefaultConfig().NumberKind(FloatNumber).SpaceKind(AnsiSpace)
	jsn := config.NewJson(make([]byte, 1024), 0)
	val := config.NewValue(map[string]interface{}{
		"a": nil, "b": true, "c": false, "d\"": -10E-1, "e": "tru\"e",
	})
	for i := 0; i < b.N; i++ {
		jsn.Reset(nil)
		val.Tojson(jsn)
	}
	b.SetBytes(int64(len(jsn.Bytes())))
}

func BenchmarkVal2JsonTyp(b *testing.B) {
	data := testdataFile("testdata/typical.json")
	config := NewDefaultConfig().NumberKind(FloatNumber).SpaceKind(AnsiSpace)
	jsn := config.NewJson(data, len(data))
	_, v := jsn.Tovalue()
	val := config.NewValue(v)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		jsn.Reset(nil)
		val.Tojson(jsn)
	}
	b.SetBytes(int64(len(jsn.Bytes())))
}

func BenchmarkVal2JsonCgz(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping code.json.gz")
	}

	data := testdataFile("testdata/code.json.gz")
	config := NewDefaultConfig().NumberKind(FloatNumber).SpaceKind(AnsiSpace)
	jsn := config.NewJson(data, len(data))
	_, v := jsn.Tovalue()
	val := config.NewValue(v)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		jsn.Reset(nil)
		val.Tojson(jsn)
	}
	b.SetBytes(int64(len(jsn.Bytes())))
}

func testdataFile(filename string) []byte {
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	var data []byte
	if strings.HasSuffix(filename, ".gz") {
		gz, err := gzip.NewReader(f)
		if err != nil {
			panic(err)
		}
		data, err = ioutil.ReadAll(gz)
		if err != nil {
			panic(err)
		}
	} else {
		data, err = ioutil.ReadAll(f)
		if err != nil {
			panic(err)
		}
	}
	return data
}

var allValueIndent, allValueCompact, pallValueIndent, pallValueCompact []byte
var mapValue []byte
var scanvalid []string
var scaninvalid []string

func init() {
	var value interface{}
	var err error

	allValueIndent, err = ioutil.ReadFile("testdata/allValueIndent")
	if err != nil {
		panic(err)
	}

	if err = json.Unmarshal(allValueIndent, &value); err != nil {
		panic(err)
	}
	if allValueCompact, err = json.Marshal(value); err != nil {
		panic(err)
	}

	pallValueIndent, err = ioutil.ReadFile("testdata/pallValueIndent")
	if err != nil {
		panic(err)
	}

	if err = json.Unmarshal(pallValueIndent, &value); err != nil {
		panic(err)
	}
	if pallValueCompact, err = json.Marshal(value); err != nil {
		panic(err)
	}

	mapValue, err = ioutil.ReadFile("testdata/map")
	if err != nil {
		panic(err)
	}

	scanvalidb, err := ioutil.ReadFile("testdata/scan_valid")
	if err != nil {
		panic(err)
	}
	scanvalid = []string{}
	for _, s := range strings.Split(string(scanvalidb), "\n") {
		if strings.Trim(s, " ") != "" {
			scanvalid = append(scanvalid, s)
		}
	}
	scanvalid = append(scanvalid, []string{
		"\"hello\xffworld\"",
		"\"hello\xc2\xc2world\"",
		"\"hello\xc2\xffworld\"",
		"\"hello\xed\xa0\x80\xed\xb0\x80world\""}...)

	scaninvalidb, err := ioutil.ReadFile("testdata/scan_invalid")
	if err != nil {
		panic(err)
	}
	scaninvalid = []string{}
	for _, s := range strings.Split(string(scaninvalidb), "\n") {
		if strings.Trim(s, " ") != "" {
			scaninvalid = append(scaninvalid, s)
		}
	}
	scaninvalid = append(scaninvalid, []string{
		"\xed\xa0\x80", // RuneError
		"\xed\xbf\xbf", // RuneError
		// raw value errors
		"\x01 42",
		"\x01 true",
		"\x01 1.2",
		" 3.4 \x01",
		"\x01 \"string\"",
		// bad-utf8
		"hello\xffworld",
		"\xff",
		"\xff\xff",
		"a\xffb",
		"\xe6\x97\xa5\xe6\x9c\xac\xff\xaa\x9e"}...)
}
