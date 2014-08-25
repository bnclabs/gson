package json

import (
    "testing"
    "reflect"
    "io/ioutil"
    "compress/gzip"
    "encoding/json"
    "os"
)

var testcases = []string {
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
    ` [ null, true, false, 10, "tru\"e" ] `,
    // object
    ` { "a": null, "b" : true,"c":false, "d\"":10, "e":"tru\"e" } `,
}

func TestParse(t *testing.T) {
    var ref interface{}

    p := NewParser(FloatNumber, AnsiSpace, false/*jsonp*/)
    for _, tcase := range testcases {
        if err := json.Unmarshal([]byte(tcase), &ref); err != nil {
            t.Errorf("failed on %v ; %v", tcase, err)
        }
        val, _, err := p.Parse([]byte(tcase))
        if err != nil {
            t.Errorf("failed on %v ; %v", tcase, err)
        }
        if reflect.DeepEqual(val, ref) == false {
            t.Errorf("failed on %v %T %v - %T %v", tcase, ref, ref, val, val)
        }
    }
}

func TestCodeJSON(t *testing.T) {
    var ref interface{}

    p := NewParser(FloatNumber, AnsiSpace, false/*jsonp*/)
    if err := json.Unmarshal(codeJSON, &ref); err != nil {
        t.Error(err)
    }
    val, _, err := p.Parse(codeJSON)
    if err != nil {
        t.Error(err)
    }
    if reflect.DeepEqual(val, ref) == false {
        t.Error("failed")
    }
}

func BenchmarkScanNumFloat(b *testing.B) {
    in := []byte("100000.23")
    for i := 0; i < b.N; i++ {
        scanNum(in, FloatNumber)
    }
}

func BenchmarkScanNumString(b *testing.B) {
    in := []byte("100000.23")
    for i := 0; i < b.N; i++ {
        scanNum(in, StringNumber)
    }
}

func BenchmarkScanString(b *testing.B) {
    in := []byte(`"汉语 / 漢語; Hàn\b \tyǔ "`)
    for i := 0; i < b.N; i++ {
        if _, _, err := scanString(in); err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkSmallJSONPkg(b *testing.B) {
    txt := []byte(`{"a": null, "b" : true,"c":false, "d\"":-10E-1, "e":"tru\"e" }`)
    p := NewParser(FloatNumber, AnsiSpace, false/*jsonp*/)
    b.SetBytes(int64(len(txt)))
    for i := 0; i < b.N; i++ {
        if _, _, err := p.Parse(txt); err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkSmallJSON(b *testing.B) {
    var m map[string]interface{}
    txt := []byte(`{"a": null, "b" : true,"c":false, "d\"":-10E-1, "e":"tru\"e" }`)
    b.SetBytes(int64(len(txt)))
    for i := 0; i < b.N; i++ {
        if err := json.Unmarshal(txt, &m); err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkCodeJSONPkg(b *testing.B) {
    p := NewParser(FloatNumber, AnsiSpace, false/*jsonp*/)
    b.SetBytes(int64(len(codeJSON)))
    for i := 0; i < b.N; i++ {
        if _, _, err := p.Parse(codeJSON); err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkCodeJSON(b *testing.B) {
    var m map[string]interface{}
    b.SetBytes(int64(len(codeJSON)))
    for i := 0; i < b.N; i++ {
        if err := json.Unmarshal(codeJSON, &m); err != nil {
            b.Fatal(err)
        }
    }
}

var codeJSON []byte

func init() {
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
    codeJSON = data
}
