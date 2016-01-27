//  Copyright (c) 2015 Couchbase, Inc.

package gson

import "fmt"
import "testing"

var _ = fmt.Sprintf("dummy")

// All test cases are folded into collate_cbor_test.go, contains only few
// missing testcases (if any) and benchmarks.

func BenchmarkCbor2CollNil(b *testing.B) {
	config := NewDefaultConfig()
	jsn := config.NewJson([]byte("null"), -1)
	cbr := config.NewCbor(make([]byte, 1024), 0)
	clt := config.NewCollate(make([]byte, 1024), 0)

	jsn.Tocbor(cbr)

	for i := 0; i < b.N; i++ {
		cbr.Tocollate(clt.Reset(nil))
	}
}

func BenchmarkCbor2CollTrue(b *testing.B) {
	config := NewDefaultConfig()
	jsn := config.NewJson([]byte("true"), -1)
	cbr := config.NewCbor(make([]byte, 1024), 0)
	clt := config.NewCollate(make([]byte, 1024), 0)

	jsn.Tocbor(cbr)

	for i := 0; i < b.N; i++ {
		cbr.Tocollate(clt.Reset(nil))
	}
}

func BenchmarkCbor2CollFalse(b *testing.B) {
	config := NewDefaultConfig()
	jsn := config.NewJson([]byte("false"), -1)
	cbr := config.NewCbor(make([]byte, 1024), 0)
	clt := config.NewCollate(make([]byte, 1024), 0)

	jsn.Tocbor(cbr)

	for i := 0; i < b.N; i++ {
		cbr.Tocollate(clt.Reset(nil))
	}
}

func BenchmarkCbor2CollF64(b *testing.B) {
	config := NewDefaultConfig()
	jsn := config.NewJson([]byte("10.121312213123123"), -1)
	cbr := config.NewCbor(make([]byte, 1024), 0)
	clt := config.NewCollate(make([]byte, 1024), 0)

	jsn.Tocbor(cbr)

	for i := 0; i < b.N; i++ {
		cbr.Tocollate(clt.Reset(nil))
	}
}

func BenchmarkCbor2CollI64(b *testing.B) {
	config := NewDefaultConfig()
	jsn := config.NewJson([]byte("123456789"), -1)
	cbr := config.NewCbor(make([]byte, 1024), 0)
	clt := config.NewCollate(make([]byte, 1024), 0)

	jsn.Tocbor(cbr)

	for i := 0; i < b.N; i++ {
		cbr.Tocollate(clt.Reset(nil))
	}
}

func BenchmarkCbor2CollMiss(b *testing.B) {
	config := NewDefaultConfig()
	jsn := config.NewJson([]byte(fmt.Sprintf(`"%s"`, MissingLiteral)), -1)
	cbr := config.NewCbor(make([]byte, 1024), 0)
	clt := config.NewCollate(make([]byte, 1024), 0)

	jsn.Tocbor(cbr)

	for i := 0; i < b.N; i++ {
		cbr.Tocollate(clt.Reset(nil))
	}
}

func BenchmarkCbor2CollStr(b *testing.B) {
	config := NewDefaultConfig()
	jsn := config.NewJson([]byte(`"hello world"`), -1)
	cbr := config.NewCbor(make([]byte, 1024), 0)
	clt := config.NewCollate(make([]byte, 1024), 0)

	jsn.Tocbor(cbr)

	for i := 0; i < b.N; i++ {
		cbr.Tocollate(clt.Reset(nil))
	}
}

func BenchmarkCbor2CollArr(b *testing.B) {
	in := []byte(`[null,true,false,"hello world",10.23122312]`)

	config := NewDefaultConfig()
	jsn := config.NewJson(in, -1)
	cbr := config.NewCbor(make([]byte, 1024), 0)
	clt := config.NewCollate(make([]byte, 1024), 0)

	jsn.Tocbor(cbr)

	for i := 0; i < b.N; i++ {
		cbr.Tocollate(clt.Reset(nil))
	}
}

func BenchmarkCbor2CollMap(b *testing.B) {
	inp := `{"key1":null,"key2":true,"key3":false,"key4":"hello world",` +
		`"key5":10.23122312}`
	config := NewDefaultConfig().SetMaxkeys(10)
	jsn := config.NewJson([]byte(inp), -1)
	cbr := config.NewCbor(make([]byte, 1024), 0)
	clt := config.NewCollate(make([]byte, 1024), 0)

	jsn.Tocbor(cbr)

	for i := 0; i < b.N; i++ {
		cbr.Tocollate(clt.Reset(nil))
	}
}

func BenchmarkCbor2CollTyp(b *testing.B) {
	data := testdataFile("testdata/typical.json")

	config := NewDefaultConfig().SetMaxkeys(10)
	jsn := config.NewJson(data, -1)
	cbr := config.NewCbor(make([]byte, 10*1024), 0)
	clt := config.NewCollate(make([]byte, 10*1024), 0)

	jsn.Tocbor(cbr)

	for i := 0; i < b.N; i++ {
		cbr.Tocollate(clt.Reset(nil))
	}
}
