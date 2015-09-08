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
	if remtxt, val := config.JsonToValue("null"); remtxt != "" {
		t.Errorf("remaining text after parsing should be empty, %q", remtxt)
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
		config.JsonToValue("nil")
	}()
}

func TestScanBool(t *testing.T) {
	var refval interface{}

	tests := []string{"true", "false"}
	config := NewDefaultConfig()
	for _, test := range tests {
		if err := json.Unmarshal([]byte(test), &refval); err != nil {
			t.Fatalf("error parsing i/p %q: %v", test, err)
		}
		if remtxt, val := config.JsonToValue(test); remtxt != "" {
			t.Errorf("remaining text after parsing should be empty, %q", remtxt)
		} else if v, ok := val.(bool); !ok || v != refval.(bool) {
			t.Errorf("%q should be parsed to %v", test, refval)
		}
	}
	// test bad input
	fn := func(input string) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		config.JsonToValue(input)
	}
	fn("trrr")
	fn("flse")
}

func TestScanIntegers(t *testing.T) {
	var ref interface{}

	ints := []string{"10", "-10"}
	for _, test := range ints {
		config := NewConfig(IntNumber, AnsiSpace, Stream)
		if err := json.Unmarshal([]byte(test), &ref); err != nil {
			t.Fatalf("error parsing i/p %q: %v", test, err)
		}
		if remtxt, val := config.JsonToValue(test); remtxt != "" {
			t.Errorf("remaining text after parsing should be empty, %q", remtxt)
		} else if v, ok := val.(int); !ok || v != int(ref.(float64)) {
			t.Errorf("%q int should be parsed to %T %v", test, val, ref)
		}

		config = NewConfig(JsonNumber, AnsiSpace, Stream)
		if remtxt, val := config.JsonToValue(test); remtxt != "" {
			t.Errorf("remaining text after parsing should be empty, %q", remtxt)
		} else if v, ok := val.(json.Number); !ok || string(v) != test {
			t.Errorf("expected {%T,%v}, got {%T,%v} %v", v, v, test, test, ok)
		}
	}

	floats := []string{"0.1", "-0.1", "10.1", "-10.1", "-10E-1",
		"-10e+1", "10E-1", "10e+1"}
	for _, test := range floats {
		config := NewConfig(FloatNumber, AnsiSpace, Stream)
		if err := json.Unmarshal([]byte(test), &ref); err != nil {
			t.Fatalf("error parsing i/p %q: %v", test, err)
		}
		_, val := config.JsonToValue(test)
		if v, ok := val.(float64); !ok || v != ref.(float64) {
			t.Errorf("%q int should be parsed to %v", test, ref)
		}

		func() {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("expected panic")
				}
			}()
			config = NewConfig(IntNumber, AnsiSpace, Stream)
			if err := json.Unmarshal([]byte(test), &ref); err != nil {
				t.Fatalf("error parsing i/p %q: %v", test, err)
			}
			if remtxt, _ := config.JsonToValue(test); remtxt == test {
				t.Errorf("expected %v got %v", test, remtxt)
			}
		}()

		config = NewConfig(JsonNumber, AnsiSpace, Stream)
		if remtxt, val := config.JsonToValue(test); remtxt != "" {
			t.Errorf("remaining text after parsing should be empty, %q", remtxt)
		} else if v, ok := val.(json.Number); !ok || string(v) != test {
			t.Errorf("%q should be parsed as String-number")
		}
	}
}

func TestScanMalformed(t *testing.T) {
	config := NewConfig(IntNumber, AnsiSpace, Stream)
	for _, tcase := range scan_invalid {
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
	var ref interface{}

	testcases := append(scan_valid, []string{
		string(mapValue),
		string(allValueIndent),
		string(allValueCompact),
		string(pallValueIndent),
		string(pallValueCompact),
	}...)

	config := NewDefaultConfig()
	for _, tcase := range testcases {
		t.Logf("%v", tcase)
		if err := json.Unmarshal([]byte(tcase), &ref); err != nil {
			t.Errorf("error parsing i/p %q: %v", tcase, err)
		}
		_, val := config.JsonToValue(tcase)
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
	config := NewDefaultConfig()
	if err := json.Unmarshal(data, &ref); err != nil {
		t.Errorf("error parsing codeJSON: %v", err)
	}
	if remtxt, val := config.JsonToValue(string(data)); remtxt != "" {
		t.Errorf("remaining text after parsing should be empty, %q", remtxt)
	} else if reflect.DeepEqual(val, ref) == false {
		t.Error("codeJSON parsing failed with reference: %v", ref)
	}
}

func BenchmarkScanNumFlt(b *testing.B) {
	in := "100000.23"
	b.SetBytes(int64(len(in)))
	for i := 0; i < b.N; i++ {
		scanNum(in, FloatNumber)
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

func BenchmarkScanNumJsonNum(b *testing.B) {
	in := "100000.23"
	b.SetBytes(int64(len(in)))
	for i := 0; i < b.N; i++ {
		scanNum(in, JsonNumber)
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

func BenchmarkScanString(b *testing.B) {
	in := `"汉语 / 漢語; Hàn\b \tyǔ "`
	scratch := make([]byte, 1024)
	b.SetBytes(int64(len(in)))
	for i := 0; i < b.N; i++ {
		scanString(in, scratch)
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
	txt := ` [null,true,false,10,"tru\"e"]`
	config := NewConfig(FloatNumber, AnsiSpace, Stream)
	b.SetBytes(int64(len(txt)))
	for i := 0; i < b.N; i++ {
		config.JsonToValue(txt)
	}
}

func BenchmarkUnmarshalArr5(b *testing.B) {
	var a []interface{}
	txt := []byte(` [null,true,false,10,"tru\"e"]`)
	b.SetBytes(int64(len(txt)))
	for i := 0; i < b.N; i++ {
		json.Unmarshal(txt, &a)
	}
}

func BenchmarkJson2ValMap5(b *testing.B) {
	txt := `{"a": null, "b" : true,"c":false, "d\"":-10E-1, "e":"tru\"e" }`
	config := NewConfig(FloatNumber, AnsiSpace, Stream)
	b.SetBytes(int64(len(txt)))
	for i := 0; i < b.N; i++ {
		config.JsonToValue(txt)
	}
}

func BenchmarkUnmarshalMap5(b *testing.B) {
	var m map[string]interface{}
	txt := []byte(`{"a": null, "b" : true,"c":false, "d\"":-10E-1, "e":"tru\"e" }`)
	b.SetBytes(int64(len(txt)))
	for i := 0; i < b.N; i++ {
		json.Unmarshal(txt, &m)
	}
}

func BenchmarkJson2ValTyp(b *testing.B) {
	txt := string(testdataFile("testdata/typical.json"))
	config := NewConfig(FloatNumber, AnsiSpace, Stream)
	b.SetBytes(int64(len(txt)))
	for i := 0; i < b.N; i++ {
		config.JsonToValue(txt)
	}
}

func BenchmarkUnmarshalTyp(b *testing.B) {
	var m map[string]interface{}
	txt := testdataFile("testdata/typical.json")
	b.SetBytes(int64(len(txt)))
	for i := 0; i < b.N; i++ {
		json.Unmarshal(txt, &m)
	}
}

func BenchmarkJson2ValCgz(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping code.json.gz")
	}

	txt := string(testdataFile("testdata/code.json.gz"))
	config := NewConfig(FloatNumber, AnsiSpace, Stream)
	b.SetBytes(int64(len(txt)))
	for i := 0; i < b.N; i++ {
		config.JsonToValue(txt)
	}
}

func BenchmarkUnmarshalCgz(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping code.json.gz")
	}

	var m map[string]interface{}
	txt := testdataFile("testdata/code.json.gz")
	b.SetBytes(int64(len(txt)))
	for i := 0; i < b.N; i++ {
		json.Unmarshal(txt, &m)
	}
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
var scan_valid []string
var scan_invalid []string

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

	scan_valid_b, err := ioutil.ReadFile("testdata/scan_valid")
	if err != nil {
		panic(err)
	}
	scan_valid = []string{}
	for _, s := range strings.Split(string(scan_valid_b), "\n") {
		if strings.Trim(s, " ") != "" {
			scan_valid = append(scan_valid, s)
		}
	}
	scan_valid = append(scan_valid, []string{
		"\"hello\xffworld\"",
		"\"hello\xc2\xc2world\"",
		"\"hello\xc2\xffworld\"",
		"\"hello\xed\xa0\x80\xed\xb0\x80world\""}...)

	scan_invalid_b, err := ioutil.ReadFile("testdata/scan_invalid")
	if err != nil {
		panic(err)
	}
	scan_invalid = []string{}
	for _, s := range strings.Split(string(scan_invalid_b), "\n") {
		if strings.Trim(s, " ") != "" {
			scan_invalid = append(scan_invalid, s)
		}
	}
	scan_invalid = append(scan_invalid, []string{
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
