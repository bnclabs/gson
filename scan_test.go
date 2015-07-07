package gson

import "compress/gzip"
import "encoding/json"
import "io/ioutil"
import "os"
import "reflect"
import "testing"

var testcases = []string{
	// null
	"null",
	// boolean
	"true",
	"false",
	// integers
	"10",
	"0.1",
	"-0.1",
	"10.1",
	"-10.1",
	"-10E-1",
	"-10e+1",
	"10E-1",
	"10e+1",
	// string
	`"true"`,
	`"tru\"e"`,
	`"tru\\e"`,
	`"tru\be"`,
	`"tru\fe"`,
	`"tru\ne"`,
	`"tru\re"`,
	`"tru\te"`,
	`"tru\u0123e"`,
	`"汉语 / 漢語; Hàn\b \t\uef24yǔ "`,
	// array
	` [  ] `,
	`[]`,
	` [ null, true, false, 10, "tru\"e" ] `,
	// object
	` { "a": null, "b" : true,"c":false, "d\"":10, "e":"tru\"e" } `,
	` {  } `,
	`{}`,
}

func TestScanNull(t *testing.T) {
	config := NewDefaultConfig()
	if val, remtxt, err := config.Parse("null"); err != nil {
		t.Errorf("error parsing `null`: %v", err)
	} else if remtxt != "" {
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
		if val, remtxt, err := config.Parse(test); err != nil {
			t.Errorf("error parsing %q %v", test, err)
		} else if remtxt != "" {
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
		if val, remtxt, err := config.Parse(test); err != nil {
			t.Errorf("error parsing %q: %v", test, err)
		} else if remtxt != "" {
			t.Errorf("remaining text after parsing should be empty, %q", remtxt)
		} else if v, ok := val.(int); !ok || v != int(ref.(float64)) {
			t.Errorf("%q int should be parsed to %T %v", test, val, ref)
		}

		config = NewConfig(StringNumber, AnsiSpace)
		if val, remtxt, err := config.Parse(test); err != nil {
			t.Errorf("error parsing %q: %v", test, err)
		} else if remtxt != "" {
			t.Errorf("remaining text after parsing should be empty, %q", remtxt)
		} else if v, ok := val.(Number); !ok || string(v) != test {
			t.Errorf("%q should be parsed as String-number")
		}
	}

	floats := []string{"0.1", "-0.1", "10.1", "-10.1", "-10E-1",
		"-10e+1", "10E-1", "10e+1"}
	for _, test := range floats {
		config := NewConfig(FloatNumber, AnsiSpace)
		if err := json.Unmarshal([]byte(test), &ref); err != nil {
			t.Fatalf("error parsing i/p %q: %v", test, err)
		}
		if val, remtxt, err := config.Parse(test); err != nil {
			t.Errorf("error parsing %q: %v", test, err)
		} else if remtxt != "" {
			t.Errorf("remaining text after parsing should be empty, %q", remtxt)
		} else if v, ok := val.(float64); !ok || v != ref.(float64) {
			t.Errorf("%q int should be parsed to %v", test, ref)
		}

		config = NewConfig(IntNumber, AnsiSpace)
		if err := json.Unmarshal([]byte(test), &ref); err != nil {
			t.Fatalf("error parsing i/p %q: %v", test, err)
		}
		if _, remtxt, err := config.Parse(test); err == nil {
			t.Errorf("expected error parsing float %q", test)
		} else if remtxt == test {
			t.Errorf("expected i/p returned as remaining text")
		}

		config = NewConfig(StringNumber, AnsiSpace)
		if val, remtxt, err := config.Parse(test); err != nil {
			t.Errorf("error parsing %q: %v", test, err)
		} else if remtxt != "" {
			t.Errorf("remaining text after parsing should be empty, %q", remtxt)
		} else if v, ok := val.(Number); !ok || string(v) != test {
			t.Errorf("%q should be parsed as String-number")
		}
	}
}

func TestScan(t *testing.T) {
	var ref interface{}

	config := NewDefaultConfig()
	for _, tcase := range testcases {
		if err := json.Unmarshal([]byte(tcase), &ref); err != nil {
			t.Errorf("error parsing i/p %q: %v", tcase, err)
		}
		if val, _, err := config.Parse(tcase); err != nil {
			t.Errorf("error parsing %q: %v", tcase, err)
		} else if reflect.DeepEqual(val, ref) == false {
			t.Errorf("%q should be parsed as: %v, got %v", tcase, ref, val)
		}
	}
}

func TestScanEmpty(t *testing.T) {
	config := NewDefaultConfig()
	if _, _, err := scanToken("", config); err != ErrorEmptyText {
		t.Errorf("fail expecting ErrorEmptyText: %v", err)
	}
}

func TestScanMalformed(t *testing.T) {
	config := NewDefaultConfig()
	// malformed true
	if _, _, err := scanToken("trre", config); err != ErrorExpectedTrue {
		t.Errorf("fail expecting ErrorExpectedTrue: %v", err)
	}
	// malformed false
	if _, _, err := scanToken("fllse", config); err != ErrorExpectedFalse {
		t.Errorf("fail expecting ErrorExpectedFalse: %v", err)
	}
	// malformed null
	if _, _, err := scanToken("nill", config); err != ErrorExpectedNil {
		t.Errorf("fail expecting ErrorExpectedNil: %v", err)
	}
	// malformed array
	if _, _, err := scanToken("[nill]", config); err != ErrorExpectedNil {
		t.Errorf("fail expecting ErrorExpectedNil in array: %v", err)
	}
	if _, _, err := scanToken("[10", config); err != ErrorExpectedClosearray {
		t.Errorf("fail expecting ErrorExpectedClosearray: %v", err)
	}
	// malformed object
	if _, _, err := scanToken("{10:10}", config); err != ErrorExpectedKey {
		t.Errorf("fail expecting ErrorExpectedString: %v", err)
	}
	if _, _, err := scanToken(`{"" 10}`, config); err != ErrorExpectedKey {
		t.Errorf("fail expecting ErrorExpectedKey: %v", err)
	}
	if _, _, err := scanToken(`{"10" 10}`, config); err != ErrorExpectedColon {
		t.Errorf("fail expecting ErrorExpectedColon: %v", err)
	}
	if _, _, err := scanToken(`{"10": nill}`, config); err != ErrorExpectedNil {
		t.Errorf("fail expecting ErrorExpectedNil: %v", err)
	}
	_, _, err := scanToken(`{"10": null`, config)
	if err != ErrorExpectedCloseobject {
		t.Errorf("fail expecting ErrorExpectedCloseobject: %v", err)
	}
	// malformed token
	if _, _, err := scanToken("<>", config); err != ErrorExpectedToken {
		t.Errorf("fail expecting ErrorExpectedToken: %v", err)
	}
}

func TestCodeJSON(t *testing.T) {
	var ref interface{}

	data := codeJSON()
	config := NewDefaultConfig()
	if err := json.Unmarshal(data, &ref); err != nil {
		t.Errorf("error parsing codeJSON: %v", err)
	}
	if val, remtxt, err := config.Parse(string(data)); err != nil {
		t.Errorf("error parsing codeJSON: %v", err)
	} else if remtxt != "" {
		t.Errorf("remaining text after parsing should be empty, %q", remtxt)
	} else if reflect.DeepEqual(val, ref) == false {
		t.Error("codeJSON parsing failed with reference: %v", err)
	}
}

//func BenchmarkScanNumFloat(b *testing.B) {
//    in := []byte("100000.23")
//    for i := 0; i < b.N; i++ {
//        scanNum(in, FloatNumber)
//    }
//}
//
//func BenchmarkScanNumString(b *testing.B) {
//    in := []byte("100000.23")
//    for i := 0; i < b.N; i++ {
//        scanNum(in, StringNumber)
//    }
//}
//
//func BenchmarkScanString(b *testing.B) {
//    in := []byte(`"汉语 / 漢語; Hàn\b \tyǔ "`)
//    for i := 0; i < b.N; i++ {
//        if _, _, err := scanString(in); err != nil {
//            b.Fatal(err)
//        }
//    }
//}
//
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
