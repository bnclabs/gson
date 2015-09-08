package gson

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

func TestJsonEmptyToCbor(t *testing.T) {
	config := NewDefaultConfig()
	out := make([]byte, 1024)
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic")
		}
	}()
	json2cbor("", out, config)
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
		_, n := config.JsonToCbor(tcase, cborout)
		if err := json.Unmarshal([]byte(tcase), &ref1); err != nil {
			t.Errorf("json.Unmarshal() failed for tcase %v: %v", tcase, err)
		}
		t.Logf("%v %v", cborout[:n], n)
		p, m := config.CborToJson(cborout, jsonout)
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
	config := NewConfig(JsonNumber, UnicodeSpace)
	ref := []byte{216, 38, 98, 49, 48}
	_, n := config.JsonToCbor("10", code)
	if bytes.Compare(code[:n], ref) != 0 {
		t.Errorf("expected %v, got %v", ref, code[:n])
	}
	// test FloatNumber
	config = NewConfig(FloatNumber, UnicodeSpace)
	_, n = config.JsonToCbor("10", code)
	_, y := config.CborToJson(code[:n], out)
	if s := string(out[:y]); s != "10.00000000000000000000" {
		t.Errorf("expected %q, got %q", "10.00000000000000000000", s)
	}
	// test IntNumber
	config = NewConfig(IntNumber, UnicodeSpace)
	_, n = config.JsonToCbor("10", code)
	_, y = config.CborToJson(code[:n], out)
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
		config = NewConfig(IntNumber, UnicodeSpace)
		config.JsonToCbor("10.2", out)
	}()
	// test FloatNumber32
	config = NewConfig(FloatNumber32, UnicodeSpace)
	_, n = config.JsonToCbor("10", code)
	_, y = config.CborToJson(code[:n], out)
	if s := string(out[:y]); s != "10.000000" {
		t.Errorf("expected %q, got %q", "10.000000", s)
	}
	// test SmartNumber32
	config = NewConfig(SmartNumber32, UnicodeSpace)
	_, n = config.JsonToCbor("10.2", code)
	_, y = config.CborToJson(code[:n], out)
	if s := string(out[:y]); s != "10.200000" {
		t.Errorf("expected %q, got %q", "10.200000", s)
	}
	// test SmartNumber32 (integer)
	config = NewConfig(SmartNumber32, UnicodeSpace)
	_, n = config.JsonToCbor("10", code)
	_, y = config.CborToJson(code[:n], out)
	if s := string(out[:y]); s != "10" {
		t.Errorf("expected %q, got %q", "10", s)
	}
	// test SmartNumber
	config = NewConfig(SmartNumber, UnicodeSpace)
	_, n = config.JsonToCbor("10.2", code)
	_, y = config.CborToJson(code[:n], out)
	if s := string(out[:y]); s != "10.19999999999999928946" {
		t.Errorf("expected %q, got %q", "10.19999999999999928946", s)
	}
	// test SmartNumber (integer)
	config = NewConfig(SmartNumber32, UnicodeSpace)
	_, n = config.JsonToCbor("10", code)
	_, y = config.CborToJson(code[:n], out)
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
	config := NewConfig(IntNumber, UnicodeSpace)
	var ref1, ref2 interface{}
	for _, tcase := range testcases {
		t.Logf("testcase - %v", tcase)
		_, n := config.JsonToCbor(tcase, cborout)
		if err := json.Unmarshal([]byte(tcase), &ref1); err != nil {
			t.Errorf("json.Unmarshal() failed for tcase %v: %v", tcase, err)
		}
		t.Logf("%v %v", cborout[:n], n)
		p, m := config.CborToJson(cborout, jsonout)
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
	_, n := config.JsonToCbor(tcase, cborout)
	if err := json.Unmarshal([]byte(tcase), &ref1); err != nil {
		t.Errorf("json.Unmarshal() failed for tcase %v: %v", tcase, err)
	}
	t.Logf("%v %v", cborout[:n], n)
	p, m := config.CborToJson(cborout, jsonout)
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
	n = valint82cbor(-1, buf)
	_, m = cbor2jsont1smallint(buf, out, config)
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
		json2cbor(in, out, config)
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
	config = NewConfig(FloatNumber, UnicodeSpace)
	config = config.ContainerEncoding(LengthPrefix)
	panicfn("[]", config)
	// test ScanLengthPrefix for property
	config = NewConfig(FloatNumber, UnicodeSpace)
	config = config.ContainerEncoding(LengthPrefix)
	panicfn("{}", config)
}

func TestFloat32(t *testing.T) {
	var ref1, ref2 interface{}

	config := NewDefaultConfig()

	buf, out := make([]byte, 64), make([]byte, 64)
	n := valfloat322cbor(float32(10.2), buf)
	if err := json.Unmarshal([]byte("10.2"), &ref1); err != nil {
		t.Errorf("json.Unmarshal() failed for %v: %v", buf[:n], err)
	}

	_, m := cbor2jsonfloat32(buf, out, config)
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
	n := valbytes2cbor([]byte{0xf5}, buf)
	NewDefaultConfig().CborToJson(buf[:n], out)
}

//---- benchmarks

func BenchmarkJson2CborNull(b *testing.B) {
	out := make([]byte, 1024)
	in := "null"
	config := NewDefaultConfig()
	b.SetBytes(int64(len(in)))
	for i := 0; i < b.N; i++ {
		config.JsonToCbor(in, out)
	}
}

func BenchmarkCbor2JsonNull(b *testing.B) {
	buf, out := make([]byte, 1024), make([]byte, 1024)
	config := NewDefaultConfig()
	_, n := config.JsonToCbor("null", buf)
	b.SetBytes(int64(n))
	for i := 0; i < b.N; i++ {
		config.CborToJson(buf[:n], out)
	}
}

func BenchmarkJson2CborInt(b *testing.B) {
	out := make([]byte, 1024)
	in := "123456567"
	config := NewDefaultConfig()
	b.SetBytes(int64(len(in)))
	for i := 0; i < b.N; i++ {
		config.JsonToCbor(in, out)
	}
}

func BenchmarkCbor2JsonInt(b *testing.B) {
	buf, out := make([]byte, 1024), make([]byte, 1024)
	config := NewDefaultConfig()
	_, n := config.JsonToCbor("123456567", buf)
	b.SetBytes(int64(n))
	for i := 0; i < b.N; i++ {
		config.CborToJson(buf[:n], out)
	}
}

func BenchmarkJson2CborFlt(b *testing.B) {
	out := make([]byte, 1024)
	in := "1234.12312"
	config := NewDefaultConfig()
	b.SetBytes(int64(len(in)))
	for i := 0; i < b.N; i++ {
		config.JsonToCbor(in, out)
	}
}

func BenchmarkCbor2JsonFlt(b *testing.B) {
	buf, out := make([]byte, 1024), make([]byte, 1024)
	config := NewDefaultConfig()
	_, n := config.JsonToCbor("1234.12312", buf)
	b.SetBytes(int64(n))
	for i := 0; i < b.N; i++ {
		config.CborToJson(buf[:n], out)
	}
}

func BenchmarkJson2CborBool(b *testing.B) {
	out := make([]byte, 1024)
	in := "false"
	config := NewDefaultConfig()
	b.SetBytes(int64(len(in)))
	for i := 0; i < b.N; i++ {
		config.JsonToCbor(in, out)
	}
}

func BenchmarkCbor2JsonBool(b *testing.B) {
	buf, out := make([]byte, 1024), make([]byte, 1024)
	config := NewDefaultConfig()
	_, n := config.JsonToCbor("false", buf)
	b.SetBytes(int64(n))
	for i := 0; i < b.N; i++ {
		config.CborToJson(buf[:n], out)
	}
}

func BenchmarkJson2CborStr(b *testing.B) {
	out := make([]byte, 1024)
	in := `"汉语 / 漢語; Hàn\b \t\uef24yǔ "`
	config := NewDefaultConfig()
	b.SetBytes(int64(len(in)))
	for i := 0; i < b.N; i++ {
		config.JsonToCbor(in, out)
	}
}

func BenchmarkCbor2JsonStr(b *testing.B) {
	buf, out := make([]byte, 1024), make([]byte, 1024)
	config := NewDefaultConfig()
	_, n := config.JsonToCbor(`"汉语 / 漢語; Hàn\b \t\uef24yǔ "`, buf)
	b.SetBytes(int64(n))
	for i := 0; i < b.N; i++ {
		config.CborToJson(buf[:n], out)
	}
}

func BenchmarkJson2CborArr(b *testing.B) {
	out := make([]byte, 1024)
	in := ` [null,true,false,10,"tru\"e"]`
	config := NewDefaultConfig()
	b.SetBytes(int64(len(in)))
	for i := 0; i < b.N; i++ {
		config.JsonToCbor(in, out)
	}
}

func BenchmarkCbor2JsonArr(b *testing.B) {
	buf, out := make([]byte, 1024), make([]byte, 1024)
	config := NewDefaultConfig()
	_, n := config.JsonToCbor(` [null,true,false,10,"tru\"e"]`, buf)
	b.SetBytes(int64(n))
	for i := 0; i < b.N; i++ {
		config.CborToJson(buf[:n], out)
	}
}

func BenchmarkJson2CborMap(b *testing.B) {
	out := make([]byte, 1024)
	in := `{"a":null,"b":true,"c":false,"d\"":10,"e":"tru\"e", "f":[1,2]}`
	config := NewDefaultConfig()
	b.SetBytes(int64(len(in)))
	for i := 0; i < b.N; i++ {
		config.JsonToCbor(in, out)
	}
}
func BenchmarkCbor2JsonMap(b *testing.B) {
	buf, out := make([]byte, 1024), make([]byte, 1024)
	in := `{"a":null,"b":true,"c":false,"d\"":10,"e":"tru\"e", "f":[1,2]}`
	config := NewDefaultConfig()
	_, n := config.JsonToCbor(in, buf)
	b.SetBytes(int64(n))
	for i := 0; i < b.N; i++ {
		config.CborToJson(buf[:n], out)
	}
}
