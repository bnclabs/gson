//  Copyright (c) 2015 Couchbase, Inc.

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

	config := NewDefaultConfig()

	jsn := config.NewJson(make([]byte, 1024), 0)
	cbr := config.NewCbor(make([]byte, 1024), 0)
	jsnback := config.NewJson(make([]byte, 1024), 0)

	var ref1, ref2 interface{}

	for _, tcase := range testcases {
		t.Logf("testcase - %v", tcase)
		json.Unmarshal([]byte(tcase), &ref1)

		jsn.Reset([]byte(tcase))
		cbr.Reset(nil)
		jsn.Tocbor(cbr)
		t.Logf("%v %v", len(cbr.Bytes()), cbr.Bytes())

		cbr.Tojson(jsnback.Reset(nil))
		if err := json.Unmarshal(jsnback.Bytes(), &ref2); err != nil {
			t.Errorf("json.Unmarshal() failed for cbor %v: %v", tcase, err)
		}

		if !reflect.DeepEqual(ref1, ref2) {
			t.Errorf("mismatch %v, got %v", ref1, ref2)
		}
	}
}

func TestCbor2JsonLengthPrefix(t *testing.T) {
	testcases := []string{
		`[null,true,false,10,"tru\"e"]`,
		`[]`,
		`{}`,
		`{"a":null,"b":true,"c":false,"d\"":10,"e":"tru\"e","f":[1,2]}`,
	}

	config := NewDefaultConfig().NumberKind(IntNumber)
	config = config.ContainerEncoding(LengthPrefix)

	jsn := config.NewJson(make([]byte, 1024), 0)
	cbr := config.NewCbor(make([]byte, 1024), 0)
	jsnback := config.NewJson(make([]byte, 1024), 0)

	for _, tcase := range testcases {
		t.Logf("testcase - %v", tcase)
		jsn.Reset([]byte(tcase))

		cbr.Reset(nil)
		jsn.Tocbor(cbr)

		jsnback.Reset(nil)
		cbr.Tojson(jsnback)

		err := compare_jsons(t, tcase, string(jsnback.Bytes()))
		if err != nil {
			t.Errorf("%v", err)
		}
	}
}

func TestScanNumber(t *testing.T) {
	// test JSONNumber
	ref := []byte{216, 38, 98, 49, 48}
	config := NewDefaultConfig().NumberKind(JSONNumber).SpaceKind(UnicodeSpace)
	cbr := config.NewCbor(make([]byte, 1024), 0)
	jsn := config.NewJson(make([]byte, 1024), 0)

	jsn.Reset([]byte("10"))
	jsn.Tocbor(cbr)
	if code := cbr.Bytes(); bytes.Compare(code, ref) != 0 {
		t.Errorf("expected %v, got %v", ref, code)
	}

	jsn.Reset(nil)
	cbr.Tojson(jsn)
	if s := string(jsn.Bytes()); s != "10" {
		t.Errorf("exected %v, got %v", "10", s)
	}

	// test FloatNumber
	config = NewDefaultConfig().NumberKind(FloatNumber).SpaceKind(UnicodeSpace)
	cbr = config.NewCbor(make([]byte, 1024), 0)
	jsn = config.NewJson(make([]byte, 1024), 0)
	jsnback := config.NewJson(make([]byte, 1024), 0)

	jsn.Reset([]byte("10"))
	jsn.Tocbor(cbr)
	cbr.Tojson(jsnback)
	if s := string(jsnback.Bytes()); s != "10" {
		t.Errorf("expected %q, got %q", "10", s)
	}

	// test IntNumber
	config = NewDefaultConfig().NumberKind(IntNumber).SpaceKind(UnicodeSpace)
	cbr = config.NewCbor(make([]byte, 1024), 0)
	jsn = config.NewJson(make([]byte, 1024), 0)
	jsnback = config.NewJson(make([]byte, 1024), 0)

	jsn.Reset([]byte("10"))
	jsn.Tocbor(cbr)
	cbr.Tojson(jsnback)
	if s := string(jsnback.Bytes()); s != "10" {
		t.Errorf("expected %q, got %q", "10", s)
	}

	// malformed IntNumber
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		config := NewDefaultConfig().NumberKind(IntNumber).SpaceKind(UnicodeSpace)
		jsn := config.NewJson(make([]byte, 1024), 0)
		cbr := config.NewCbor(make([]byte, 1024), 0)
		jsn.Reset([]byte("10.2"))
		jsn.Tocbor(cbr)
	}()

	// test FloatNumber32
	config = NewDefaultConfig().NumberKind(FloatNumber32).SpaceKind(UnicodeSpace)
	jsn = config.NewJson(make([]byte, 1024), 0)
	cbr = config.NewCbor(make([]byte, 1024), 0)
	jsnback = config.NewJson(make([]byte, 1024), 0)

	jsn.Reset([]byte("10"))
	jsn.Tocbor(cbr)
	cbr.Tojson(jsnback)
	if s := string(jsnback.Bytes()); s != "10" {
		t.Errorf("expected %q, got %q", "10", s)
	}

	// test SmartNumber32
	config = NewDefaultConfig().NumberKind(SmartNumber32).SpaceKind(UnicodeSpace)
	jsn = config.NewJson(make([]byte, 1024), 0)
	cbr = config.NewCbor(make([]byte, 1024), 0)
	jsnback = config.NewJson(make([]byte, 1024), 0)

	jsn.Reset([]byte("10.2"))
	jsn.Tocbor(cbr)
	cbr.Tojson(jsnback)
	if s := string(jsnback.Bytes()); s != "10.2" {
		t.Errorf("expected %q, got %q", "10.2", s)
	}

	// test SmartNumber32 (integer)
	config = NewDefaultConfig().NumberKind(SmartNumber32).SpaceKind(UnicodeSpace)
	jsn = config.NewJson(make([]byte, 1024), 0)
	cbr = config.NewCbor(make([]byte, 1024), 0)
	jsnback = config.NewJson(make([]byte, 1024), 0)

	jsn.Reset([]byte("10"))
	jsn.Tocbor(cbr)
	cbr.Tojson(jsnback)
	if s := string(jsnback.Bytes()); s != "10" {
		t.Errorf("expected %q, got %q", "10", s)
	}

	// test SmartNumber
	config = NewDefaultConfig().NumberKind(SmartNumber).SpaceKind(UnicodeSpace)
	jsn = config.NewJson(make([]byte, 1024), 0)
	cbr = config.NewCbor(make([]byte, 1024), 0)
	jsnback = config.NewJson(make([]byte, 1024), 0)

	jsn.Reset([]byte("10.2"))
	jsn.Tocbor(cbr)
	cbr.Tojson(jsnback)
	if s := string(jsnback.Bytes()); s != "10.2" {
		t.Errorf("expected %q, got %q", "10.2", s)
	}

	// test SmartNumber (integer)
	config = NewDefaultConfig().NumberKind(SmartNumber32).SpaceKind(UnicodeSpace)
	jsn = config.NewJson(make([]byte, 1024), 0)
	cbr = config.NewCbor(make([]byte, 1024), 0)
	jsnback = config.NewJson(make([]byte, 1024), 0)

	jsn.Reset([]byte("10"))
	jsn.Tocbor(cbr)
	cbr.Tojson(jsnback)
	if s := string(jsnback.Bytes()); s != "10" {
		t.Errorf("expected %q, got %q", "10", s)
	}
}

func TestJsonNumber(t *testing.T) {

	// for number as integer.
	var ref1, ref2 interface{}
	testcases := []string{
		"255", "256", "-255", "-256", "65535", "65536", "-65535", "-65536",
		"4294967295", "4294967296", "-4294967295", "-4294967296",
		"9223372036854775807", "-9223372036854775807", "-9223372036854775808",
	}
	config := NewDefaultConfig().NumberKind(IntNumber).SpaceKind(UnicodeSpace)
	cbr := config.NewCbor(make([]byte, 1024), 0)
	jsn := config.NewJson(make([]byte, 1024), 0)
	jsnback := config.NewJson(make([]byte, 1024), 0)

	for _, tcase := range testcases {
		t.Logf("testcase - %v", tcase)
		json.Unmarshal([]byte(tcase), &ref1)

		jsn.Reset([]byte(tcase))
		jsn.Tocbor(cbr.Reset(nil))
		t.Logf("%v %v", len(cbr.Bytes()), cbr.Bytes())

		cbr.Tojson(jsnback.Reset(nil))
		if err := json.Unmarshal(jsnback.Bytes(), &ref2); err != nil {
			t.Errorf("json.Unmarshal() failed for cbor %v: %v", tcase, err)
		}
		if !reflect.DeepEqual(ref1, ref2) {
			t.Errorf("mismatch %v, got %v", ref1, ref2)
		}
	}

	// test float-number
	tcase := "10.2"
	json.Unmarshal([]byte(tcase), &ref1)

	config = NewDefaultConfig()
	jsn = config.NewJson(make([]byte, 1024), 0)
	cbr = config.NewCbor(make([]byte, 1024), 0)

	jsn.Reset([]byte(tcase))
	jsn.Tocbor(cbr)

	t.Logf("%v %v", len(cbr.Bytes()), cbr.Bytes())

	cbr.Tojson(jsnback.Reset(nil))
	if err := json.Unmarshal(jsnback.Bytes(), &ref2); err != nil {
		t.Errorf("json.Unmarshal() failed for cbor %v: %v", tcase, err)
	}
	if !reflect.DeepEqual(ref1, ref2) {
		t.Errorf("mismatch %v, got %v", ref1, ref2)
	}

	// negative small integers
	buf, out := make([]byte, 64), make([]byte, 64)
	n := valint82cbor(-1, buf)
	_, m := cbor2jsont1smallint(buf[:n], out, config)
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
		t.Logf("%v", tcase)
		panicfn(tcase, config)
	}
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

func TestJsonString(t *testing.T) {
	config := NewDefaultConfig()
	buf, out := make([]byte, 64), make([]byte, 64)

	ref := `"汉语 / 漢語; Hàn\b \t\uef24yǔ "`
	n := tag2cbor(uint64(tagJsonString), buf)
	x := valtext2cbor(ref, buf[n:])
	n += x

	_, m := cbor2json(buf[:n], out, config)
	if err := compare_jsons(t, ref, string(out[:m])); err != nil {
		t.Errorf("%v", err)
	}
}

func TestByteString(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic")
		}
	}()
	buf := make([]byte, 16)
	n := valbytes2cbor([]byte{0xf5}, buf)
	config := NewDefaultConfig()
	config.NewCbor(buf[:n], -1).Tojson(config.NewJson(make([]byte, 16), 0))
}

//---- benchmarks

func BenchmarkJson2CborNull(b *testing.B) {
	config := NewDefaultConfig()
	jsn := config.NewJson([]byte("null"), -1)
	cbr := config.NewCbor(make([]byte, 1024), 0)

	for i := 0; i < b.N; i++ {
		cbr.Reset(nil)
		jsn.Tocbor(cbr)
	}

	b.SetBytes(int64(len(jsn.Bytes())))
}

func BenchmarkCbor2JsonNull(b *testing.B) {
	config := NewDefaultConfig()
	jsn := config.NewJson([]byte("null"), -1)
	cbr := config.NewCbor(make([]byte, 1024), 0)

	jsn.Tocbor(cbr)

	for i := 0; i < b.N; i++ {
		jsn.Reset(nil)
		cbr.Tojson(jsn)
	}
	b.SetBytes(int64(len(cbr.Bytes())))
}

func BenchmarkJson2CborInt(b *testing.B) {
	config := NewDefaultConfig()
	jsn := config.NewJson([]byte("123456567"), -1)
	cbr := config.NewCbor(make([]byte, 1024), 0)

	for i := 0; i < b.N; i++ {
		cbr.Reset(nil)
		jsn.Tocbor(cbr)
	}
	b.SetBytes(int64(len(jsn.Bytes())))
}

func BenchmarkCbor2JsonInt(b *testing.B) {
	config := NewDefaultConfig()
	jsn := config.NewJson([]byte("123456567"), -1)
	cbr := config.NewCbor(make([]byte, 1024), 0)

	jsn.Tocbor(cbr)
	for i := 0; i < b.N; i++ {
		jsn.Reset(nil)
		cbr.Tojson(jsn)
	}
	b.SetBytes(int64(len(cbr.Bytes())))
}

func BenchmarkJson2CborFlt(b *testing.B) {
	config := NewDefaultConfig()
	jsn := config.NewJson([]byte("1234.12312"), -1)
	cbr := config.NewCbor(make([]byte, 1024), 0)

	for i := 0; i < b.N; i++ {
		cbr.Reset(nil)
		jsn.Tocbor(cbr)
	}
	b.SetBytes(int64(len(jsn.Bytes())))
}

func BenchmarkCbor2JsonFlt(b *testing.B) {
	config := NewDefaultConfig()
	jsn := config.NewJson([]byte("1234.12312"), -1)
	cbr := config.NewCbor(make([]byte, 1024), 0)

	jsn.Tocbor(cbr)

	for i := 0; i < b.N; i++ {
		jsn.Reset(nil)
		cbr.Tojson(jsn)
	}
	b.SetBytes(int64(len(cbr.Bytes())))
}

func BenchmarkJson2CborBool(b *testing.B) {
	config := NewDefaultConfig()
	jsn := config.NewJson([]byte("false"), -1)
	cbr := config.NewCbor(make([]byte, 1024), 0)

	for i := 0; i < b.N; i++ {
		cbr.Reset(nil)
		jsn.Tocbor(cbr)
	}

	b.SetBytes(int64(len(jsn.Bytes())))
}

func BenchmarkCbor2JsonBool(b *testing.B) {
	config := NewDefaultConfig()
	jsn := config.NewJson([]byte("false"), -1)
	cbr := config.NewCbor(make([]byte, 1024), 0)

	jsn.Tocbor(cbr)

	for i := 0; i < b.N; i++ {
		jsn.Reset(nil)
		cbr.Tojson(jsn)
	}
	b.SetBytes(int64(len(cbr.Bytes())))
}

func BenchmarkJson2CborStr(b *testing.B) {
	config := NewDefaultConfig()
	jsn := config.NewJson([]byte(`"汉语 / 漢語; Hàn\b \t\uef24yǔ "`), -1)
	cbr := config.NewCbor(make([]byte, 1024), 0)

	for i := 0; i < b.N; i++ {
		cbr.Reset(nil)
		jsn.Tocbor(cbr)
	}
	b.SetBytes(int64(len(jsn.Bytes())))
}

func BenchmarkCbor2JsonStr(b *testing.B) {
	config := NewDefaultConfig()
	jsn := config.NewJson([]byte(`"汉语 / 漢語; Hàn\b \t\uef24yǔ "`), -1)
	cbr := config.NewCbor(make([]byte, 1024), 0)

	jsn.Tocbor(cbr)

	for i := 0; i < b.N; i++ {
		jsn.Reset(nil)
		cbr.Tojson(jsn)
	}
	b.SetBytes(int64(len(cbr.Bytes())))
}

func BenchmarkJson2CborArr(b *testing.B) {
	in := ` [null,true,false,10,"tru\"e"]`
	config := NewDefaultConfig()
	jsn := config.NewJson([]byte(in), -1)
	cbr := config.NewCbor(make([]byte, 1024), 0)

	for i := 0; i < b.N; i++ {
		cbr.Reset(nil)
		jsn.Tocbor(cbr)
	}
	b.SetBytes(int64(len(jsn.Bytes())))
}

func BenchmarkCbor2JsonArr(b *testing.B) {
	in := ` [null,true,false,10,"tru\"e"]`
	config := NewDefaultConfig()
	jsn := config.NewJson([]byte(in), -1)
	cbr := config.NewCbor(make([]byte, 1024), 0)

	jsn.Tocbor(cbr)

	for i := 0; i < b.N; i++ {
		jsn.Reset(nil)
		cbr.Tojson(jsn)
	}
	b.SetBytes(int64(len(cbr.Bytes())))
}

func BenchmarkJson2CborMap(b *testing.B) {
	in := `{"a":null,"b":true,"c":false,"d\"":10,"e":"tru\"e", "f":[1,2]}`
	config := NewDefaultConfig()
	jsn := config.NewJson([]byte(in), -1)
	cbr := config.NewCbor(make([]byte, 1024), 0)

	for i := 0; i < b.N; i++ {
		cbr.Reset(nil)
		jsn.Tocbor(cbr)
	}
	b.SetBytes(int64(len(jsn.Bytes())))
}

func BenchmarkCbor2JsonMap(b *testing.B) {
	in := `{"a":null,"b":true,"c":false,"d\"":10,"e":"tru\"e", "f":[1,2]}`
	config := NewDefaultConfig()
	jsn := config.NewJson([]byte(in), -1)
	cbr := config.NewCbor(make([]byte, 1024), 0)

	jsn.Tocbor(cbr)

	for i := 0; i < b.N; i++ {
		jsn.Reset(nil)
		cbr.Tojson(jsn)
	}
	b.SetBytes(int64(len(cbr.Bytes())))
}

func BenchmarkJson2CborTyp(b *testing.B) {
	in := testdataFile("testdata/typical.json")
	config := NewDefaultConfig()
	jsn := config.NewJson(in, -1)
	cbr := config.NewCbor(make([]byte, 10*1024), 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cbr.Reset(nil)
		jsn.Tocbor(cbr)
	}
	b.SetBytes(int64(len(jsn.Bytes())))
}

func BenchmarkCbor2JsonTyp(b *testing.B) {
	in := testdataFile("testdata/typical.json")
	config := NewDefaultConfig()
	jsn := config.NewJson(in, -1)
	cbr := config.NewCbor(make([]byte, 10*1024), 0)

	jsn.Tocbor(cbr)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		jsn.Reset(nil)
		cbr.Tojson(jsn)
	}
	b.SetBytes(int64(len(cbr.Bytes())))
}
