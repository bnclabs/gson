//  Copyright (c) 2015 Couchbase, Inc.

package gson

import "testing"
import "bytes"
import "reflect"
import "encoding/json"
import "fmt"
import "os"
import "strings"
import "compress/gzip"
import "io/ioutil"

var _ = fmt.Sprintf("dummy")

func TestConfig(t *testing.T) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(make([]byte, 16), 0)
	val := config.NewValue(10.2)
	val.Tocbor(cbr)
	if value := cbr.Tovalue(); !reflect.DeepEqual(val.data, value) {
		t.Errorf("expected %v got %v", val.data, value)
	}
}

type testLocal byte

func TestUndefined(t *testing.T) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(make([]byte, 16), 0)
	val := config.NewValue(CborUndefined(cborSimpleUndefined))
	val.Tocbor(cbr)
	if value := cbr.Tovalue(); !reflect.DeepEqual(val.data, value) {
		t.Errorf("expected %v got %v", val.data, value)
	}
	// test unknown type.
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		config.NewValue(testLocal(10)).Tocbor(cbr.Reset(nil))
	}()
}

func TestJsonToValue(t *testing.T) {
	config := NewDefaultConfig().SetSpaceKind(AnsiSpace)
	jsn := config.NewJson([]byte(`"abcd"  "xyz" "10" `), -1)
	jsnrmn, value := jsn.Tovalue()
	if string(jsnrmn.Bytes()) != `  "xyz" "10" ` {
		t.Errorf("expected %q, got %q", `  "xyz" "10" `, string(jsnrmn.Bytes()))
	}

	jsnback := config.NewJson(make([]byte, 1024), 0)
	config.NewValue(value).Tojson(jsnback)
	if ref := `"abcd"`; string(jsnback.Bytes()) != ref {
		t.Errorf("expected %v, got %v", ref, string(jsnback.Bytes()))
	}
}

func TestJsonToValues(t *testing.T) {
	var s string
	uni_s := `"汉语 / 漢語; Hàn\b \t\uef24yǔ "`
	json.Unmarshal([]byte(uni_s), &s)
	ref := []interface{}{"abcd", "xyz", "10", s}

	config := NewDefaultConfig().SetSpaceKind(AnsiSpace)
	jsn := config.NewJson([]byte(`"abcd"  "xyz" "10" `+uni_s), -1)
	if values := jsn.Tovalues(); !reflect.DeepEqual(values, ref) {
		t.Errorf("expected %v, got %v", ref, values)
	}
}

func TestParseJsonPointer(t *testing.T) {
	config := NewDefaultConfig()
	jptr := config.NewJsonpointer("/a/b")
	refsegs := [][]byte{[]byte("a"), []byte("b")}
	if segments := jptr.Segments(); !reflect.DeepEqual(segments, refsegs) {
		t.Errorf("expected %v, got %v", refsegs, segments)
	}
}

func TestToJsonPointer(t *testing.T) {
	config := NewDefaultConfig()
	refptr := config.NewJsonpointer("/a/b")
	jptr := config.NewJsonpointer("").ResetSegments([]string{"a", "b"})

	if bytes.Compare(jptr.path, refptr.path) != 0 {
		t.Errorf("expected %v, got %v", refptr.path, jptr.path)
	}
}

func TestGsonToCollate(t *testing.T) {
	config := NewDefaultConfig().SetNumberKind(IntNumber)
	clt := config.NewCollate(make([]byte, 1024), 0)
	config.NewValue(map[string]interface{}{"a": 10, "b": 20}).Tocollate(clt)
	ref := map[string]interface{}{"a": int64(10), "b": int64(20)}
	if value := clt.Tovalue(); !reflect.DeepEqual(ref, value) {
		t.Errorf("expected %v, got %v", ref, value)
	}
}

func TestCborToCollate(t *testing.T) {
	config := NewDefaultConfig().SetNumberKind(IntNumber)
	cbr := config.NewCbor(make([]byte, 1024), 0)
	clt := config.NewCollate(make([]byte, 1024), 0)
	out := config.NewCbor(make([]byte, 1024), 0)

	o := [][2]interface{}{
		[2]interface{}{"a", uint64(10)},
		[2]interface{}{"b", uint64(20)},
	}
	refm := CborMap2golangMap(o)

	value := config.NewValue(o).Tocbor(cbr).Tocollate(clt).Tocbor(out).Tovalue()
	if !reflect.DeepEqual(refm, value) {
		t.Errorf("expected %v, got %v", refm, value)
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
