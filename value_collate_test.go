package gson

import "fmt"
import "testing"

var _ = fmt.Sprintf("dummy")

// All test cases are folded into collate_value_test.go, contains only few
// missing testcases (if any) and benchmarks.

func BenchmarkVal2CollNil(b *testing.B) {
	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 1024), 0)
	val := config.NewValue(nil)

	for i := 0; i < b.N; i++ {
		val.Tocollate(clt.Reset(nil))
	}
}

func BenchmarkVal2CollTrue(b *testing.B) {
	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 1024), 0)
	val := config.NewValue(interface{}(true))

	for i := 0; i < b.N; i++ {
		val.Tocollate(clt.Reset(nil))
	}
}

func BenchmarkVal2CollFalse(b *testing.B) {
	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 1024), 0)
	val := config.NewValue(interface{}(false))

	for i := 0; i < b.N; i++ {
		val.Tocollate(clt.Reset(nil))
	}
}

func BenchmarkVal2CollF64(b *testing.B) {
	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 1024), 0)
	val := config.NewValue(interface{}(float64(10.121312213123123)))

	for i := 0; i < b.N; i++ {
		val.Tocollate(clt.Reset(nil))
	}
}

func BenchmarkVal2CollI64(b *testing.B) {
	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 1024), 0)
	val := config.NewValue(interface{}(int64(123456789)))

	for i := 0; i < b.N; i++ {
		val.Tocollate(clt.Reset(nil))
	}
}

func BenchmarkVal2CollMiss(b *testing.B) {
	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 1024), 0)
	val := config.NewValue(interface{}(MissingLiteral))

	for i := 0; i < b.N; i++ {
		val.Tocollate(clt.Reset(nil))
	}
}

func BenchmarkVal2CollStr(b *testing.B) {
	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 1024), 0)
	val := config.NewValue(interface{}("hello world"))

	for i := 0; i < b.N; i++ {
		val.Tocollate(clt.Reset(nil))
	}
}

func BenchmarkVal2CollArr(b *testing.B) {
	arr := []interface{}{nil, true, false, "hello world", 10.23122312}
	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 1024), 0)
	val := config.NewValue(interface{}(arr))

	for i := 0; i < b.N; i++ {
		val.Tocollate(clt.Reset(nil))
	}
}

func BenchmarkVal2CollMap(b *testing.B) {
	obj := map[string]interface{}{
		"key1": nil, "key2": true, "key3": false, "key4": "hello world",
		"key5": 10.23122312,
	}
	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 1024), 0)
	val := config.NewValue(interface{}(obj))

	for i := 0; i < b.N; i++ {
		val.Tocollate(clt.Reset(nil))
	}
}

func BenchmarkVal2CollTyp(b *testing.B) {
	config := NewDefaultConfig()
	jsn := config.NewJson(testdataFile("testdata/typical.json"), -1)
	clt := config.NewCollate(make([]byte, 10*1024), 0)
	_, value := jsn.Tovalue()
	val := config.NewValue(value)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		val.Tocollate(clt.Reset(nil))
	}
}
