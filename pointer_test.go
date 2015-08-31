package gson

import "testing"
import "reflect"
import "strings"
import "sort"
import "fmt"
import "encoding/json"

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
		[2]interface{}{"/g/n/r", []string{"g", "n", "r"}},
	}

	// test ParsePointer
	config := NewDefaultConfig()
	for _, tcase := range tcasesJSONPointers {
		in, ref := tcase[0].(string), tcase[1].([]string)
		t.Logf("input pointer %q", in)
		segments := config.ParsePointer(in, []string{})
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
		n := config.EncodePointer(ref, out)
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

	// test list pointers
	txt := string(testdataFile("testdata/typical.json"))
	_, value := config.Parse(txt)
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
	_, doc := config.Parse(txt)
	t.Logf("%v", doc)
	for _, tcase := range testcases {
		ptr := tcase[0].(string)
		t.Logf("%v", ptr)
		if val := config.Get(ptr, doc); !reflect.DeepEqual(val, tcase[1]) {
			t.Errorf("for %v expected %v, got %v", ptr, tcase[1], val)
		}
	}
}

func TestPointerSetExample(t *testing.T) {
	config := NewDefaultConfig()
	var doc interface{}
	doc = []interface{}{"hello"}
	ptr := "/-"
	doc, old := config.Set(ptr, doc, "world")
	if !reflect.DeepEqual(old, "world") {
		t.Errorf("for %v expected %v, got %v", ptr, "world", old)
	} else if v := doc.([]interface{}); !reflect.DeepEqual(v[0], "hello") {
		t.Errorf("for %v expected %v, got %v", ptr, "hello", v[0])
	} else if !reflect.DeepEqual(v[1], "world") {
		t.Errorf("for %v expected %v, got %v", ptr, "world", v[1])
	}
}

func TestPointerSet(t *testing.T) {
	txt := `{"a": 10, "arr": [1,2], "-": [[1]], "dict": {"a":10, "b":20}}`
	ref := `{"b":1,"a":11,"arr":[10,20,30],"-":[[1,30],30],"dict":{"a":1,"b":2}}`
	testcases := [][3]interface{}{
		[3]interface{}{"/a", 11.0, 10.0},
		[3]interface{}{"/b", 1.0, 1.0},
		[3]interface{}{"/arr/0", 10.0, 1.0},
		[3]interface{}{"/arr/1", 20.0, 2.0},
		[3]interface{}{"/arr/-", 30.0, 30.0},
		[3]interface{}{"/-/-", 30.0, 30.0},
		[3]interface{}{"/-/0/-", 30.0, 30.0},
		[3]interface{}{"/dict/a", 1.0, 10.0},
		[3]interface{}{"/dict/b", 2.0, 20.0},
	}
	config := NewDefaultConfig()
	_, doc := config.Parse(txt)
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
	var refval interface{}
	if err := json.Unmarshal([]byte(ref), &refval); err != nil {
		t.Fatalf("unmarshal: %v", err)
	} else if !reflect.DeepEqual(refval, doc) {
		t.Errorf("expected %v, got %v", refval, doc)
	}
}

func TestPointerDelExample(t *testing.T) {
	config := NewDefaultConfig()
	var doc interface{}
	doc = []interface{}{"hello", "world"}
	ptr := "/1"
	doc, old := config.Delete(ptr, doc)
	if !reflect.DeepEqual(old, "world") {
		t.Errorf("for %v expected %v, got %v", ptr, "world", old)
	} else if v := doc.([]interface{}); len(v) != 1 {
		t.Errorf("for %v expected length %v, got %v", ptr, 1, len(v))
	} else if !reflect.DeepEqual(v[0], "hello") {
		t.Errorf("for %v expected %v, got %v", ptr, "hello", v[0])
	}
}

func TestPointerDel(t *testing.T) {
	txt := `{"a": 10, "-": [1], "arr": [1,2], "dict": {"a":10, "b":20}}`
	testcases := [][2]interface{}{
		[2]interface{}{"/a", 10.0},
		[2]interface{}{"/arr/1", 2.0},
		[2]interface{}{"/arr/0", 1.0},
		[2]interface{}{"/-/0", 1.0},
		[2]interface{}{"/dict/a", 10.0},
		[2]interface{}{"/dict/b", 20.0},
	}
	var val interface{}
	config := NewDefaultConfig()
	_, doc := config.Parse(txt)
	for _, tcase := range testcases {
		ptr := tcase[0].(string)
		doc, val = config.Delete(ptr, doc)
		if !reflect.DeepEqual(val, tcase[1]) {
			t.Errorf("for %v expected %v, got %v", ptr, tcase[1], val)
		}
	}
	remtxt := `{"arr": [], "-": [], "dict":{}}"`
	_, remdoc := config.Parse(remtxt)
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

func BenchmarkListPointers(b *testing.B) {
	config := NewDefaultConfig()
	txt := string(testdataFile("testdata/typical.json"))
	_, doc := config.Parse(txt)
	b.SetBytes(int64(len(txt)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config.ListPointers(doc)
	}
}

func BenchmarkPtrGet(b *testing.B) {
	config := NewDefaultConfig()
	txt := string(testdataFile("testdata/typical.json"))
	_, doc := config.Parse(txt)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config.Get("/projects/Sherri/members/0", doc)
	}
}

func BenchmarkPtrSet(b *testing.B) {
	config := NewDefaultConfig()
	txt := string(testdataFile("testdata/typical.json"))
	_, doc := config.Parse(txt)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config.Set("/projects/Sherri/members/0", doc, 10)
	}
}

func BenchmarkPtrDelete(b *testing.B) {
	config := NewDefaultConfig()
	txt := string(testdataFile("testdata/typical.json"))
	_, doc := config.Parse(txt)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config.Delete("/projects/Sherri/members/0", doc)
		config.Set("/projects/Sherri/members/-", doc, 10)
	}
}
