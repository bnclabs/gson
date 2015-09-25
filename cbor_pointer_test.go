//  Copyright (c) 2015 Couchbase, Inc.

package gson

import "testing"
import "fmt"

var _ = fmt.Sprintf("dummy")

func TestCborPointer(t *testing.T) {
	tcasesJSONPointers := []string{
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
	buf, out := make([]byte, 1024), make([]byte, 1024)
	config := NewDefaultConfig()
	for _, tcase := range tcasesJSONPointers {
		t.Logf(tcase)
		n := config.JsonPointerToCbor([]byte(tcase), buf)
		m := config.CborToJsonPointer(buf[:n], out)
		if result := string(out[:m]); tcase != result {
			t.Errorf("expected %q, got %q", tcase, result)
		}
	}
}

func BenchmarkJsonPtr2Cbor3(b *testing.B) {
	config := NewDefaultConfig()
	jsonptr := []byte("/foo/g/0")
	out := make([]byte, 1024)
	b.SetBytes(int64(len(jsonptr)))
	for i := 0; i < b.N; i++ {
		config.JsonPointerToCbor(jsonptr, out)
	}
}

func BenchmarkJsonPtr2Cbor4(b *testing.B) {
	config := NewDefaultConfig()
	jsonptr := []byte("/foo/g~1n~1r/0/hello")
	out := make([]byte, 1024)
	b.SetBytes(int64(len(jsonptr)))
	for i := 0; i < b.N; i++ {
		config.JsonPointerToCbor(jsonptr, out)
	}
}

func BenchmarkJsonPtrCbor5(b *testing.B) {
	config, out := NewDefaultConfig(), make([]byte, 1024)
	n := config.ToJsonPointer([]string{"/a", "ab", "a~b", "a/b", "a~/~/b"}, out)
	jsonptr := make([]byte, 1024)
	copy(jsonptr, out[:n])

	b.SetBytes(int64(n))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config.JsonPointerToCbor(jsonptr[:n], out)
	}
}

func BenchmarkCborJsonPtr5(b *testing.B) {
	config, out := NewDefaultConfig(), make([]byte, 1024)
	n := config.ToJsonPointer([]string{"/a", "ab", "a~b", "a/b", "a~/~/b"}, out)
	jsonptr := make([]byte, 1024)
	copy(jsonptr, out[:n]) // shall copy out to new location

	m := config.JsonPointerToCbor(jsonptr[:n], out)

	jsonout := make([]byte, 1024)
	b.SetBytes(int64(m))
	for i := 0; i < b.N; i++ {
		config.CborToJsonPointer(out[:m], jsonout)
	}
}
