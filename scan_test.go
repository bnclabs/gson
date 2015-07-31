package gson

import "compress/gzip"
import "encoding/json"
import "io/ioutil"
import "os"
import "strings"
import "reflect"
import "testing"
import "fmt"

var _ = fmt.Sprintf("dummy")

func TestScan(t *testing.T) {
	var ref interface{}

	testcases := append(scan_valid, []string{
		string(mapValue),
		string(allValueIndent),
		string(allValueCompact),
		string(allValueIndent),
		string(allValueCompact),
		string(pallValueIndent),
		string(pallValueCompact),
		string(pallValueIndent),
		string(pallValueCompact),
	}...)

	config := NewDefaultConfig()
	for _, tcase := range testcases {
		t.Logf("%v", tcase)
		if err := json.Unmarshal([]byte(tcase), &ref); err != nil {
			t.Errorf("error parsing i/p %q: %v", tcase, err)
		}
		if val, _ := config.Parse(tcase); reflect.DeepEqual(val, ref) == false {
			t.Errorf("%q should be parsed as: %v, got %v", tcase, ref, val)
		}
	}
}

func TestScanNull(t *testing.T) {
	config := NewDefaultConfig()
	if val, remtxt := config.Parse("null"); remtxt != "" {
		t.Errorf("remaining text after parsing should be empty, %q", remtxt)
	} else if val != nil {
		t.Errorf("`null` should be parsed to nil")
	}
}

func TestScanBool(t *testing.T) {
	var refval interface{}

	tests := []string{"true", "false"}
	config := NewDefaultConfig()
	for _, test := range tests {
		if err := json.Unmarshal([]byte(test), &refval); err != nil {
			t.Fatalf("error parsing i/p %q: %v", test, err)
		}
		if val, remtxt := config.Parse(test); remtxt != "" {
			t.Errorf("remaining text after parsing should be empty, %q", remtxt)
		} else if v, ok := val.(bool); !ok || v != refval.(bool) {
			t.Errorf("%q should be parsed to %v", test, refval)
		}
	}
}

func TestScanIntegers(t *testing.T) {
	var ref interface{}

	ints := []string{"10", "-10"}
	for _, test := range ints {
		config := NewConfig(IntNumber, AnsiSpace)
		if err := json.Unmarshal([]byte(test), &ref); err != nil {
			t.Fatalf("error parsing i/p %q: %v", test, err)
		}
		if val, remtxt := config.Parse(test); remtxt != "" {
			t.Errorf("remaining text after parsing should be empty, %q", remtxt)
		} else if v, ok := val.(int); !ok || v != int(ref.(float64)) {
			t.Errorf("%q int should be parsed to %T %v", test, val, ref)
		}

		config = NewConfig(StringNumber, AnsiSpace)
		if val, remtxt := config.Parse(test); remtxt != "" {
			t.Errorf("remaining text after parsing should be empty, %q", remtxt)
		} else if v, ok := val.(Number); !ok || string(v) != test {
			t.Errorf("expected {%T,%v}, got {%T,%v} %v", v, v, test, test, ok)
		}
	}

	floats := []string{"0.1", "-0.1", "10.1", "-10.1", "-10E-1",
		"-10e+1", "10E-1", "10e+1"}
	for _, test := range floats {
		config := NewConfig(FloatNumber, AnsiSpace)
		if err := json.Unmarshal([]byte(test), &ref); err != nil {
			t.Fatalf("error parsing i/p %q: %v", test, err)
		}
		val, _ := config.Parse(test)
		if v, ok := val.(float64); !ok || v != ref.(float64) {
			t.Errorf("%q int should be parsed to %v", test, ref)
		}

		func() {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("expected panic")
				}
			}()
			config = NewConfig(IntNumber, AnsiSpace)
			if err := json.Unmarshal([]byte(test), &ref); err != nil {
				t.Fatalf("error parsing i/p %q: %v", test, err)
			}
			if _, remtxt := config.Parse(test); remtxt == test {
				t.Errorf("expected %v got %v", test, remtxt)
			}
		}()

		config = NewConfig(StringNumber, AnsiSpace)
		if val, remtxt := config.Parse(test); remtxt != "" {
			t.Errorf("remaining text after parsing should be empty, %q", remtxt)
		} else if v, ok := val.(Number); !ok || string(v) != test {
			t.Errorf("%q should be parsed as String-number")
		}
	}
}

func TestScanEmpty(t *testing.T) {
	config := NewDefaultConfig()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic")
		}
	}()
	scanToken("", config)
}

func TestScanMalformed(t *testing.T) {
	config := NewConfig(IntNumber, AnsiSpace)
	for _, tcase := range scan_invalid {
		func() {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("expected panic")
				}
			}()
			t.Logf("%v", tcase)
			scanToken(tcase, config)
		}()
	}
}

func TestCodeJSON(t *testing.T) {
	var ref interface{}

	data := codeJSON()
	config := NewDefaultConfig()
	if err := json.Unmarshal(data, &ref); err != nil {
		t.Errorf("error parsing codeJSON: %v", err)
	}
	if val, remtxt := config.Parse(string(data)); remtxt != "" {
		t.Errorf("remaining text after parsing should be empty, %q", remtxt)
	} else if reflect.DeepEqual(val, ref) == false {
		t.Error("codeJSON parsing failed with reference: %v", ref)
	}
}

func BenchmarkScanNumFloat(b *testing.B) {
	in := "100000.23"
	for i := 0; i < b.N; i++ {
		scanNum(in, FloatNumber)
	}
}

func BenchmarkScanNumString(b *testing.B) {
	in := "100000.23"
	for i := 0; i < b.N; i++ {
		scanNum(in, StringNumber)
	}
}

func BenchmarkScanString(b *testing.B) {
	in := []byte(`"汉语 / 漢語; Hàn\b \tyǔ "`)
	for i := 0; i < b.N; i++ {
		scanString(in)
	}
}

//func BenchmarkSmallJSONPkg(b *testing.B) {
//    txt := []byte(`{"a": null, "b" : true,"c":false, "d\"":-10E-1, "e":"tru\"e" }`)
//    p := NewParser(FloatNumber, AnsiSpace, false /*jsonp*/)
//    b.SetBytes(int64(len(txt)))
//    for i := 0; i < b.N; i++ {
//        if _, _, err := p.Parse(txt); err != nil {
//            b.Fatal(err)
//        }
//    }
//}
//
//func BenchmarkSmallJSON(b *testing.B) {
//    var m map[string]interface{}
//    txt := []byte(`{"a": null, "b" : true,"c":false, "d\"":-10E-1, "e":"tru\"e" }`)
//    b.SetBytes(int64(len(txt)))
//    for i := 0; i < b.N; i++ {
//        if err := json.Unmarshal(txt, &m); err != nil {
//            b.Fatal(err)
//        }
//    }
//}
//
//func BenchmarkCodeJSONPkg(b *testing.B) {
//    p := NewParser(FloatNumber, AnsiSpace, false /*jsonp*/)
//    b.SetBytes(int64(len(codeJSON)))
//    for i := 0; i < b.N; i++ {
//        if _, _, err := p.Parse(codeJSON); err != nil {
//            b.Fatal(err)
//        }
//    }
//}
//
//func BenchmarkCodeJSON(b *testing.B) {
//    var m map[string]interface{}
//    b.SetBytes(int64(len(codeJSON)))
//    for i := 0; i < b.N; i++ {
//        if err := json.Unmarshal(codeJSON, &m); err != nil {
//            b.Fatal(err)
//        }
//    }
//}
//
func codeJSON() []byte {
	f, err := os.Open("testdata/code.json.gz")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		panic(err)
	}
	data, err := ioutil.ReadAll(gz)
	if err != nil {
		panic(err)
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
