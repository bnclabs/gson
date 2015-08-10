package cbor

import "testing"
import "fmt"
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

func TestTypicalPointers(t *testing.T) {
	config := NewConfig(FloatNumber, UnicodeSpace)
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
		//t.Logf("%v", cbordoc)
		n = config.Get(cbordoc, cborptr[:n], item)
		val, _ := config.Decode(item[:n])
		if !reflect.DeepEqual(cborMap2golangMap(val), ref) {
			fmsg := "expected {%T,%v} for ptr %q, got {%T,%v}"
			t.Fatalf(fmsg, ref, ref, ptr, val, val)
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
