package gson

import (
	"testing"
)

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
		segments := config.ParsePointer(tcase.in)
		if len(segments) != len(tcase.ref) {
			t.Errorf("expected %v, got %v", len(tcase.ref), len(segments))
		} else {
			for i, x := range tcase.ref {
				if string(segments[i]) != string(x) {
					t.Errorf("expected %q, got %q", tcase.ref, segments[i])
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

func BenchmarkParsePtrSmall(b *testing.B) {
	config := NewDefaultConfig()
	path := "/foo/g/0"
	for i := 0; i < b.N; i++ {
		config.ParsePointer(path)
	}
}

func BenchmarkParsePtrMed(b *testing.B) {
	config := NewDefaultConfig()
	path := "/foo/g~1n~1r/0/hello"
	for i := 0; i < b.N; i++ {
		config.ParsePointer(path)
	}
}

func BenchmarkParsePtrLarge(b *testing.B) {
	config := NewDefaultConfig()
	out := make([]byte, 1024)
	n := config.EncodePointer([]string{"a", "ab", "a~b", "a/b", "a~/~/b"}, out)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config.ParsePointer(string(out[:n]))
	}
}

func BenchmarkEncPtrLarge(b *testing.B) {
	config := NewDefaultConfig()
	path := []string{"a", "ab", "a~b", "a/b", "a~/~/b"}
	out := make([]byte, 1024)
	for i := 0; i < b.N; i++ {
		config.EncodePointer(path, out)
	}
}
