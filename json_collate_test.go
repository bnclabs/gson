//  Copyright (c) 2015 Couchbase, Inc.

package gson

import "fmt"
import "testing"

var _ = fmt.Sprintf("dummy")

// All test cases are folded into collate_json_test.go, contains only few
// missing testcases (if any) and benchmarks.

func BenchmarkJson2CollNil(b *testing.B) {
	config := NewDefaultConfig()
	jsn := config.NewJson([]byte("null"), -1)
	clt := config.NewCollate(make([]byte, 1024), 0)

	for i := 0; i < b.N; i++ {
		jsn.Tocollate(clt.Reset(nil))
	}
}

func BenchmarkJson2CollTrue(b *testing.B) {
	config := NewDefaultConfig()
	jsn := config.NewJson([]byte("true"), -1)
	clt := config.NewCollate(make([]byte, 1024), 0)

	for i := 0; i < b.N; i++ {
		jsn.Tocollate(clt.Reset(nil))
	}
}

func BenchmarkJson2CollFalse(b *testing.B) {
	config := NewDefaultConfig()
	jsn := config.NewJson([]byte("false"), -1)
	clt := config.NewCollate(make([]byte, 1024), 0)

	for i := 0; i < b.N; i++ {
		jsn.Tocollate(clt.Reset(nil))
	}
}

func BenchmarkJson2CollF64(b *testing.B) {
	config := NewDefaultConfig()
	jsn := config.NewJson([]byte("10.121312213123123"), -1)
	clt := config.NewCollate(make([]byte, 1024), 0)

	for i := 0; i < b.N; i++ {
		jsn.Tocollate(clt.Reset(nil))
	}
}

func BenchmarkJson2CollI64(b *testing.B) {
	config := NewDefaultConfig()
	jsn := config.NewJson([]byte("123456789"), -1)
	clt := config.NewCollate(make([]byte, 1024), 0)

	for i := 0; i < b.N; i++ {
		jsn.Tocollate(clt.Reset(nil))
	}
}

func BenchmarkJson2CollMiss(b *testing.B) {
	inp := fmt.Sprintf(`"%s"`, MissingLiteral)

	config := NewDefaultConfig()
	jsn := config.NewJson([]byte(inp), -1)
	clt := config.NewCollate(make([]byte, 1024), 0)

	for i := 0; i < b.N; i++ {
		jsn.Tocollate(clt.Reset(nil))
	}
}

func BenchmarkJson2CollStr(b *testing.B) {
	config := NewDefaultConfig()
	jsn := config.NewJson([]byte(`"hello world"`), -1)
	clt := config.NewCollate(make([]byte, 1024), 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		jsn.Tocollate(clt.Reset(nil))
	}
}

func BenchmarkJson2CollArr(b *testing.B) {
	inp := `[null,true,false,"hello world",10.23122312]`
	config := NewDefaultConfig()
	jsn := config.NewJson([]byte(inp), -1)
	clt := config.NewCollate(make([]byte, 1024), 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		jsn.Tocollate(clt.Reset(nil))
	}
}

func BenchmarkJson2CollMap(b *testing.B) {
	inp := `{"key1":null,"key2":true,"key3":false,"key4":"hello world",` +
		`"key5":10.23122312}`
	config := NewDefaultConfig().SetMaxkeys(10)
	jsn := config.NewJson([]byte(inp), -1)
	clt := config.NewCollate(make([]byte, 1024), 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		jsn.Tocollate(clt.Reset(nil))
	}
}

func BenchmarkJson2CollTyp(b *testing.B) {
	inp := testdataFile("testdata/typical.json")
	config := NewDefaultConfig().SetMaxkeys(10)
	jsn := config.NewJson(inp, -1)
	clt := config.NewCollate(make([]byte, 10*1024), 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		jsn.Tocollate(clt.Reset(nil))
	}
}
