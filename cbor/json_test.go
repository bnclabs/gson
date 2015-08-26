package cbor

import "testing"
import "fmt"
import "bytes"
import "reflect"
import "encoding/json"

var _ = fmt.Sprintf("dummy")

func TestSkipWS(t *testing.T) {
	ref := "hello  "
	if got := skipWS("  hello  ", AnsiSpace); got != ref {
		t.Errorf("expected %v got %v", ref, got)
	}
}

func TestScanEmpty(t *testing.T) {
	config := NewDefaultConfig()
	out := make([]byte, 1024)
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic")
		}
	}()
	scanToken("", out, config)
}

func TestJson(t *testing.T) {
	testcases := []string{
		// null
		"null",
		// boolean
		"true",
		"false",
		// integers
		"10",
		"0.1",
		"-0.1",
		"10.1",
		"-10.1",
		"-10E-1",
		"-10e+1",
		"10E-1",
		"10e+1",
		// string
		`"true"`,
		`"tru\"e"`,
		`"tru\\e"`,
		`"tru\be"`,
		`"tru\fe"`,
		`"tru\ne"`,
		`"tru\re"`,
		`"tru\te"`,
		`"tru\u0123e"`,
		`"汉语 / 漢語; Hàn\b \t\uef24yǔ "`,
		// array
		`[]`,
		` [null,true,false,10,"tru\"e"]`,
		// object
		`{}`,
		`{"a":null,"b":true,"c":false,"d\"":10,"e":"tru\"e", "f":[1,2]}`,
	}
	cborout, jsonout := make([]byte, 1024), make([]byte, 1024)
	config := NewDefaultConfig()
	var ref1, ref2 interface{}
	for _, tcase := range testcases {
		t.Logf("testcase - %v", tcase)
		_, n := config.ParseJson(tcase, cborout)
		if err := json.Unmarshal([]byte(tcase), &ref1); err != nil {
			t.Errorf("json.Unmarshal() failed for tcase %v: %v", tcase, err)
		}
		t.Logf("%v %v", cborout[:n], n)
		p, m := config.ToJson(cborout, jsonout)
		if p != n {
			t.Errorf("expected %v, got %v", n, p)
		}
		if err := json.Unmarshal(jsonout[:m], &ref2); err != nil {
			t.Errorf("json.Unmarshal() failed for cbor %v: %v", tcase, err)
		}
		if !reflect.DeepEqual(ref1, ref2) {
			t.Errorf("mismatch %v, got %v", ref1, ref2)
		}
	}
}

func TestScanNumber(t *testing.T) {
	code, out := make([]byte, 1024), make([]byte, 1024)
	// test JsonNumber
	ref := []byte{216, 38, 98, 49, 48}
	_, n := scanNum("10", JsonNumber, code)
	if bytes.Compare(code[:n], ref) != 0 {
		t.Errorf("expected %v, got %v", ref, code[:n])
	}
	// test FloatNumber
	_, n = scanNum("10", FloatNumber, code)
	_, y := decodeTojson(code[:n], out)
	if s := string(out[:y]); s != "10.00000000000000000000" {
		t.Errorf("expected %q, got %q", "10.00000000000000000000", s)
	}
	// test IntNumber
	_, n = scanNum("10", IntNumber, code)
	_, y = decodeTojson(code[:n], out)
	if s := string(out[:y]); s != "10" {
		t.Errorf("expected %q, got %q", "10", s)
	}
	// malformed IntNumber
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		scanNum("10.2", IntNumber, out)
	}()
	// test FloatNumber32
	_, n = scanNum("10", FloatNumber32, code)
	_, y = decodeTojson(code[:n], out)
	if s := string(out[:y]); s != "10.000000" {
		t.Errorf("expected %q, got %q", "10.000000", s)
	}
	// test SmartNumber32
	_, n = scanNum("10.2", SmartNumber32, code)
	_, y = decodeTojson(code[:n], out)
	if s := string(out[:y]); s != "10.200000" {
		t.Errorf("expected %q, got %q", "10.200000", s)
	}
	// test SmartNumber32 (integer)
	_, n = scanNum("10", SmartNumber32, code)
	_, y = decodeTojson(code[:n], out)
	if s := string(out[:y]); s != "10" {
		t.Errorf("expected %q, got %q", "10", s)
	}
	// test SmartNumber
	_, n = scanNum("10.2", SmartNumber, code)
	_, y = decodeTojson(code[:n], out)
	if s := string(out[:y]); s != "10.19999999999999928946" {
		t.Errorf("expected %q, got %q", "10.19999999999999928946", s)
	}
	// test SmartNumber (integer)
	_, n = scanNum("10", SmartNumber32, code)
	_, y = decodeTojson(code[:n], out)
	if s := string(out[:y]); s != "10" {
		t.Errorf("expected %q, got %q", "10", s)
	}
}

func TestJsonNumber(t *testing.T) {
	// for number as integer.
	testcases := []string{
		"255", "256", "-255", "-256", "65535", "65536", "-65535", "-65536",
		"4294967295", "4294967296", "-4294967295", "-4294967296",
		"9223372036854775807", "-9223372036854775807", "-9223372036854775808",
	}
	cborout, jsonout := make([]byte, 1024), make([]byte, 1024)
	config := NewConfig(IntNumber, UnicodeSpace, Stream)
	var ref1, ref2 interface{}
	for _, tcase := range testcases {
		t.Logf("testcase - %v", tcase)
		_, n := config.ParseJson(tcase, cborout)
		if err := json.Unmarshal([]byte(tcase), &ref1); err != nil {
			t.Errorf("json.Unmarshal() failed for tcase %v: %v", tcase, err)
		}
		t.Logf("%v %v", cborout[:n], n)
		p, m := config.ToJson(cborout, jsonout)
		if p != n {
			t.Errorf("expected %v, got %v", n, p)
		}
		if err := json.Unmarshal(jsonout[:m], &ref2); err != nil {
			t.Errorf("json.Unmarshal() failed for cbor %v: %v", tcase, err)
		}
		if !reflect.DeepEqual(ref1, ref2) {
			t.Errorf("mismatch %v, got %v", ref1, ref2)
		}
	}
	out := make([]byte, 64)
	// test float-number
	tcase := "10.2"
	config = NewDefaultConfig()
	_, n := config.ParseJson(tcase, cborout)
	if err := json.Unmarshal([]byte(tcase), &ref1); err != nil {
		t.Errorf("json.Unmarshal() failed for tcase %v: %v", tcase, err)
	}
	t.Logf("%v %v", cborout[:n], n)
	p, m := config.ToJson(cborout, jsonout)
	if p != n {
		t.Errorf("expected %v, got %v", n, p)
	}
	if err := json.Unmarshal(jsonout[:m], &ref2); err != nil {
		t.Errorf("json.Unmarshal() failed for cbor %v: %v", tcase, err)
	}
	if !reflect.DeepEqual(ref1, ref2) {
		t.Errorf("mismatch %v, got %v", ref1, ref2)
	}
	// negative small integers
	buf := make([]byte, 64)
	n = encodeInt8(-1, buf)
	_, m = decodeType1SmallIntTojson(buf, out)
	if v := string(out[:m]); v != "-1" {
		t.Errorf("expected -1, got %v", v)
	}
}

func TestScanBadToken(t *testing.T) {
	out := make([]byte, 64)
	panicfn := func(in string, config *Config) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		scanToken(in, out, config)
	}
	testcases := []string{
		"    ",
		"nil",
		"treu",
		"fale",
		"[  ",
		"[10  ",
		"[10,  ",
		"[10true",
		"{10",
		`{"10"true`,
		`{"10":true  `,
		`{"10":true10`,
		`(`,
		`"`,
	}
	config := NewDefaultConfig()
	for _, tcase := range testcases {
		panicfn(tcase, config)
	}
	// test ScanLengthPrefix for array
	config = NewConfig(FloatNumber, UnicodeSpace, LengthPrefix)
	panicfn("[]", config)
	// test ScanLengthPrefix for property
	config = NewConfig(FloatNumber, UnicodeSpace, LengthPrefix)
	panicfn("{}", config)
}

func TestFloat32(t *testing.T) {
	var ref1, ref2 interface{}

	buf, out := make([]byte, 64), make([]byte, 64)
	n := encodeFloat32(float32(10.2), buf)
	if err := json.Unmarshal([]byte("10.2"), &ref1); err != nil {
		t.Errorf("json.Unmarshal() failed for %v: %v", buf[:n], err)
	}

	_, m := decodeFloat32Tojson(buf, out)
	t.Logf("json - %v", string(out[:m]))
	if err := json.Unmarshal(out[:m], &ref2); err != nil {
		t.Errorf("json.Unmarshal() failed for cbor %v: %v", buf[:n], err)
	}
	if !reflect.DeepEqual(ref1, ref2) {
		t.Errorf("mismatch %v, got %v", ref1, ref2)
	}
}

func TestByteString(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic")
		}
	}()
	buf, out := make([]byte, 16), make([]byte, 16)
	n := encodeBytes([]byte{0xf5}, buf)
	NewDefaultConfig().ToJson(buf[:n], out)
}

func BenchmarkParseJsonN(b *testing.B) {
	out := make([]byte, 1024)
	in := "null"
	config := NewDefaultConfig()
	b.SetBytes(int64(len(in)))
	for i := 0; i < b.N; i++ {
		config.ParseJson(in, out)
	}
}

func BenchmarkToJsonN(b *testing.B) {
	buf, out := make([]byte, 1024), make([]byte, 1024)
	config := NewDefaultConfig()
	_, n := config.ParseJson("null", buf)
	b.SetBytes(int64(n))
	for i := 0; i < b.N; i++ {
		config.ToJson(buf[:n], out)
	}
}

func BenchmarkParseJsonI(b *testing.B) {
	out := make([]byte, 1024)
	in := "123456567"
	config := NewDefaultConfig()
	b.SetBytes(int64(len(in)))
	for i := 0; i < b.N; i++ {
		config.ParseJson(in, out)
	}
}

func BenchmarkToJsonI(b *testing.B) {
	buf, out := make([]byte, 1024), make([]byte, 1024)
	config := NewDefaultConfig()
	_, n := config.ParseJson("123456567", buf)
	b.SetBytes(int64(n))
	for i := 0; i < b.N; i++ {
		config.ToJson(buf[:n], out)
	}
}

func BenchmarkParseJsonF(b *testing.B) {
	out := make([]byte, 1024)
	in := "1234.12312"
	config := NewDefaultConfig()
	b.SetBytes(int64(len(in)))
	for i := 0; i < b.N; i++ {
		config.ParseJson(in, out)
	}
}

func BenchmarkToJsonF(b *testing.B) {
	buf, out := make([]byte, 1024), make([]byte, 1024)
	config := NewDefaultConfig()
	_, n := config.ParseJson("1234.12312", buf)
	b.SetBytes(int64(n))
	for i := 0; i < b.N; i++ {
		config.ToJson(buf[:n], out)
	}
}

func BenchmarkParseJsonB(b *testing.B) {
	out := make([]byte, 1024)
	in := "false"
	config := NewDefaultConfig()
	b.SetBytes(int64(len(in)))
	for i := 0; i < b.N; i++ {
		config.ParseJson(in, out)
	}
}

func BenchmarkToJsonB(b *testing.B) {
	buf, out := make([]byte, 1024), make([]byte, 1024)
	config := NewDefaultConfig()
	_, n := config.ParseJson("false", buf)
	b.SetBytes(int64(n))
	for i := 0; i < b.N; i++ {
		config.ToJson(buf[:n], out)
	}
}

func BenchmarkParseJsonS(b *testing.B) {
	out := make([]byte, 1024)
	in := `"汉语 / 漢語; Hàn\b \t\uef24yǔ "`
	config := NewDefaultConfig()
	b.SetBytes(int64(len(in)))
	for i := 0; i < b.N; i++ {
		config.ParseJson(in, out)
	}
}

func BenchmarkToJsonS(b *testing.B) {
	buf, out := make([]byte, 1024), make([]byte, 1024)
	config := NewDefaultConfig()
	_, n := config.ParseJson(`"汉语 / 漢語; Hàn\b \t\uef24yǔ "`, buf)
	b.SetBytes(int64(n))
	for i := 0; i < b.N; i++ {
		config.ToJson(buf[:n], out)
	}
}

func BenchmarkParseJsonA(b *testing.B) {
	out := make([]byte, 1024)
	in := ` [null,true,false,10,"tru\"e"]`
	config := NewDefaultConfig()
	b.SetBytes(int64(len(in)))
	for i := 0; i < b.N; i++ {
		config.ParseJson(in, out)
	}
}

func BenchmarkToJsonA(b *testing.B) {
	buf, out := make([]byte, 1024), make([]byte, 1024)
	config := NewDefaultConfig()
	_, n := config.ParseJson(` [null,true,false,10,"tru\"e"]`, buf)
	b.SetBytes(int64(n))
	for i := 0; i < b.N; i++ {
		config.ToJson(buf[:n], out)
	}
}

func BenchmarkParseJsonM(b *testing.B) {
	out := make([]byte, 1024)
	in := `{"a":null,"b":true,"c":false,"d\"":10,"e":"tru\"e", "f":[1,2]}`
	config := NewDefaultConfig()
	b.SetBytes(int64(len(in)))
	for i := 0; i < b.N; i++ {
		config.ParseJson(in, out)
	}
}
func BenchmarkToJsonM(b *testing.B) {
	buf, out := make([]byte, 1024), make([]byte, 1024)
	in := `{"a":null,"b":true,"c":false,"d\"":10,"e":"tru\"e", "f":[1,2]}`
	config := NewDefaultConfig()
	_, n := config.ParseJson(in, buf)
	b.SetBytes(int64(n))
	for i := 0; i < b.N; i++ {
		config.ToJson(buf[:n], out)
	}
}
