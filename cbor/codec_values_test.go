package cbor

import "testing"
import "reflect"
import "encoding/json"
import "io/ioutil"
import "strings"
import "fmt"
import "os"
import "compress/gzip"
import "time"
import "regexp"
import "math/big"

var _ = fmt.Sprintf("dummy")

func TestCborMajor(t *testing.T) {
	if typ := major(0xff); typ != 0xe0 {
		t.Errorf("fail major() got %v wanted 0xe0", typ)
	}
}

func TestCborNil(t *testing.T) {
	buf := make([]byte, 10)
	config := NewDefaultConfig()

	if n := encode(nil, buf, config); n != 1 {
		t.Errorf("fail encode nil: %v want 1", n)
	} else if item, m := decode(buf); m != 1 {
		t.Errorf("fail decode on nil len: %v want 1", m)
	} else if item != nil {
		t.Errorf("fail decode on nil: %x", item)
	}
}

func TestCborTrue(t *testing.T) {
	buf := make([]byte, 10)
	config := NewDefaultConfig()

	if n := encode(true, buf, config); n != 1 {
		t.Errorf("fail encode true: %v want 1", n)
	} else if item, m := decode(buf); m != 1 {
		t.Errorf("fail decode on true len: %v want 1", m)
	} else if item.(bool) != true {
		t.Errorf("fail decode on true: %v", item)
	}
}

func TestCborFalse(t *testing.T) {
	buf := make([]byte, 10)
	config := NewDefaultConfig()

	if n := encode(false, buf, config); n != 1 {
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
	config := NewDefaultConfig()
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
		n := encode(test[0], buf, config)
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
		encode(uint64(9223372036854775808), buf, config)
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
	config := NewDefaultConfig()
	n := encode(ref, buf, config)
	t.Logf("%v", buf)
	val, m := decode(buf)
	if n != 5 || n != m || !reflect.DeepEqual(val, ref) {
		t.Errorf("fail code float32: %v %v %T(%v)", n, m, val, val)
	}
}

func TestCborFloat64(t *testing.T) {
	buf, ref := make([]byte, 10), float64(10.11)
	config := NewDefaultConfig()
	n := encode(ref, buf, config)
	val, m := decode(buf)
	if n != 9 || n != m || !reflect.DeepEqual(val, ref) {
		t.Errorf("fail code float32: %v %v %T(%v)", n, m, val, val)
	}
}

func TestCborBytes(t *testing.T) {
	buf, ref := make([]byte, 200), make([]uint8, 100)
	config := NewDefaultConfig()
	for i := 0; i < len(ref); i++ {
		ref[i] = uint8(i)
	}
	n := encode(ref, buf, config)
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
	config := NewDefaultConfig()
	n := encode(ref, buf, config)
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

	// encoding use LengthPrefix
	config := NewConfig(FloatNumber, UnicodeSpace, LengthPrefix)
	n := encode(ref, buf, config)
	val, m := decode(buf)
	if n != 22 || n != m || !reflect.DeepEqual(ref, val) {
		t.Errorf("fail code text: %v %v %T(%v)", n, m, val, val)
	}
	// encoding use Stream
	config = NewConfig(FloatNumber, UnicodeSpace, Stream)
	n = encode(ref, buf, config)
	val, m = decode(buf)
	if n != 23 || n != m || !reflect.DeepEqual(ref, val) {
		t.Errorf("fail code text: %v %v %T(%v)", n, m, val, val)
	}
}

func TestCborMap(t *testing.T) {
	buf := make([]byte, 1024)
	ref := [][2]interface{}{
		[2]interface{}{10.2, "hello world"},
		[2]interface{}{"hello world", 10.2},
	}
	// encoding use LengthPrefix
	config := NewConfig(FloatNumber, UnicodeSpace, LengthPrefix)
	n := encode(ref, buf, config)
	val, m := decode(buf)
	if n != 43 || n != m || !reflect.DeepEqual(ref, val) {
		t.Errorf("fail code text: %v %v %T(%v)", n, m, val, val)
	}
	// encoding use Stream
	config = NewConfig(FloatNumber, UnicodeSpace, Stream)
	n = encode(ref, buf, config)
	val, m = decode(buf)
	if n != 44 || n != m || !reflect.DeepEqual(ref, val) {
		t.Errorf("fail code text: %v %v %T(%v)", n, m, val, val)
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
		p := config.CborEncode(value, cborout)     // golang -> cbor
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
	p := config.CborEncode(value, cborout)     // golang -> cbor
	_, q = config.ToJson(cborout[:p], jsonout) // cbor -> json
	if err := json.Unmarshal(jsonout[:q], &outval); err != nil {
		t.Fatalf("error parsing %v", err)
	} else if err := json.Unmarshal(data, &value); err != nil {
		t.Fatalf("error parsing code.json: %v", err)
	} else if !reflect.DeepEqual(outval, value) {
		t.Errorf("expected %v", value)
		t.Fatalf("got %v", outval)
	}
}

func TestCborMalformed(t *testing.T) {
	config := NewConfig(IntNumber, AnsiSpace, Stream)
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

	// json->cbor->json
	_, n := config.ParseJson(string(data), cborout) // json -> cbor
	_, q := config.ToJson(cborout[:n], jsonout)
	t.Logf("%v %v %v %v", n, q, len(data), len(jsonout[:q]))
	if err := json.Unmarshal(jsonout[:q], &outval); err != nil {
		t.Logf("%v", string(jsonout[:q]))
		t.Fatalf("error parsing code.json.gz: %v", err)
	} else {
		if !reflect.DeepEqual(ref, outval) {
			t.Errorf("expected %v", ref)
			t.Fatalf("got %v", outval)
		}
	}

	// cbor->golang->cbor->json->golang
	value, _ := config.Decode(cborout[:n])     // cbor -> golang
	p := config.CborEncode(value, cborout)     // golang -> cbor
	_, q = config.ToJson(cborout[:p], jsonout) // cbor -> json
	if err := json.Unmarshal(jsonout[:q], &outval); err != nil {
		t.Fatalf("error parsing %v", err)
	} else {
		if !reflect.DeepEqual(outval, ref) {
			t.Errorf("expected %v", value)
			t.Fatalf("got %v", outval)
		}
	}
}

func TestCborTypical(t *testing.T) {
	config := NewDefaultConfig()
	cbordoc, jsonout := make([]byte, 1024*1024), make([]byte, 1024*1024)

	txt := string(testdataFile("../testdata/typical.json"))
	_, n := config.ParseJson(txt, cbordoc)
	p, q := config.ToJson(cbordoc[:n], jsonout)
	if p != n {
		t.Errorf("expected %v, got %v", n, q)
	}
	var ref, out interface{}
	if err := json.Unmarshal([]byte(txt), &ref); err != nil {
		t.Errorf("error parsing typical.json: %v", err)
	} else if err := json.Unmarshal(jsonout[:q], &out); err != nil {
		t.Errorf("error parsing typical.json: %v", err)
	} else if !reflect.DeepEqual(ref, out) {
		t.Errorf("expected %v", ref)
		t.Errorf("got      %v", out)
	}
}

//---- test cases for tag function

func TestDateTime(t *testing.T) {
	buf := make([]byte, 64)
	config := NewDefaultConfig()

	ref, err := time.Parse(time.RFC3339, "2006-01-02T15:04:05+07:00")
	if err != nil {
		t.Errorf("time.Parse() failed: %v", err)
	}

	n := config.CborEncode(ref, buf)
	item, m := config.Decode(buf[:n])
	if n != 28 || n != m {
		t.Errorf("expected %v got %v %v", 28, n, m)
	}
	if !ref.Equal(item.(time.Time)) {
		t.Errorf("expected %v got %v", ref, item.(time.Time))
	}

	// malformed.
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		buf[5] = 'a'
		config.Decode(buf[:n])
	}()
}

func TestEpoch(t *testing.T) {
	buf := make([]byte, 64)
	config := NewDefaultConfig()

	// positive and negative epoch
	for _, val := range [2]int64{1000000, -100000} {
		ref := Epoch(val)
		n := config.CborEncode(ref, buf)
		item, m := config.Decode(buf[:n])
		if n != 6 || n != m {
			t.Errorf("expected %v got %v %v", 6, n, m)
		}
		if !reflect.DeepEqual(ref, item) {
			t.Errorf("expected %v got %v", ref, item)
		}
	}
	// malformed epoch
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		buf[1] = 0x5a // instead of 0x3a
		config.Decode(buf)
	}()
	// malformed epoch
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		buf := make([]byte, 16)
		n := encodeTag(tagEpoch, buf)
		n += encodeBytes([]byte{1, 2}, buf[n:])
		config.Decode(buf)
	}()
}

func TestEpochMicro(t *testing.T) {
	buf := make([]byte, 64)
	config := NewDefaultConfig()
	// positive and negative epoch in uS.
	for _, val := range [2]float64{1000000.123456, -100000.123456} {
		ref := EpochMicro(val)
		n := config.CborEncode(ref, buf)
		item, m := config.Decode(buf[:n])
		if n != 10 || n != m {
			t.Errorf("expected %v got %v %v", 10, n, m)
		}
		if !reflect.DeepEqual(ref, item) {
			t.Errorf("expected %v got %v", ref, item)
		}
	}
}

func TestBigNum(t *testing.T) {
	buf := make([]byte, 64)
	config := NewDefaultConfig()
	// positive and negative bignums
	for _, val := range [2]int64{1000, -1000} {
		bigx := big.NewInt(9223372036854775807)
		bigy := big.NewInt(val)
		bigz := big.NewInt(0).Mul(bigx, bigy)
		n := config.CborEncode(bigz, buf)
		item, m := config.Decode(buf[:n])
		if n != 12 || n != m {
			t.Errorf("expected %v got %v %v", 12, n, m)
		}
		if bigz.Cmp(item.(*big.Int)) != 0 {
			t.Errorf("expected %v got %v", bigz, item.(*big.Int))
		}
	}
}

func TestDecimalFraction(t *testing.T) {
	buf := make([]byte, 64)
	config := NewDefaultConfig()
	// for positive
	refs := []DecimalFraction{
		DecimalFraction([2]interface{}{int64(-10), int64(-23)}),
		DecimalFraction([2]interface{}{int64(-10), int64(23)}),
		DecimalFraction([2]interface{}{int64(10), int64(-23)}),
		DecimalFraction([2]interface{}{int64(10), int64(23)}),
	}
	for _, ref := range refs {
		n := config.CborEncode(ref, buf)
		item, m := config.Decode(buf[:n])
		if n != 3 || n != m {
			t.Errorf("expected %v got %v %v", 3, n, m)
		}
		if !reflect.DeepEqual(ref, item) {
			t.Errorf("expected %v got %v", ref, item)
		}
	}
}

func TestBigFloat(t *testing.T) {
	buf := make([]byte, 64)
	config := NewDefaultConfig()
	refs := []BigFloat{
		BigFloat([2]interface{}{int64(-10), int64(-23)}),
		BigFloat([2]interface{}{int64(-10), int64(23)}),
		BigFloat([2]interface{}{int64(10), int64(-23)}),
		BigFloat([2]interface{}{int64(10), int64(23)}),
	}
	for _, ref := range refs {
		n := config.CborEncode(ref, buf)
		item, m := config.Decode(buf[:n])
		if n != 3 || n != m {
			t.Errorf("expected %v got %v %v", 3, n, m)
		}
		if !reflect.DeepEqual(ref, item) {
			t.Errorf("expected %v got %v", ref, item)
		}
	}
}

func TestCbor(t *testing.T) {
	buf := make([]byte, 64)
	config := NewDefaultConfig()
	ref := Cbor([]byte("hello world"))
	n := config.CborEncode(ref, buf)
	item, m := config.Decode(buf[:n])
	if n != 14 || n != m {
		t.Errorf("expected %v got %v %v", 14, n, m)
	}
	if !reflect.DeepEqual(ref, item) {
		t.Errorf("exptected %v got %v", ref, item)
	}
}

func TestRegexp(t *testing.T) {
	buf := make([]byte, 64)
	config := NewDefaultConfig()
	ref, err := regexp.Compile(`a([0-9]t*)+`)
	if err != nil {
		t.Errorf("compiling regex")
	}
	n := config.CborEncode(ref, buf)
	item, m := config.Decode(buf[:n])
	if n != 14 || n != m {
		t.Errorf("expected %v got %v %v", 14, n, m)
	}
	if ref.String() != (item.(*regexp.Regexp)).String() {
		t.Errorf("expected %v got %v", ref, item)
	}
	// malformed reg-ex
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		n := encodeTag(tagRegexp, buf)
		n += encodeText(`a([0-9]t*+`, buf[n:])
		config.Decode(buf)
	}()
}

func TestCborPrefix(t *testing.T) {
	buf := make([]byte, 64)
	config := NewDefaultConfig()
	ref := CborPrefix([]byte("hello world"))
	n := config.CborEncode(ref, buf)
	item, m := config.Decode(buf[:n])
	if n != 15 || n != m {
		t.Errorf("expected %v got %v %v", 15, n, m)
	}
	if !reflect.DeepEqual(ref, item) {
		t.Errorf("exptected %v got %v", ref, item)
	}
}

//---- benchmarks

func BenchmarkEncodeNull(b *testing.B) {
	buf := make([]byte, 10)
	config := NewDefaultConfig()
	for i := 0; i < b.N; i++ {
		encode(nil, buf, config)
	}
}

func BenchmarkDecodeNull(b *testing.B) {
	buf := make([]byte, 10)
	config := NewDefaultConfig()
	n := encode(nil, buf, config)
	for i := 0; i < b.N; i++ {
		decode(buf[:n])
	}
}

func BenchmarkEncodeTrue(b *testing.B) {
	buf := make([]byte, 64)
	config := NewDefaultConfig()
	for i := 0; i < b.N; i++ {
		encode(true, buf, config)
	}
}

func BenchmarkDecodeTrue(b *testing.B) {
	buf := make([]byte, 10)
	config := NewDefaultConfig()
	n := encode(true, buf, config)
	for i := 0; i < b.N; i++ {
		decode(buf[:n])
	}
}

func BenchmarkEncodeFalse(b *testing.B) {
	buf := make([]byte, 10)
	config := NewDefaultConfig()
	for i := 0; i < b.N; i++ {
		encode(false, buf, config)
	}
}

func BenchmarkDecodeFalse(b *testing.B) {
	buf := make([]byte, 10)
	config := NewDefaultConfig()
	n := encode(false, buf, config)
	for i := 0; i < b.N; i++ {
		decode(buf[:n])
	}
}

func BenchmarkEncodeUint8(b *testing.B) {
	buf := make([]byte, 64)
	config := NewDefaultConfig()
	for i := 0; i < b.N; i++ {
		encode(uint8(255), buf, config)
	}
}

func BenchmarkDecodeUint8(b *testing.B) {
	buf := make([]byte, 64)
	config := NewDefaultConfig()
	n := encode(uint8(255), buf, config)
	for i := 0; i < b.N; i++ {
		decode(buf[:n])
	}
}

func BenchmarkEncodeInt8(b *testing.B) {
	buf := make([]byte, 64)
	config := NewDefaultConfig()
	for i := 0; i < b.N; i++ {
		encode(int8(-128), buf, config)
	}
}

func BenchmarkDecodeInt8(b *testing.B) {
	buf := make([]byte, 64)
	config := NewDefaultConfig()
	n := encode(int8(-128), buf, config)
	for i := 0; i < b.N; i++ {
		decode(buf[:n])
	}
}

func BenchmarkEncodeUint16(b *testing.B) {
	buf := make([]byte, 64)
	config := NewDefaultConfig()
	for i := 0; i < b.N; i++ {
		encode(uint16(65535), buf, config)
	}
}

func BenchmarkDecodeUint16(b *testing.B) {
	buf := make([]byte, 64)
	config := NewDefaultConfig()
	n := encode(uint16(65535), buf, config)
	for i := 0; i < b.N; i++ {
		decode(buf[:n])
	}
}

func BenchmarkEncodeInt16(b *testing.B) {
	buf := make([]byte, 64)
	config := NewDefaultConfig()
	for i := 0; i < b.N; i++ {
		encode(int16(-32768), buf, config)
	}
}

func BenchmarkDecodeInt16(b *testing.B) {
	buf := make([]byte, 64)
	config := NewDefaultConfig()
	n := encode(int16(-32768), buf, config)
	for i := 0; i < b.N; i++ {
		decode(buf[:n])
	}
}

func BenchmarkEncodeUint32(b *testing.B) {
	buf := make([]byte, 64)
	config := NewDefaultConfig()
	for i := 0; i < b.N; i++ {
		encode(uint32(4294967295), buf, config)
	}
}

func BenchmarkDecodeUint32(b *testing.B) {
	buf := make([]byte, 64)
	config := NewDefaultConfig()
	n := encode(uint32(4294967295), buf, config)
	for i := 0; i < b.N; i++ {
		decode(buf[:n])
	}
}

func BenchmarkEncodeInt32(b *testing.B) {
	buf := make([]byte, 64)
	config := NewDefaultConfig()
	for i := 0; i < b.N; i++ {
		encode(int32(-2147483648), buf, config)
	}
}

func BenchmarkDecodeInt32(b *testing.B) {
	buf := make([]byte, 64)
	config := NewDefaultConfig()
	n := encode(int32(-2147483648), buf, config)
	for i := 0; i < b.N; i++ {
		decode(buf[:n])
	}
}

func BenchmarkEncodeUint64(b *testing.B) {
	buf := make([]byte, 64)
	config := NewDefaultConfig()
	for i := 0; i < b.N; i++ {
		encode(uint64(18446744073709551615), buf, config)
	}
}

func BenchmarkDecodeUint64(b *testing.B) {
	buf := make([]byte, 64)
	config := NewDefaultConfig()
	n := encode(uint64(18446744073709551615), buf, config)
	for i := 0; i < b.N; i++ {
		decode(buf[:n])
	}
}

func BenchmarkEncodeInt64(b *testing.B) {
	buf := make([]byte, 64)
	config := NewDefaultConfig()
	for i := 0; i < b.N; i++ {
		encode(int64(-2147483648), buf, config)
	}
}

func BenchmarkDecodeInt64(b *testing.B) {
	buf := make([]byte, 64)
	config := NewDefaultConfig()
	n := encode(int64(-2147483648), buf, config)
	for i := 0; i < b.N; i++ {
		decode(buf[:n])
	}
}

func BenchmarkEncodeFlt32(b *testing.B) {
	buf := make([]byte, 64)
	config := NewDefaultConfig()
	for i := 0; i < b.N; i++ {
		encode(float32(10.2), buf, config)
	}
}

func BenchmarkDecodeFlt32(b *testing.B) {
	buf := make([]byte, 64)
	config := NewDefaultConfig()
	n := encode(float32(10.2), buf, config)
	for i := 0; i < b.N; i++ {
		decode(buf[:n])
	}
}

func BenchmarkEncodeFlt64(b *testing.B) {
	buf := make([]byte, 64)
	config := NewDefaultConfig()
	for i := 0; i < b.N; i++ {
		encode(float64(10.2), buf, config)
	}
}

func BenchmarkDecodeFlt64(b *testing.B) {
	buf := make([]byte, 64)
	config := NewDefaultConfig()
	n := encode(float64(10.2), buf, config)
	for i := 0; i < b.N; i++ {
		decode(buf[:n])
	}
}

func BenchmarkEncodeBytes(b *testing.B) {
	buf := make([]byte, 64)
	config := NewDefaultConfig()
	bs := []byte("hello world")
	for i := 0; i < b.N; i++ {
		encode(bs, buf, config)
	}
}

func BenchmarkDecodeBytes(b *testing.B) {
	buf := make([]byte, 64)
	config := NewDefaultConfig()
	n := encode([]byte("hello world"), buf, config)
	for i := 0; i < b.N; i++ {
		decode(buf[:n])
	}
}

func BenchmarkEncodeText(b *testing.B) {
	buf := make([]byte, 64)
	config := NewDefaultConfig()
	text := "hello world"
	for i := 0; i < b.N; i++ {
		encode(text, buf, config)
	}
}

func BenchmarkDecodeText(b *testing.B) {
	buf := make([]byte, 64)
	config := NewDefaultConfig()
	n := encode("hello world", buf, config)
	for i := 0; i < b.N; i++ {
		decode(buf[:n])
	}
}

func BenchmarkEncodeArr0(b *testing.B) {
	buf, arr := make([]byte, 1024), make([]interface{}, 0)
	config := NewDefaultConfig()
	for i := 0; i < b.N; i++ {
		encode(arr, buf, config)
	}
}

func BenchmarkDecodeArr0(b *testing.B) {
	buf, arr := make([]byte, 1024), make([]interface{}, 0)
	config := NewDefaultConfig()
	n := encode(arr, buf, config)
	for i := 0; i < b.N; i++ {
		decode(buf[:n])
	}
}

func BenchmarkEncodeArr5(b *testing.B) {
	buf := make([]byte, 1024)
	config := NewDefaultConfig()
	arr := []interface{}{5, 5.0, "hello world", true, nil}
	for i := 0; i < b.N; i++ {
		encode(arr, buf, config)
	}
}

func BenchmarkDecodeArr5(b *testing.B) {
	buf := make([]byte, 1024)
	config := NewDefaultConfig()
	arr := []interface{}{5, 5.0, "hello world", true, nil}
	n := encode(arr, buf, config)
	for i := 0; i < b.N; i++ {
		decode(buf[:n])
	}
}

func BenchmarkEncodeMap0(b *testing.B) {
	buf := make([]byte, 1024)
	config := NewDefaultConfig()
	m := make([][2]interface{}, 0)
	for i := 0; i < b.N; i++ {
		encode(m, buf, config)
	}
}

func BenchmarkDecodeMap0(b *testing.B) {
	buf := make([]byte, 1024)
	config := NewDefaultConfig()
	m := make([][2]interface{}, 0)
	n := encode(m, buf, config)
	for i := 0; i < b.N; i++ {
		decode(buf[:n])
	}
}

func BenchmarkEncodeMap5(b *testing.B) {
	buf := make([]byte, 1024)
	config := NewDefaultConfig()
	m := [][2]interface{}{
		[2]interface{}{"key0", 5}, [2]interface{}{"key1", 5.0},
		[2]interface{}{"key2", "hello world"},
		[2]interface{}{"key3", true}, [2]interface{}{"key4", nil},
	}
	for i := 0; i < b.N; i++ {
		encode(m, buf, config)
	}
}

func BenchmarkDecodeMap5(b *testing.B) {
	buf := make([]byte, 1024)
	config := NewDefaultConfig()
	m := [][2]interface{}{
		[2]interface{}{"key0", 5}, [2]interface{}{"key1", 5.0},
		[2]interface{}{"key2", "hello world"},
		[2]interface{}{"key3", true}, [2]interface{}{"key4", nil},
	}
	n := encode(m, buf, config)
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
