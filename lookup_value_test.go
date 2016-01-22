//  Copyright (c) 2015 Couchbase, Inc.

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
	_, value := config.NewJson([]byte(txt), -1).Tovalue()
	val := config.NewValue(value)
	ptr := config.NewJsonpointer("")

	t.Logf("%v", txt)

	for _, tcase := range testcases {
		t.Logf("%v", tcase[0].(string))

		ptr = ptr.ResetPath(tcase[0].(string))
		if item := val.Get(ptr); !reflect.DeepEqual(item, tcase[1]) {
			fmsg := "for %q expected %v, got %v"
			t.Errorf(fmsg, string(ptr.Path()), tcase[1], item)
		}
	}
}

func TestPointerSetExample(t *testing.T) {
	config := NewDefaultConfig()
	val := config.NewValue([]interface{}{"hello"})
	ptr := config.NewJsonpointer("/-")
	val.Set(ptr, "world")
	nitem, oitem := val.Set(ptr, "world")
	if !reflect.DeepEqual(oitem, "world") {
		t.Errorf("for %v expected %v, got %v", ptr.Path(), "world", oitem)
	} else if v := nitem.([]interface{}); !reflect.DeepEqual(v[0], "hello") {
		t.Errorf("for %v expected %v, got %v", ptr.Path(), "hello", v[0])
	} else if !reflect.DeepEqual(v[1], "world") {
		t.Errorf("for %v expected %v, got %v", ptr.Path(), "world", v[1])
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
	_, value := config.NewJson([]byte(txt), -1).Tovalue()
	val := config.NewValue(value)
	ptr := config.NewJsonpointer("")

	for _, tcase := range testcases {
		t.Logf("%v", tcase[0].(string))

		ptr.ResetPath(tcase[0].(string))
		value, oitem := val.Set(ptr, tcase[1])
		val = config.NewValue(value)
		if !reflect.DeepEqual(oitem, tcase[2]) {
			fmsg := "for %q expected %v, got %v"
			t.Errorf(fmsg, string(ptr.Path()), tcase[2], oitem)
		}

		item := val.Get(ptr)
		if !reflect.DeepEqual(item, tcase[1]) {
			fmsg := "for %q expected %v, got %v"
			t.Errorf(fmsg, string(ptr.Path()), tcase[1], item)
		}
	}

	var refval interface{}
	if err := json.Unmarshal([]byte(ref), &refval); err != nil {
		t.Fatalf("unmarshal: %v", err)
	} else if !reflect.DeepEqual(refval, val.data) {
		t.Errorf("expected %v, got %v", refval, val.data)
	}
}

func TestPointerDelExample(t *testing.T) {
	config := NewDefaultConfig()
	val := config.NewValue([]interface{}{"hello", "world"})
	ptr := config.NewJsonpointer("/1")

	value, deleted := val.Delete(ptr)
	if !reflect.DeepEqual(deleted, "world") {
		t.Errorf("for %v expected %v, got %v", ptr, "world", deleted)
	} else if v := value.([]interface{}); len(v) != 1 {
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

	config := NewDefaultConfig()
	_, value := config.NewJson([]byte(txt), -1).Tovalue()
	val := config.NewValue(value)
	ptr := config.NewJsonpointer("")

	for _, tcase := range testcases {
		t.Logf("%v", tcase)

		ptr.ResetPath(tcase[0].(string))
		value, deleted := val.Delete(ptr)
		if !reflect.DeepEqual(deleted, tcase[1]) {
			t.Errorf("for %v expected %v, got %v", ptr, tcase[1], deleted)
		}
		val = config.NewValue(value)
	}

	remtxt := `{"arr": [], "-": [], "dict":{}}"`
	_, remvalue := config.NewJson([]byte(remtxt), -1).Tovalue()
	if !reflect.DeepEqual(val.data, remvalue) {
		t.Errorf("expected %v, got %v", remvalue, val.data)
	}
}

func BenchmarkValueGet(b *testing.B) {
	config := NewDefaultConfig()
	data := testdataFile("testdata/typical.json")
	_, value := config.NewJson(data, -1).Tovalue()
	val := config.NewValue(value)
	ptr := config.NewJsonpointer("/projects/Sherri/members/0")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		val.Get(ptr)
	}
}

func BenchmarkValueSet(b *testing.B) {
	config := NewDefaultConfig()
	data := testdataFile("testdata/typical.json")
	_, value := config.NewJson(data, -1).Tovalue()
	val := config.NewValue(value)
	ptr := config.NewJsonpointer("/projects/Sherri/members/0")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		val.Set(ptr, 10)
	}
}

func BenchmarkValueDelete(b *testing.B) {
	config := NewDefaultConfig()
	data := testdataFile("testdata/typical.json")
	_, value := config.NewJson(data, -1).Tovalue()
	val := config.NewValue(value)

	ptrd := config.NewJsonpointer("/projects/Sherri/members/0")
	ptrs := config.NewJsonpointer("/projects/Sherri/members/-")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		val.Delete(ptrd)
		val.Set(ptrs, 10)
	}
}
