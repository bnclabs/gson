package gson

import "testing"
import "reflect"
import "strings"
import "sort"
import "fmt"

var _ = fmt.Sprintf("dummy")

var tcasesJSONPointers = []struct {
	in  string
	ref []string
}{
	{``, []string{}},
	{`/`, []string{""}},
	{"/foo", []string{"foo"}},
	{"/foo/0", []string{"foo", "0"}},
	{"/a~1b", []string{"a/b"}},
	{"/c%d", []string{"c%d"}},
	{"/e^f", []string{"e^f"}},
	{"/g|h", []string{"g|h"}},
	{"/i\\j", []string{"i\\j"}},
	{"/k\"l", []string{"k\"l"}},
	{"/ ", []string{" "}},
	{"/m~0n", []string{"m~n"}},
	{"/g~1n~1r", []string{"g/n/r"}},
	{"/g/n/r", []string{"g", "n", "r"}},
}

func TestParsePointer(t *testing.T) {
	config := NewDefaultConfig()
	for _, tcase := range tcasesJSONPointers {
		t.Logf("input pointer %q", tcase.in)
		segments := config.ParsePointer(tcase.in, []string{})
		if len(segments) != len(tcase.ref) {
			t.Errorf("expected %v, got %v", len(tcase.ref), len(segments))
		} else {
			for i, x := range tcase.ref {
				if string(segments[i]) != string(x) {
					t.Errorf("expected %q, got %q", string(x), segments[i])
				}
			}
		}
	}
}

func TestEncodePointer(t *testing.T) {
	config := NewDefaultConfig()
	out := make([]byte, 1024)
	for _, tcase := range tcasesJSONPointers {
		t.Logf("input %v", tcase.ref)
		n := config.EncodePointer(tcase.ref, out)
		if outs := string(out[:n]); outs != tcase.in {
			t.Errorf("expected %q, %q", tcase.in, outs)
		}
	}
}

func TestTypicalPointers(t *testing.T) {
	refs := strings.Split(string(testdataFile("testdata/typical_pointers")), "\n")
	refs = refs[:len(refs)-1] // skip the last empty line
	sort.Strings(refs)
	config := NewDefaultConfig()

	txt := string(testdataFile("testdata/typical.json"))
	value, _ := config.Parse(txt)
	pointers := config.ListPointers(value)
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

func TestPointerGet(t *testing.T) {
	txt := `{"a": 10, "arr": [1,2], "dict": {"a":10, "b":20}}`
	testcases := [][2]interface{}{
		[2]interface{}{"/a", 10.0},
		[2]interface{}{"/arr/0", 1.0},
		[2]interface{}{"/arr/1", 2.0},
		[2]interface{}{"/dict/a", 10.0},
		[2]interface{}{"/dict/b", 20.0},
	}
	config := NewDefaultConfig()
	doc, _ := config.Parse(txt)
	t.Logf("%v", doc)
	for _, tcase := range testcases {
		ptr := tcase[0].(string)
		t.Logf("%v", ptr)
		if val := config.Get(ptr, doc); !reflect.DeepEqual(val, tcase[1]) {
			t.Errorf("for %v expected %v, got %v", ptr, tcase[1], val)
		}
	}
}

func TestPointerSet(t *testing.T) {
	txt := `{"a": 10, "arr": [1,2], "dict": {"a":10, "b":20}}`
	testcases := [][3]interface{}{
		[3]interface{}{"/a", 11.0, 10.0},
		[3]interface{}{"/b", 12.0, 12.0},
		[3]interface{}{"/arr/0", 10.0, 1.0},
		[3]interface{}{"/arr/1", 20.0, 2.0},
		[3]interface{}{"/arr/-", 30.0, 30.0},
		[3]interface{}{"/dict/a", 1.0, 10.0},
		[3]interface{}{"/dict/b", 2.0, 20.0},
	}
	config := NewDefaultConfig()
	doc, _ := config.Parse(txt)
	for _, tcase := range testcases {
		ptr := tcase[0].(string)
		t.Logf("%v", ptr)
		doc, old := config.Set(ptr, doc, tcase[1])
		if !reflect.DeepEqual(old, tcase[2]) {
			t.Errorf("for %v expected %v, got %v", ptr, tcase[2], old)
		}
		val := config.Get(ptr, doc)
		if !reflect.DeepEqual(val, tcase[1]) {
			t.Errorf("for %v expected %v, got %v", ptr, tcase[1], val)
		}
	}
}

func TestPointerDel(t *testing.T) {
	txt := `{"a": 10, "arr": [1,2], "dict": {"a":10, "b":20}}`
	testcases := [][2]interface{}{
		[2]interface{}{"/a", 10.0},
		[2]interface{}{"/arr/1", 2.0},
		[2]interface{}{"/arr/0", 1.0},
		[2]interface{}{"/dict/a", 10.0},
		[2]interface{}{"/dict/b", 20.0},
	}
	var val interface{}
	config := NewDefaultConfig()
	doc, _ := config.Parse(txt)
	for _, tcase := range testcases {
		ptr := tcase[0].(string)
		doc, val = config.Delete(ptr, doc)
		if !reflect.DeepEqual(val, tcase[1]) {
			t.Errorf("for %v expected %v, got %v", ptr, tcase[1], val)
		}
	}
	remtxt := `{"arr": [], "dict":{}}"`
	remdoc, _ := config.Parse(remtxt)
	if !reflect.DeepEqual(doc, remdoc) {
		t.Errorf("expected %v, got %v", remdoc, doc)
	}
}

func BenchmarkParsePtrSmall(b *testing.B) {
	config := NewDefaultConfig()
	path := "/foo/g/0"
	segments := []string{}
	b.SetBytes(int64(len(path)))
	for i := 0; i < b.N; i++ {
		config.ParsePointer(path, segments)
	}
}

func BenchmarkParsePtrMed(b *testing.B) {
	config := NewDefaultConfig()
	path := "/foo/g~1n~1r/0/hello"
	segments := []string{}
	b.SetBytes(int64(len(path)))
	for i := 0; i < b.N; i++ {
		config.ParsePointer(path, segments)
	}
}

func BenchmarkParsePtrLarge(b *testing.B) {
	config := NewDefaultConfig()
	out := make([]byte, 1024)
	n := config.EncodePointer([]string{"a", "ab", "a~b", "a/b", "a~/~/b"}, out)
	segments := []string{}
	b.SetBytes(int64(n))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config.ParsePointer(string(out[:n]), segments)
	}
}

func BenchmarkEncPtrLarge(b *testing.B) {
	config := NewDefaultConfig()
	path := []string{"a", "ab", "a~b", "a/b", "a~/~/b"}
	out := make([]byte, 1024)
	b.SetBytes(15)
	for i := 0; i < b.N; i++ {
		config.EncodePointer(path, out)
	}
}
