package json

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
    for _, tcase := range tcasesJSONPointers {
        parts := parsePointer([]byte(tcase.in))
        if len(parts) != len(tcase.ref) {
            t.Errorf("failed on %q %v", tcase.in, parts)
        } else {
            for i, x := range tcase.ref {
                if string(parts[i]) != string(x) {
                    t.Errorf("failed on %q %q", tcase.in, parts[i])
                }
            }
        }
    }
}

func TestEncodePointer(t *testing.T) {
    out := make([]byte, 0, 1024)
    for _, tcase := range tcasesJSONPointers {
        outs := encodePointer(tcase.ref, out)
        if outs != tcase.in {
            t.Errorf("failed on %q %q", tcase.in, outs)
        }
    }
}

func BenchmarkParsePointerSmall(b *testing.B) {
    path := "/foo/g/0"
    for i := 0; i < b.N; i++ {
        parsePointer([]byte(path))
    }
}

func BenchmarkParsePointerMedium(b *testing.B) {
    path := "/foo/g~1n~1r/0/hello"
    for i := 0; i < b.N; i++ {
        parsePointer([]byte(path))
    }
}

func BenchmarkParsePointerLarge(b *testing.B) {
    out := make([]byte, 0, 1024)
    path := encodePointer([]string{"a", "ab", "a~b", "a/b", "a~/~/b"}, out)
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        parsePointer([]byte(path))
    }
}

func BenchmarkEncodePointerLarge(b *testing.B) {
    path := []string{"a", "ab", "a~b", "a/b", "a~/~/b"}
    out := make([]byte, 0, 1024)
    for i := 0; i < b.N; i++ {
        encodePointer(path, out)
    }
}
