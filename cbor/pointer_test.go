package cbor

import "testing"
import "fmt"
import "encoding/json"
import "reflect"
import "github.com/prataprc/gson"

var _ = fmt.Sprintf("dummy")

var tcasesJSONPointers = []string{
	``,
	`/`,
	"/foo",
	"/foo/0",
	"/a~1b",
	"/c%d",
	"/e^f",
	"/g|h",
	"/i\\j",
	"/k\"l",
	"/ ",
	"/m~0n",
	"/g~1n~1r",
	"/g/n/r",
}

func TestCborPointer(t *testing.T) {
	buf, out := make([]byte, 1024), make([]byte, 1024)
	config := NewDefaultConfig()
	for _, tcase := range tcasesJSONPointers {
		t.Logf(tcase)
		n := config.FromJsonPointer([]byte(tcase), buf)
		m := config.ToJsonPointer(buf[:n], out)
		if result := string(out[:m]); tcase != result {
			t.Errorf("expected %q, got %q", tcase, result)
		}
	}
}

func TestCborTypicalPointers(t *testing.T) {
	config := NewDefaultConfig()
	cborptr := make([]byte, 1024)
	cbordoc := make([]byte, 1024*1024)
	item := make([]byte, 10*1024)

	gsonc := gson.NewDefaultConfig()
	txt := string(testdataFile("../testdata/typical.json"))
	doc, _ := gsonc.Parse(txt)
	pointers := gsonc.ListPointers(doc)
	_, n := config.ParseJson(txt, cbordoc)
	cbordoc = cbordoc[:n]
	for _, ptr := range pointers {
		if ln := len(ptr); ln > 0 && ptr[ln-1] == '-' {
			continue
		}
		t.Logf("pointer %v", ptr)
		ref := gsonc.Get(ptr, doc)
		n := config.FromJsonPointer([]byte(ptr), cborptr)
		t.Logf("%v", cbordoc)
		n = config.Get(cbordoc, cborptr[:n], item)
		val, _ := config.Decode(item[:n])
		if !reflect.DeepEqual(CborMap2golangMap(val), ref) {
			fmsg := "expected {%T,%v} for ptr %q, got {%T,%v}"
			t.Fatalf(fmsg, ref, ref, ptr, val, val)
		}
	}
}

func TestCborGet(t *testing.T) {
	txt := `{"a": 10, "arr": [1,2], "dict": {"a":10, "b":20}}`
	testcases := [][2]interface{}{
		//[2]interface{}{"/a", 10.0},
		[2]interface{}{"/arr/0", 1.0},
		[2]interface{}{"/arr/1", 2.0},
		[2]interface{}{"/dict/a", 10.0},
		[2]interface{}{"/dict/b", 20.0},
	}

	cbordoc, cborptr := make([]byte, 1024), make([]byte, 1024)
	item := make([]byte, 1024)
	config := NewDefaultConfig()
	_, n := config.ParseJson(txt, cbordoc)
	cbordoc = cbordoc[:n]
	for _, tcase := range testcases {
		ptr := tcase[0].(string)
		ref := tcase[1]
		t.Logf("%v", ptr)
		n := config.FromJsonPointer([]byte(ptr), cborptr)
		n = config.Get(cbordoc, cborptr[:n], item)
		val, _ := config.Decode(item[:n]) // cbor->json
		if !reflect.DeepEqual(CborMap2golangMap(val), ref) {
			fmsg := "expected {%T,%v}, for ptr %v, got {%T,%v}"
			t.Fatalf(fmsg, ref, ref, ptr, val, val)
		}
	}
}

func TestCborSet(t *testing.T) {
	txt := `{"a": 10, "arr": [1,2], "dict": {"a":10, "b":20}}`
	testcases := [][4]interface{}{
		[4]interface{}{
			"/a", 11.0, `10`, `{"a":11,"arr":[1,2],"dict":{"a":10,"b":20}}`},
		[4]interface{}{
			"/arr/0", 10.0, `1`, `{"a":11,"arr":[10,2],"dict":{"a":10,"b":20}}`},
		[4]interface{}{
			"/arr/1", 20.0, `2`, `{"a":11,"arr":[10,20],"dict":{"a":10,"b":20}}`},
		[4]interface{}{
			"/dict/a", 1.0, `10`, `{"a":11,"arr":[10,20],"dict":{"a":1,"b":20}}`},
		[4]interface{}{
			"/dict/b", 2.0, `20`, `{"a":11,"arr":[10,20],"dict":{"a":1,"b":2}}`},
	}

	// cbor initialization
	cbordoc, cbordocnew := make([]byte, 1024), make([]byte, 1024)
	cborptr := make([]byte, 1024)
	item, itemold := make([]byte, 1024), make([]byte, 1024)
	config := NewDefaultConfig()
	_, n := config.ParseJson(txt, cbordoc)
	cbordoc = cbordoc[:n]

	for _, tcase := range testcases {
		ptr := tcase[0].(string)
		t.Logf("%v", ptr)
		n := config.Encode(tcase[1], item)
		item = item[:n]

		config.FromJsonPointer([]byte(ptr), cborptr)
		n, m := config.Set(cbordoc, cborptr /*[:n]*/, item, cbordocnew, itemold)
		copy(cbordoc, cbordocnew[:n])
		cbordoc = cbordoc[:n]

		val, _ := config.Decode(cbordocnew[:n])
		var vref, iref interface{}
		if err := json.Unmarshal([]byte(tcase[3].(string)), &vref); err != nil {
			t.Fatalf("parsing json: %v", err)
		} else if v := CborMap2golangMap(val); !reflect.DeepEqual(v, vref) {
			t.Fatalf("for %v expected %v, got %v", ptr, vref, v)
		}
		oval, _ := config.Decode(itemold[:m])
		if err := json.Unmarshal([]byte(tcase[2].(string)), &iref); err != nil {
			t.Fatalf("parsing json: %v", err)
		} else if ov := CborMap2golangMap(oval); !reflect.DeepEqual(ov, iref) {
			t.Fatalf("for %v item expected %v, got %v", ptr, iref, ov)
		}
	}
}

func TestCborPrepend(t *testing.T) {
	txt := `{"a": 10, "arr": [1,2], "dict": {"a":10, "b":20}}`
	ftxt := `{"a": 10, "b": 20, "arr": [1,2,3],"dict":{"a":10,"b":20,"c":30}}`

	// cbor initialization
	cbordoc, cbordocnew := make([]byte, 1024), make([]byte, 1024)
	cborptr, item := make([]byte, 1024), make([]byte, 1024)

	config := NewDefaultConfig()
	_, n := config.ParseJson(txt, cbordoc)

	// prepend "/", {"b": 20}
	t.Logf(`prepend "/", {"b": 20}`)
	i := config.FromJsonPointer([]byte(""), cborptr)
	m := config.EncodeMapItems([][2]interface{}{[2]interface{}{"b", 20}}, item)
	n = config.Prepend(cbordoc[:n], cborptr[:i], item[:m], cbordocnew)
	copy(cbordoc, cbordocnew[:n])

	// prepend "/arr" 3.0
	t.Logf(`prepend "/arr" 3.0`)
	config.FromJsonPointer([]byte("/arr"), cborptr)
	m = config.Encode(float64(3.0), item)
	n = config.Prepend(cbordoc[:n], cborptr, item[:m], cbordocnew)
	copy(cbordoc, cbordocnew[:n])

	// prepend "/dict/c" 30.0
	t.Logf(`prepend "/dict/c" 30.0`)
	config.FromJsonPointer([]byte("/dict"), cborptr)
	m = config.EncodeMapItems([][2]interface{}{[2]interface{}{"c", 30}}, item)
	n = config.Prepend(cbordoc[:n], cborptr, item[:m], cbordocnew)
	copy(cbordoc, cbordocnew[:n])

	val, _ := config.Decode(cbordoc[:n])
	if err := json.Unmarshal([]byte(ftxt), &val); err != nil {
		t.Fatalf("parsing json: %v", err)
	} else if v := CborMap2golangMap(val); !reflect.DeepEqual(v, val) {
		t.Fatalf("finally exptected %v, got %v", val, v)
	}

	// parent doc as an array
	t.Logf(`prepend "" to [1,2]`)
	txt = `[1,2]`
	ftxt = `[1,2,3]`
	_, n = config.ParseJson(txt, cbordoc)
	config.FromJsonPointer([]byte(""), cborptr)
	m = config.Encode(float64(3.0), item)
	n = config.Prepend(cbordoc[:n], cborptr, item[:m], cbordocnew)
	copy(cbordoc, cbordocnew[:n])

	val, _ = config.Decode(cbordoc[:n])
	if err := json.Unmarshal([]byte(ftxt), &val); err != nil {
		t.Fatalf("parsing json: %v", err)
	} else if v := CborMap2golangMap(val); !reflect.DeepEqual(v, val) {
		t.Fatalf("finally exptected %v, got %v", val, v)
	}
}

func TestCborDel(t *testing.T) {
	txt := `{"a": "10", "arr": [1,2], "dict": {"a":10, "b":20}}`
	testcases := [][3]interface{}{
		[3]interface{}{
			"/a", `"10"`, `{"arr": [1,2], "dict": {"a":10, "b":20}}`},
		[3]interface{}{
			"/arr/1", `2`, `{"arr": [1], "dict": {"a":10, "b":20}}`},
		[3]interface{}{
			"/dict/a", `10`, `{"arr": [1], "dict": {"b":20}}`},
		[3]interface{}{
			"/dict", `{"b":20}`, `{"arr": [1]}`},
		[3]interface{}{
			"/arr", `[1]`, `{}`},
	}

	// cbor initialization
	cbordoc, cbordocnew := make([]byte, 1024), make([]byte, 1024)
	cborptr, itemold := make([]byte, 1024), make([]byte, 1024)
	config := NewDefaultConfig()
	_, n := config.ParseJson(txt, cbordoc)
	cbordoc = cbordoc[:n]

	for _, tcase := range testcases {
		ptr := tcase[0].(string)
		t.Logf("%v", ptr)

		x := config.FromJsonPointer([]byte(ptr), cborptr)
		n, m := config.Delete(cbordoc, cborptr[:x], cbordocnew, itemold)
		copy(cbordoc, cbordocnew[:n])
		cbordoc = cbordoc[:n]

		val, _ := config.Decode(cbordocnew[:n])
		var vref, iref interface{}
		if err := json.Unmarshal([]byte(tcase[2].(string)), &vref); err != nil {
			t.Fatalf("parsing json: %v", err)
		} else if v := CborMap2golangMap(val); !reflect.DeepEqual(v, vref) {
			t.Fatalf("for %v expected %v, got %v", ptr, vref, v)
		}
		oval, _ := config.Decode(itemold[:m])
		if err := json.Unmarshal([]byte(tcase[1].(string)), &iref); err != nil {
			t.Fatalf("parsing json: %v", err)
		} else if ov := CborMap2golangMap(oval); !reflect.DeepEqual(ov, iref) {
			t.Fatalf("for %v item expected %v, got %v", ptr, iref, ov)
		}
	}
}

func BenchmarkPtrJsonCborS(b *testing.B) {
	config := NewDefaultConfig()
	jsonptr := []byte("/foo/g/0")
	out := make([]byte, 1024)
	b.SetBytes(int64(len(jsonptr)))
	for i := 0; i < b.N; i++ {
		config.FromJsonPointer(jsonptr, out)
	}
}

func BenchmarkPtrJsonCborM(b *testing.B) {
	config := NewDefaultConfig()
	jsonptr := []byte("/foo/g~1n~1r/0/hello")
	out := make([]byte, 1024)
	b.SetBytes(int64(len(jsonptr)))
	for i := 0; i < b.N; i++ {
		config.FromJsonPointer(jsonptr, out)
	}
}

func BenchmarkPtrJsonCborL(b *testing.B) {
	gsonc, out := gson.NewDefaultConfig(), make([]byte, 1024)
	n := gsonc.EncodePointer([]string{"/a", "ab", "a~b", "a/b", "a~/~/b"}, out)
	jsonptr := make([]byte, 1024)
	copy(jsonptr, out[:n])

	config := NewDefaultConfig()
	b.SetBytes(int64(n))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config.FromJsonPointer(jsonptr[:n], out)
	}
}

func BenchmarkPtrCborJsonL(b *testing.B) {
	gsonc, out := gson.NewDefaultConfig(), make([]byte, 1024)
	n := gsonc.EncodePointer([]string{"/a", "ab", "a~b", "a/b", "a~/~/b"}, out)
	jsonptr := make([]byte, 1024)
	copy(jsonptr, out[:n]) // shall copy out to new location

	config := NewDefaultConfig()
	m := config.FromJsonPointer(jsonptr[:n], out)

	jsonout := make([]byte, 1024)
	b.SetBytes(int64(m))
	for i := 0; i < b.N; i++ {
		config.ToJsonPointer(out[:m], jsonout)
	}
}

func BenchmarkPtrCborGet(b *testing.B) {
	config := NewDefaultConfig()
	txt := string(testdataFile("../testdata/typical.json"))

	cbordoc := make([]byte, 10*1024)
	_, n := config.ParseJson(txt, cbordoc)
	cborptr := make([]byte, 10*1024)
	m := config.FromJsonPointer([]byte("/projects/Sherri/members/0"), cborptr)
	item := make([]byte, 10*1024)

	var p int
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p = config.Get(cbordoc[:n], cborptr[:m], item)
	}
	b.SetBytes(int64(p))
}

func BenchmarkPtrCborSet(b *testing.B) {
	//config := NewDefaultConfig()
	config := NewConfig(FloatNumber, UnicodeSpace, Stream)
	txt := string(testdataFile("../testdata/typical.json"))

	cbordoc := make([]byte, 10*1024)
	_, n := config.ParseJson(txt, cbordoc)

	cborptr := make([]byte, 10*1024)
	m := config.FromJsonPointer([]byte("/projects/Sherri/members/0"), cborptr)

	item := make([]byte, 10*1024)
	itemref := 10
	p := config.Encode(itemref, item)

	newdoc := make([]byte, 10*1024)
	old := make([]byte, 10*1024)

	var x, y int
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x, y = config.Set(cbordoc[:n], cborptr[:m], item[:p], newdoc, old)
	}
	b.SetBytes(int64(x + y))
}

func BenchmarkPtrCborPrepend(b *testing.B) {
	config := NewDefaultConfig()
	txt := string(testdataFile("../testdata/typical.json"))

	cbordoc := make([]byte, 10*1024)
	_, n := config.ParseJson(txt, cbordoc)
	cborptr1, cborptr2 := make([]byte, 10*1024), make([]byte, 10*1024)
	p := config.FromJsonPointer([]byte("/projects/Sherri/members"), cborptr1)
	q := config.FromJsonPointer([]byte("/projects/Sherri/members/0"), cborptr2)
	item := make([]byte, 10*1024)
	refitem := 10.0
	m := config.Encode(refitem, item)

	newdoc := make([]byte, 10*1024)

	var x, z int
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x = config.Prepend(cbordoc[:n], cborptr1[:p], item[:m], newdoc)
		n, m = config.Delete(newdoc[:x], cborptr2[:q], cbordoc, item)
	}
	if val, _ := config.Decode(item); !reflect.DeepEqual(val, refitem) {
		b.Fatalf("exptected %v, got %v", refitem, val)
	}
	b.SetBytes(int64(x + z))
}
