package gson

import "testing"
import "reflect"
import "encoding/json"

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
	_, doc := config.JsonToValue(txt)
	t.Logf("%v", doc)
	for _, tcase := range testcases {
		ptr := tcase[0].(string)
		t.Logf("%v", ptr)
		if val := config.DocGet(ptr, doc); !reflect.DeepEqual(val, tcase[1]) {
			t.Errorf("for %v expected %v, got %v", ptr, tcase[1], val)
		}
	}
}

func TestPointerSetExample(t *testing.T) {
	config := NewDefaultConfig()
	var doc interface{}
	doc = []interface{}{"hello"}
	ptr := "/-"
	doc, old := config.DocSet(ptr, doc, "world")
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
	_, doc := config.JsonToValue(txt)
	for _, tcase := range testcases {
		ptr := tcase[0].(string)
		t.Logf("%v", ptr)
		doc, old := config.DocSet(ptr, doc, tcase[1])
		if !reflect.DeepEqual(old, tcase[2]) {
			t.Errorf("for %v expected %v, got %v", ptr, tcase[2], old)
		}
		val := config.DocGet(ptr, doc)
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
	doc, old := config.DocDelete(ptr, doc)
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
	_, doc := config.JsonToValue(txt)
	for _, tcase := range testcases {
		ptr := tcase[0].(string)
		doc, val = config.DocDelete(ptr, doc)
		if !reflect.DeepEqual(val, tcase[1]) {
			t.Errorf("for %v expected %v, got %v", ptr, tcase[1], val)
		}
	}
	remtxt := `{"arr": [], "-": [], "dict":{}}"`
	_, remdoc := config.JsonToValue(remtxt)
	if !reflect.DeepEqual(doc, remdoc) {
		t.Errorf("expected %v, got %v", remdoc, doc)
	}
}

func BenchmarkPointerGet(b *testing.B) {
	config := NewDefaultConfig()
	txt := string(testdataFile("testdata/typical.json"))
	_, doc := config.JsonToValue(txt)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config.DocGet("/projects/Sherri/members/0", doc)
	}
}

func BenchmarkPointerSet(b *testing.B) {
	config := NewDefaultConfig()
	txt := string(testdataFile("testdata/typical.json"))
	_, doc := config.JsonToValue(txt)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config.DocSet("/projects/Sherri/members/0", doc, 10)
	}
}

func BenchmarkPointerDelete(b *testing.B) {
	config := NewDefaultConfig()
	txt := string(testdataFile("testdata/typical.json"))
	_, doc := config.JsonToValue(txt)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config.DocDelete("/projects/Sherri/members/0", doc)
		config.DocSet("/projects/Sherri/members/-", doc, 10)
	}
}
