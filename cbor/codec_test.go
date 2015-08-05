package cbor

import "testing"
import "reflect"
import "encoding/json"
import "io/ioutil"
import "strings"
import "fmt"
import "os"
import "compress/gzip"

var _ = fmt.Sprintf("dummy")

func TestCborMajor(t *testing.T) {
	if typ := major(0xff); typ != 0xe0 {
		t.Errorf("fail major() got %v wanted 0xe0", typ)
	}
}

func TestCborNil(t *testing.T) {
	buf := make([]byte, 10)

	if n := encode(nil, buf); n != 1 {
		t.Errorf("fail encode nil: %v want 1", n)
	} else if item, m := decode(buf); m != 1 {
		t.Errorf("fail decode on nil len: %v want 1", m)
	} else if item != nil {
		t.Errorf("fail decode on nil: %x", item)
	}
}

func TestCborTrue(t *testing.T) {
	buf := make([]byte, 10)

	if n := encode(true, buf); n != 1 {
		t.Errorf("fail encode true: %v want 1", n)
	} else if item, m := decode(buf); m != 1 {
		t.Errorf("fail decode on true len: %v want 1", m)
	} else if item.(bool) != true {
		t.Errorf("fail decode on true: %v", item)
	}
}

func TestCborFalse(t *testing.T) {
	buf := make([]byte, 10)

	if n := encode(false, buf); n != 1 {
		t.Errorf("fail encode false: %v want 1", n)
	} else if item, m := decode(buf); m != 1 {
		t.Errorf("fail decode on false len: %v want 1", m)
	} else if item.(bool) != false {
		t.Errorf("fail decode on false: %v", item)
	}
}

func TestCborUint8(t *testing.T) {
	buf := make([]byte, 20)
	for i := uint16(0); i <= 255; i++ {
		n := encodeUint8(uint8(i), buf)
		val, m := decode(buf)
		if i < 24 && n != 1 {
			t.Errorf("fail code uint8(%v) < 24, got %v", i, n)
		} else if i > 24 && n != 2 {
			t.Errorf("fail code uint8(%v) > 24, got %v", i, n)
		}
		if n != m || val.(uint64) != uint64(i) {
			t.Errorf("fail codec uint8: %v %v %v", n, m, val)
		}
	}
}

func TestCborInt8(t *testing.T) {
	buf := make([]byte, 20)
	for i := int16(-128); i <= 127; i++ {
		n := encodeInt8(int8(i), buf)
		val, m := decode(buf)
		if -23 <= i && i <= 23 && n != 1 {
			t.Errorf("fail code int8(%v), expected 1 got %v", i, n)
		} else if -23 > i && i > 23 && n != 2 {
			t.Errorf("fail code int8(%v), expected 2 got %v", i, n)
		}
		if num1, ok := val.(int64); ok && (n != m || num1 != int64(i)) {
			t.Errorf("fail codec uint8: %v %v %v", n, m, num1)
		} else if num2, ok := val.(uint64); ok && (n != m || num2 != uint64(i)) {
			t.Errorf("fail codec uint8: %v %v %v", n, m, num2)
		}
	}
}

func TestCborNum(t *testing.T) {
	buf := make([]byte, 20)
	tests := [][2]interface{}{
		[2]interface{}{'a', uint64(97)},
		[2]interface{}{byte(0), uint64(0)},
		[2]interface{}{byte(23), uint64(23)},
		[2]interface{}{byte(24), uint64(24)},
		[2]interface{}{byte(255), uint64(255)},
		[2]interface{}{uint8(0), uint64(0)},
		[2]interface{}{uint8(23), uint64(23)},
		[2]interface{}{uint8(24), uint64(24)},
		[2]interface{}{uint8(255), uint64(255)},
		[2]interface{}{int8(-128), int64(-128)},
		[2]interface{}{int8(-24), int64(-24)},
		[2]interface{}{int8(-24), int64(-24)},
		[2]interface{}{int8(-1), int64(-1)},
		[2]interface{}{int8(-0), uint64(-0)},
		[2]interface{}{int8(0), uint64(0)},
		[2]interface{}{int8(23), uint64(23)},
		[2]interface{}{int8(24), uint64(24)},
		[2]interface{}{int8(127), uint64(127)},
		[2]interface{}{uint16(0), uint64(0)},
		[2]interface{}{uint16(23), uint64(23)},
		[2]interface{}{uint16(24), uint64(24)},
		[2]interface{}{uint16(255), uint64(255)},
		[2]interface{}{uint16(65535), uint64(65535)},
		[2]interface{}{int16(-32768), int64(-32768)},
		[2]interface{}{int16(-256), int64(-256)},
		[2]interface{}{int16(-255), int64(-255)},
		[2]interface{}{int16(-129), int64(-129)},
		[2]interface{}{int16(-128), int64(-128)},
		[2]interface{}{int16(-127), int64(-127)},
		[2]interface{}{int16(-24), int64(-24)},
		[2]interface{}{int16(-23), int64(-23)},
		[2]interface{}{int16(-1), int64(-1)},
		[2]interface{}{int16(-0), uint64(0)},
		[2]interface{}{int16(0), uint64(0)},
		[2]interface{}{int16(23), uint64(23)},
		[2]interface{}{int16(24), uint64(24)},
		[2]interface{}{int16(127), uint64(127)},
		[2]interface{}{int16(255), uint64(255)},
		[2]interface{}{int16(32767), uint64(32767)},
		[2]interface{}{uint32(0), uint64(0)},
		[2]interface{}{uint32(23), uint64(23)},
		[2]interface{}{uint32(24), uint64(24)},
		[2]interface{}{uint32(255), uint64(255)},
		[2]interface{}{uint32(65535), uint64(65535)},
		[2]interface{}{uint32(4294967295), uint64(4294967295)},
		[2]interface{}{int32(-2147483648), int64(-2147483648)},
		[2]interface{}{int32(-32769), int64(-32769)},
		[2]interface{}{int32(-32768), int64(-32768)},
		[2]interface{}{int32(-32767), int64(-32767)},
		[2]interface{}{int32(-256), int64(-256)},
		[2]interface{}{int32(-255), int64(-255)},
		[2]interface{}{int32(-129), int64(-129)},
		[2]interface{}{int32(-128), int64(-128)},
		[2]interface{}{int32(-127), int64(-127)},
		[2]interface{}{int32(-24), int64(-24)},
		[2]interface{}{int32(-23), int64(-23)},
		[2]interface{}{int32(-1), int64(-1)},
		[2]interface{}{int32(-0), uint64(-0)},
		[2]interface{}{int32(0), uint64(0)},
		[2]interface{}{int32(23), uint64(23)},
		[2]interface{}{int32(24), uint64(24)},
		[2]interface{}{int32(127), uint64(127)},
		[2]interface{}{int32(32767), uint64(32767)},
		[2]interface{}{int32(65535), uint64(65535)},
		[2]interface{}{int32(2147483647), uint64(2147483647)},
		[2]interface{}{int(-2147483648), int64(-2147483648)},
		[2]interface{}{uint(2147483647), uint64(2147483647)},
		[2]interface{}{uint64(0), uint64(0)},
		[2]interface{}{uint64(23), uint64(23)},
		[2]interface{}{uint64(24), uint64(24)},
		[2]interface{}{uint64(255), uint64(255)},
		[2]interface{}{uint64(65535), uint64(65535)},
		[2]interface{}{uint64(4294967295), uint64(4294967295)},
		[2]interface{}{uint64(18446744073709551615), uint64(18446744073709551615)},
		[2]interface{}{int64(-9223372036854775808), int64(-9223372036854775808)},
		[2]interface{}{int64(-2147483649), int64(-2147483649)},
		[2]interface{}{int64(-2147483648), int64(-2147483648)},
		[2]interface{}{int64(-2147483647), int64(-2147483647)},
		[2]interface{}{int64(-32769), int64(-32769)},
		[2]interface{}{int64(-32768), int64(-32768)},
		[2]interface{}{int64(-32767), int64(-32767)},
		[2]interface{}{int64(-256), int64(-256)},
		[2]interface{}{int64(-255), int64(-255)},
		[2]interface{}{int64(-129), int64(-129)},
		[2]interface{}{int64(-128), int64(-128)},
		[2]interface{}{int64(-127), int64(-127)},
		[2]interface{}{int64(-24), int64(-24)},
		[2]interface{}{int64(-23), int64(-23)},
		[2]interface{}{int64(-1), int64(-1)},
		[2]interface{}{int64(-0), uint64(-0)},
		[2]interface{}{int64(0), uint64(0)},
		[2]interface{}{int64(23), uint64(23)},
		[2]interface{}{int64(24), uint64(24)},
		[2]interface{}{int64(127), uint64(127)},
		[2]interface{}{int64(32767), uint64(32767)},
		[2]interface{}{int64(2147483647), uint64(2147483647)},
		[2]interface{}{int64(4294967295), uint64(4294967295)},
		[2]interface{}{int64(9223372036854775807), uint64(9223372036854775807)},
	}
	for _, test := range tests {
		n := encode(test[0], buf)
		val, m := decode(buf)
		//t.Logf("executing test case %v", test)
		if n != m || !reflect.DeepEqual(val, test[1]) {
			t.Errorf(
				"fail codec Num: %v %v %T(%v) %T(%v)",
				n, m, val, val, test[1], test[1])
		}
	}
	// test case for number exceeding int64
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic decoding int64 > 9223372036854775807")
			}
		}()
		encode(uint64(9223372036854775808), buf)
		buf[0] = (buf[0] & 0x1f) | type1 // fix as negative integer
		decode(buf)
	}()
}

func TestCborFloat16(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic while decoding float16")
		}
	}()
	decode([]byte{0xf9, 0, 0, 0, 0})
}

func TestCborFloat32(t *testing.T) {
	buf, ref := make([]byte, 10), float32(10.11)
	n := encode(ref, buf)
	t.Logf("%v", buf)
	val, m := decode(buf)
	if n != 5 || n != m || !reflect.DeepEqual(val, ref) {
		t.Errorf("fail code float32: %v %v %T(%v)", n, m, val, val)
	}
}

func TestCborFloat64(t *testing.T) {
	buf, ref := make([]byte, 10), float64(10.11)
	n := encode(ref, buf)
	val, m := decode(buf)
	if n != 9 || n != m || !reflect.DeepEqual(val, ref) {
		t.Errorf("fail code float32: %v %v %T(%v)", n, m, val, val)
	}
}

func TestCborBytes(t *testing.T) {
	buf, ref := make([]byte, 200), make([]uint8, 100)
	for i := 0; i < len(ref); i++ {
		ref[i] = uint8(i)
	}
	n := encode(ref, buf)
	val, m := decode(buf)
	if n != 102 || n != m || !reflect.DeepEqual(val, ref) {
		t.Errorf("fail code bytes: %v %v %T(%v)", n, m, val, val)
	}
	// test byte-start
	if n := encodeBytesStart(buf); n != 1 {
		t.Errorf("fail code bytes-start len: %v wanted 1", n)
	} else if val, m := decode(buf); m != n {
		t.Errorf("fail code bytes-start size: %v wanted %v", m, n)
	} else if !reflect.DeepEqual(val, Indefinite(0x5f)) {
		t.Errorf("fail code bytes-start: %v wanted 0x5f", buf[0])
	}
}

func TestCborText(t *testing.T) {
	buf, ref := make([]byte, 200), "hello world"
	n := encode(ref, buf)
	val, m := decode(buf)
	if n != 12 || n != m || !reflect.DeepEqual(val, ref) {
		t.Errorf("fail code text: %v %v %T(%v)", n, m, val, val)
	}

	// test text-start
	if n := encodeTextStart(buf); n != 1 {
		t.Errorf("fail code text-start len: %v wanted 1", n)
	} else if val, m := decode(buf); m != n {
		t.Errorf("fail code text-start size: %v wanted %v", m, n)
	} else if !reflect.DeepEqual(val, Indefinite(0x7f)) {
		t.Errorf("fail code text-start: %x wanted 0x7f", buf[0])
	}
}

func TestCborArray(t *testing.T) {
	buf := make([]byte, 1024)
	ref := []interface{}{10.2, "hello world"}
	n := encode(ref, buf)
	val, m := decode(buf)
	if n != 22 || n != m || !reflect.DeepEqual(ref, val) {
		t.Errorf("fail code text: %v %v %T(%v)", n, m, val, val)
	}
	// test text-start
	n = encodeArrayStart(buf)
	n += encode(uint64(1), buf[n:])
	n += encodeBreakStop(buf[n:])
	if n != 3 {
		t.Errorf("fail code text-start len: %v wanted 3", n)
	} else if val, m := decode(buf[:n]); m != n {
		t.Errorf("fail code text-start size : %v wanted %v", m, n)
	} else if ref := []interface{}{uint64(1)}; !reflect.DeepEqual(val, ref) {
		t.Errorf("fail code text-start: {%T,%v} wanted {%T,%v}", val, val, ref, ref)
	}
}

func TestCborMap(t *testing.T) {
	buf := make([]byte, 1024)
	ref := [][2]interface{}{
		[2]interface{}{10.2, "hello world"},
		[2]interface{}{"hello world", 10.2},
	}
	n := encode(ref, buf)
	val, m := decode(buf)
	if n != 43 || n != m || !reflect.DeepEqual(ref, val) {
		t.Errorf("fail code text: %v %v %T(%v)", n, m, val, val)
	}
	// test text-start
	n = encodeMapStart(buf)
	n += encode("a", buf[n:])
	n += encode(1, buf[n:])
	n += encodeBreakStop(buf[n:])
	ref = [][2]interface{}{[2]interface{}{"a", uint64(1)}}
	if n != 5 {
		t.Errorf("fail code text-start len: %v wanted 5", n)
	} else if val, m := decode(buf[:n]); m != n {
		t.Errorf("fail code text-start size : %v wanted %v", m, n)
	} else if !reflect.DeepEqual(val, ref) {
		t.Errorf("fail code text-start: %v wanted %v", val, ref)
	}
}

func TestCborBreakStop(t *testing.T) {
	buf := make([]byte, 10)
	if n := encodeBreakStop(buf); n != 1 {
		t.Errorf("fail code text-start len: %v wanted 1", n)
	} else if val, m := decode(buf); m != n {
		t.Errorf("fail code text-start: %v wanted %v", m, n)
	} else if !reflect.DeepEqual(val, BreakStop(0xff)) {
		t.Errorf("fail code text-start: %x wanted 0xff", buf[0])
	}
}

func TestCborUndefined(t *testing.T) {
	buf := make([]byte, 10)

	if n := encodeUndefined(buf); n != 1 {
		t.Errorf("fail encode undefined: %v want 1", n)
	} else if item, m := decode(buf); m != 1 {
		t.Errorf("fail decode on undefined len: %v want 1", m)
	} else if item.(Undefined) != Undefined(simpleUndefined) {
		t.Errorf("fail decode on undefined: %T %v", item, item)
	}
}

func TestCborReserved(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic while decoding reserved")
		}
	}()
	decode([]byte{hdr(type0, 28)})
}

func TestCborMaster(t *testing.T) {
	var outval, ref interface{}

	testcases := append(scan_valid, []string{
		string(mapValue),
		string(allValueIndent),
		string(allValueCompact),
		string(pallValueIndent),
		string(pallValueCompact),
	}...)

	config := NewDefaultConfig()
	cborout, jsonout := make([]byte, 1024*1024), make([]byte, 1024*1024)
	for _, tcase := range testcases {
		t.Logf("%v", tcase)
		if err := json.Unmarshal([]byte(tcase), &ref); err != nil {
			t.Fatalf("error parsing %q: %v", tcase, err)
		}
		// test ParseJson/ToJson
		_, n := config.ParseJson(tcase, cborout)    // json -> cbor
		_, q := config.ToJson(cborout[:n], jsonout) // cbor -> json
		if err := json.Unmarshal(jsonout[:q], &outval); err != nil {
			t.Fatalf("error parsing %q: %v", jsonout[:q], err)
		} else if !reflect.DeepEqual(outval, ref) {
			t.Fatalf("expected '%v', got '%v'", ref, outval)
		}

		value, _ := config.Decode(cborout[:n])     // cbor -> golang
		p := config.Encode(value, cborout)         // golang -> cbor
		_, q = config.ToJson(cborout[:p], jsonout) // cbor -> json
		if err := json.Unmarshal(jsonout[:q], &outval); err != nil {
			t.Fatalf("error parsing %q: %v", jsonout[:q], err)
		} else if !reflect.DeepEqual(outval, ref) {
			t.Fatalf("expected {%T,%v}, got {%T,%v}", value, value, outval, outval)
		}
	}
}

func TestCborSmartnum(t *testing.T) {
	var outval, ref interface{}

	data := testdataFile("../testdata/smartnum")
	config := NewDefaultConfig()
	cborout, jsonout := make([]byte, 1024*1024), make([]byte, 1024*1024)

	if err := json.Unmarshal(data, &ref); err != nil {
		t.Fatalf("error parsing code.json.gz: %v", err)
	}

	// test ParseJson/ToJson
	_, n := config.ParseJson(string(data), cborout) // json -> cbor
	_, q := config.ToJson(cborout[:n], jsonout)
	if err := json.Unmarshal(jsonout[:q], &outval); err != nil {
		t.Logf("%v", string(jsonout[:q]))
		t.Fatalf("error parsing code.json.gz: %v", err)
	} else if !reflect.DeepEqual(ref, outval) {
		t.Errorf("expected %v", ref)
		t.Errorf("got-json %v", string(jsonout[:q]))
		t.Fatalf("got %v", outval)
	}

	value, _ := config.Decode(cborout[:n])     // cbor -> golang
	p := config.Encode(value, cborout)         // golang -> cbor
	_, q = config.ToJson(cborout[:p], jsonout) // cbor -> json
	if err := json.Unmarshal(jsonout[:q], &outval); err != nil {
		t.Fatalf("error parsing %v", err)
	} else if err := json.Unmarshal(data, &value); err != nil {
		t.Fatalf("error parsing code.json: %v", err)
	} else if !reflect.DeepEqual(outval, value) {
		t.Fatalf("expected %v", value)
		t.Fatalf("got %v", outval)
	}
}

func TestCborMalformed(t *testing.T) {
	config := NewConfig(IntNumber, AnsiSpace)
	out := make([]byte, 1024)
	for _, tcase := range scan_invalid {
		func() {
			defer func() {
				if tcase == `"g-clef: \uD834\uDD1E"` {
				} else if r := recover(); r == nil {
					t.Fatalf("expected panic")
				}
			}()
			t.Logf("%v", tcase)
			config.ParseJson(tcase, out)
		}()
	}
}

func TestCborCodeJSON(t *testing.T) {
	var ref, outval interface{}

	config := NewDefaultConfig()
	cborout, jsonout := make([]byte, 10*1024*1024), make([]byte, 10*1024*1024)
	data := testdataFile("../testdata/code.json.gz")

	if err := json.Unmarshal(data, &ref); err != nil {
		t.Fatalf("error parsing code.json.gz: %v", err)
	}
	ref = fixFloats(ref)

	// test ParseJson/ToJson
	_, n := config.ParseJson(string(data), cborout) // json -> cbor
	_, q := config.ToJson(cborout[:n], jsonout)
	t.Logf("%v %v %v %v", n, q, len(data), len(jsonout[:q]))
	if err := json.Unmarshal(jsonout[:q], &outval); err != nil {
		t.Logf("%v", string(jsonout[:q]))
		t.Fatalf("error parsing code.json.gz: %v", err)
	} else {
		outval = fixFloats(outval)
		if !reflect.DeepEqual(ref, outval) {
			t.Errorf("expected %v", ref)
			t.Fatalf("got %v", outval)
		}
	}

	value, _ := config.Decode(cborout[:n])     // cbor -> golang
	p := config.Encode(value, cborout)         // golang -> cbor
	_, q = config.ToJson(cborout[:p], jsonout) // cbor -> json
	if err := json.Unmarshal(jsonout[:q], &outval); err != nil {
		t.Fatalf("error parsing %v", err)
	} else {
		outval = fixFloats(outval)
		if !reflect.DeepEqual(outval, ref) {
			t.Fatalf("expected %v", value)
			t.Fatalf("got %v", outval)
		}
	}
}

func BenchmarkEncodeNull(b *testing.B) {
	buf := make([]byte, 10)
	for i := 0; i < b.N; i++ {
		encode(nil, buf)
	}
}

func BenchmarkDecodeNull(b *testing.B) {
	buf := make([]byte, 10)
	n := encode(nil, buf)
	for i := 0; i < b.N; i++ {
		decode(buf[:n])
	}
}

func BenchmarkEncodeTrue(b *testing.B) {
	buf := make([]byte, 64)
	for i := 0; i < b.N; i++ {
		encode(true, buf)
	}
}

func BenchmarkDecodeTrue(b *testing.B) {
	buf := make([]byte, 10)
	n := encode(true, buf)
	for i := 0; i < b.N; i++ {
		decode(buf[:n])
	}
}

func BenchmarkEncodeFalse(b *testing.B) {
	buf := make([]byte, 10)
	for i := 0; i < b.N; i++ {
		encode(false, buf)
	}
}

func BenchmarkDecodeFalse(b *testing.B) {
	buf := make([]byte, 10)
	n := encode(false, buf)
	for i := 0; i < b.N; i++ {
		decode(buf[:n])
	}
}

func BenchmarkEncodeUint8(b *testing.B) {
	buf := make([]byte, 64)
	for i := 0; i < b.N; i++ {
		encode(uint8(255), buf)
	}
}

func BenchmarkDecodeUint8(b *testing.B) {
	buf := make([]byte, 64)
	n := encode(uint8(255), buf)
	for i := 0; i < b.N; i++ {
		decode(buf[:n])
	}
}

func BenchmarkEncodeInt8(b *testing.B) {
	buf := make([]byte, 64)
	for i := 0; i < b.N; i++ {
		encode(int8(-128), buf)
	}
}

func BenchmarkDecodeInt8(b *testing.B) {
	buf := make([]byte, 64)
	n := encode(int8(-128), buf)
	for i := 0; i < b.N; i++ {
		decode(buf[:n])
	}
}

func BenchmarkEncodeUint16(b *testing.B) {
	buf := make([]byte, 64)
	for i := 0; i < b.N; i++ {
		encode(uint16(65535), buf)
	}
}

func BenchmarkDecodeUint16(b *testing.B) {
	buf := make([]byte, 64)
	n := encode(uint16(65535), buf)
	for i := 0; i < b.N; i++ {
		decode(buf[:n])
	}
}

func BenchmarkEncodeInt16(b *testing.B) {
	buf := make([]byte, 64)
	for i := 0; i < b.N; i++ {
		encode(int16(-32768), buf)
	}
}

func BenchmarkDecodeInt16(b *testing.B) {
	buf := make([]byte, 64)
	n := encode(int16(-32768), buf)
	for i := 0; i < b.N; i++ {
		decode(buf[:n])
	}
}

func BenchmarkEncodeUint32(b *testing.B) {
	buf := make([]byte, 64)
	for i := 0; i < b.N; i++ {
		encode(uint32(4294967295), buf)
	}
}

func BenchmarkDecodeUint32(b *testing.B) {
	buf := make([]byte, 64)
	n := encode(uint32(4294967295), buf)
	for i := 0; i < b.N; i++ {
		decode(buf[:n])
	}
}

func BenchmarkEncodeInt32(b *testing.B) {
	buf := make([]byte, 64)
	for i := 0; i < b.N; i++ {
		encode(int32(-2147483648), buf)
	}
}

func BenchmarkDecodeInt32(b *testing.B) {
	buf := make([]byte, 64)
	n := encode(int32(-2147483648), buf)
	for i := 0; i < b.N; i++ {
		decode(buf[:n])
	}
}

func BenchmarkEncodeUint64(b *testing.B) {
	buf := make([]byte, 64)
	for i := 0; i < b.N; i++ {
		encode(uint64(18446744073709551615), buf)
	}
}

func BenchmarkDecodeUint64(b *testing.B) {
	buf := make([]byte, 64)
	n := encode(uint64(18446744073709551615), buf)
	for i := 0; i < b.N; i++ {
		decode(buf[:n])
	}
}

func BenchmarkEncodeInt64(b *testing.B) {
	buf := make([]byte, 64)
	for i := 0; i < b.N; i++ {
		encode(int64(-2147483648), buf)
	}
}

func BenchmarkDecodeInt64(b *testing.B) {
	buf := make([]byte, 64)
	n := encode(int64(-2147483648), buf)
	for i := 0; i < b.N; i++ {
		decode(buf[:n])
	}
}

func BenchmarkEncodeFlt32(b *testing.B) {
	buf := make([]byte, 64)
	for i := 0; i < b.N; i++ {
		encode(float32(10.2), buf)
	}
}

func BenchmarkDecodeFlt32(b *testing.B) {
	buf := make([]byte, 64)
	n := encode(float32(10.2), buf)
	for i := 0; i < b.N; i++ {
		decode(buf[:n])
	}
}

func BenchmarkEncodeFlt64(b *testing.B) {
	buf := make([]byte, 64)
	for i := 0; i < b.N; i++ {
		encode(float64(10.2), buf)
	}
}

func BenchmarkDecodeFlt64(b *testing.B) {
	buf := make([]byte, 64)
	n := encode(float64(10.2), buf)
	for i := 0; i < b.N; i++ {
		decode(buf[:n])
	}
}

func BenchmarkEncodeBytes(b *testing.B) {
	buf := make([]byte, 64)
	bs := []byte("hello world")
	for i := 0; i < b.N; i++ {
		encode(bs, buf)
	}
}

func BenchmarkDecodeBytes(b *testing.B) {
	buf := make([]byte, 64)
	n := encode([]byte("hello world"), buf)
	for i := 0; i < b.N; i++ {
		decode(buf[:n])
	}
}

func BenchmarkEncodeText(b *testing.B) {
	buf := make([]byte, 64)
	text := "hello world"
	for i := 0; i < b.N; i++ {
		encode(text, buf)
	}
}

func BenchmarkDecodeText(b *testing.B) {
	buf := make([]byte, 64)
	n := encode("hello world", buf)
	for i := 0; i < b.N; i++ {
		decode(buf[:n])
	}
}

func BenchmarkEncodeArr0(b *testing.B) {
	buf, arr := make([]byte, 1024), make([]interface{}, 0)
	for i := 0; i < b.N; i++ {
		encode(arr, buf)
	}
}

func BenchmarkDecodeArr0(b *testing.B) {
	buf, arr := make([]byte, 1024), make([]interface{}, 0)
	n := encode(arr, buf)
	for i := 0; i < b.N; i++ {
		decode(buf[:n])
	}
}

func BenchmarkEncodeArr5(b *testing.B) {
	buf := make([]byte, 1024)
	arr := []interface{}{5, 5.0, "hello world", true, nil}
	for i := 0; i < b.N; i++ {
		encode(arr, buf)
	}
}

func BenchmarkDecodeArr5(b *testing.B) {
	buf := make([]byte, 1024)
	arr := []interface{}{5, 5.0, "hello world", true, nil}
	n := encode(arr, buf)
	for i := 0; i < b.N; i++ {
		decode(buf[:n])
	}
}

func BenchmarkEncodeMap0(b *testing.B) {
	buf := make([]byte, 1024)
	m := make([][2]interface{}, 0)
	for i := 0; i < b.N; i++ {
		encode(m, buf)
	}
}

func BenchmarkDecodeMap0(b *testing.B) {
	buf := make([]byte, 1024)
	m := make([][2]interface{}, 0)
	n := encode(m, buf)
	for i := 0; i < b.N; i++ {
		decode(buf[:n])
	}
}

func BenchmarkEncodeMap5(b *testing.B) {
	buf := make([]byte, 1024)
	m := [][2]interface{}{
		[2]interface{}{"key0", 5}, [2]interface{}{"key1", 5.0},
		[2]interface{}{"key2", "hello world"},
		[2]interface{}{"key3", true}, [2]interface{}{"key4", nil},
	}
	for i := 0; i < b.N; i++ {
		encode(m, buf)
	}
}

func BenchmarkDecodeMap5(b *testing.B) {
	buf := make([]byte, 1024)
	m := [][2]interface{}{
		[2]interface{}{"key0", 5}, [2]interface{}{"key1", 5.0},
		[2]interface{}{"key2", "hello world"},
		[2]interface{}{"key3", true}, [2]interface{}{"key4", nil},
	}
	n := encode(m, buf)
	for i := 0; i < b.N; i++ {
		decode(buf[:n])
	}
}

var allValueIndent, allValueCompact, pallValueIndent, pallValueCompact []byte
var mapValue []byte
var scan_valid []string
var scan_invalid []string

func init() {
	var value interface{}
	var err error

	allValueIndent, err = ioutil.ReadFile("../testdata/allValueIndent")
	if err != nil {
		panic(err)
	}

	if err = json.Unmarshal(allValueIndent, &value); err != nil {
		panic(err)
	}
	if allValueCompact, err = json.Marshal(value); err != nil {
		panic(err)
	}

	pallValueIndent, err = ioutil.ReadFile("../testdata/pallValueIndent")
	if err != nil {
		panic(err)
	}

	if err = json.Unmarshal(pallValueIndent, &value); err != nil {
		panic(err)
	}
	if pallValueCompact, err = json.Marshal(value); err != nil {
		panic(err)
	}

	mapValue, err = ioutil.ReadFile("../testdata/map")
	if err != nil {
		panic(err)
	}

	scan_valid_b, err := ioutil.ReadFile("../testdata/scan_valid")
	if err != nil {
		panic(err)
	}
	scan_valid = []string{}
	for _, s := range strings.Split(string(scan_valid_b), "\n") {
		if strings.Trim(s, " ") != "" {
			scan_valid = append(scan_valid, s)
		}
	}
	scan_valid = append(scan_valid, []string{
		"\"hello\xffworld\"",
		"\"hello\xc2\xc2world\"",
		"\"hello\xc2\xffworld\"",
		"\"hello\xed\xa0\x80\xed\xb0\x80world\""}...)

	scan_invalid_b, err := ioutil.ReadFile("../testdata/scan_invalid")
	if err != nil {
		panic(err)
	}
	scan_invalid = []string{}
	for _, s := range strings.Split(string(scan_invalid_b), "\n") {
		if strings.Trim(s, " ") != "" {
			scan_invalid = append(scan_invalid, s)
		}
	}
	scan_invalid = append(scan_invalid, []string{
		"\xed\xa0\x80", // RuneError
		"\xed\xbf\xbf", // RuneError
		// raw value errors
		"\x01 42",
		"\x01 true",
		"\x01 1.2",
		" 3.4 \x01",
		"\x01 \"string\"",
		// bad-utf8
		"hello\xffworld",
		"\xff",
		"\xff\xff",
		"a\xffb",
		"\xe6\x97\xa5\xe6\x9c\xac\xff\xaa\x9e"}...)
}

func testdataFile(filename string) []byte {
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	var data []byte
	if strings.HasSuffix(filename, ".gz") {
		gz, err := gzip.NewReader(f)
		if err != nil {
			panic(err)
		}
		data, err = ioutil.ReadAll(gz)
		if err != nil {
			panic(err)
		}
	} else {
		data, err = ioutil.ReadAll(f)
		if err != nil {
			panic(err)
		}
	}
	return data
}

func fixFloats(val interface{}) interface{} {
	switch v := val.(type) {
	case float64:
		return float32(v)
	case []interface{}:
		for i, x := range v {
			v[i] = fixFloats(x)
		}
		return v
	case map[string]interface{}:
		for p, q := range v {
			v[p] = fixFloats(q)
		}
		return v
	}
	return val
}
