//  Copyright (c) 2015 Couchbase, Inc.

package gson

import "testing"
import "strings"
import "sort"
import "fmt"

var _ = fmt.Sprintf("dummy")

func TestParsePointer(t *testing.T) {
	var tcasesJSONPointers = [][2]interface{}{
		[2]interface{}{``, []string{}},
		[2]interface{}{`/`, []string{""}},
		[2]interface{}{"/foo", []string{"foo"}},
		[2]interface{}{"/foo/0", []string{"foo", "0"}},
		[2]interface{}{"/a~1b", []string{"a/b"}},
		[2]interface{}{"/c%d", []string{"c%d"}},
		[2]interface{}{"/e^f", []string{"e^f"}},
		[2]interface{}{"/g|h", []string{"g|h"}},
		[2]interface{}{"/i\\j", []string{"i\\j"}},
		[2]interface{}{"/k\"l", []string{"k\"l"}},
		[2]interface{}{"/ ", []string{" "}},
		[2]interface{}{"/m~0n", []string{"m~n"}},
		[2]interface{}{"/g~1n~1r", []string{"g/n/r"}},
		[2]interface{}{"/g/汉语/r", []string{"g", "汉语", "r"}},
	}

	// test ParseJsonPointer
	config := NewDefaultConfig()
	for _, tcase := range tcasesJSONPointers {
		in, ref := tcase[0].(string), tcase[1].([]string)
		t.Logf("input pointer %q", in)
		segments := config.ParseJsonPointer(in, []string{})
		if len(segments) != len(ref) {
			t.Errorf("expected %v, got %v", len(ref), len(segments))
		} else {
			for i, x := range ref {
				if string(segments[i]) != string(x) {
					t.Errorf("expected %q, got %q", string(x), segments[i])
				}
			}
		}
	}

	// test encode pointers
	out := make([]byte, 1024)
	for _, tcase := range tcasesJSONPointers {
		in, ref := tcase[0].(string), tcase[1].([]string)
		t.Logf("input %v", ref)
		n := config.ToJsonPointer(ref, out)
		if outs := string(out[:n]); outs != in {
			t.Errorf("expected %q, %q", in, outs)
		}
	}
}

func TestTypicalPointers(t *testing.T) {
	refs := strings.Split(string(testdataFile("testdata/typical_pointers")), "\n")
	refs = refs[:len(refs)-1] // skip the last empty line
	sort.Strings(refs)
	config := NewDefaultConfig()

	txt := string(testdataFile("testdata/typical.json"))
	_, value := config.JsonToValue(txt)

	// test list pointers
	pointers := config.ListPointers(value, make([]string, 0, 1024))
	sort.Strings(pointers)
	if len(refs) != len(pointers) {
		t.Errorf("expected %v, got %v", len(refs), len(pointers))
	}
	for i, r := range refs {
		if r != pointers[i] {
			t.Errorf("expected %v, got %v", r, pointers[i])
		}
	}

	// test list pointers for document using [][2]interface{} for map.
	value = GolangMap2cborMap(value)
	pointers = config.ListPointers(value, make([]string, 0, 1024))
	sort.Strings(pointers)
	if len(refs) != len(pointers) {
		t.Errorf("expected %v, got %v", len(refs), len(pointers))
	}
	for i, r := range refs {
		if r != pointers[i] {
			t.Errorf("expected %v, got %v", r, pointers[i])
		}
	}
}

func BenchmarkParseJsonPtr3(b *testing.B) {
	config := NewDefaultConfig()
	path := "/foo/g/0"
	segments := make([]string, 0, 16)
	b.SetBytes(int64(len(path)))
	for i := 0; i < b.N; i++ {
		config.ParseJsonPointer(path, segments)
	}
}

func BenchmarkParseJsonPtr4(b *testing.B) {
	config := NewDefaultConfig()
	path := "/foo/g~1n~1r/0/hello"
	segments := make([]string, 0, 16)
	b.SetBytes(int64(len(path)))
	for i := 0; i < b.N; i++ {
		config.ParseJsonPointer(path, segments)
	}
}

func BenchmarkParseJsonPtr5(b *testing.B) {
	config := NewDefaultConfig()
	out := make([]byte, 1024)
	n := config.ToJsonPointer([]string{"a", "ab", "a~b", "a/b", "a~/~/b"}, out)
	segments := make([]string, 0, 16)
	b.SetBytes(int64(n))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config.ParseJsonPointer(bytes2str(out[:n]), segments)
	}
}

func BenchmarkToJsonPtr5(b *testing.B) {
	config := NewDefaultConfig()
	path := []string{"a", "ab", "a~b", "a/b", "a~/~/b"}
	out := make([]byte, 1024)
	b.SetBytes(15)
	for i := 0; i < b.N; i++ {
		config.ToJsonPointer(path, out)
	}
}

func BenchmarkListPointers(b *testing.B) {
	config := NewDefaultConfig()
	txt := string(testdataFile("testdata/typical.json"))
	_, doc := config.JsonToValue(txt)
	pointers := make([]string, 0, 1024)
	b.SetBytes(int64(len(txt)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config.ListPointers(doc, pointers)
	}
}
