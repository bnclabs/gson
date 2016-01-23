//  Copyright (c) 2015 Couchbase, Inc.

package gson

import "testing"
import "fmt"
import "reflect"
import "encoding/json"
import "time"
import "regexp"
import "math/big"

var _ = fmt.Sprintf("dummy")

func TestCborNil(t *testing.T) {
	buf := make([]byte, 10)
	config := NewDefaultConfig()

	if n := value2cbor(nil, buf, config); n != 1 {
		t.Errorf("fail value2nil nil: %v want 1", n)
	} else if item, m := cbor2value(buf, config); m != 1 {
		t.Errorf("fail cbor2value on nil len: %v want 1", m)
	} else if item != nil {
		t.Errorf("fail cbor2value on nil: %x", item)
	}
}

func TestCborTrue(t *testing.T) {
	buf := make([]byte, 10)
	config := NewDefaultConfig()

	if n := value2cbor(true, buf, config); n != 1 {
		t.Errorf("fail value2cbor true: %v want 1", n)
	} else if item, m := cbor2value(buf, config); m != 1 {
		t.Errorf("fail cbor2value on true len: %v want 1", m)
	} else if item.(bool) != true {
		t.Errorf("fail cbor2value on true: %v", item)
	}
}

func TestCborFalse(t *testing.T) {
	buf := make([]byte, 10)
	config := NewDefaultConfig()

	if n := value2cbor(false, buf, config); n != 1 {
		t.Errorf("fail value2cbor false: %v want 1", n)
	} else if item, m := cbor2value(buf, config); m != 1 {
		t.Errorf("fail cbor2value on false len: %v want 1", m)
	} else if item.(bool) != false {
		t.Errorf("fail cbor2value on false: %v", item)
	}
}

func TestCborUint8(t *testing.T) {
	config := NewDefaultConfig()
	buf := make([]byte, 20)
	for i := uint16(0); i <= 255; i++ {
		n := valuint82cbor(uint8(i), buf)
		val, m := cbor2value(buf, config)
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
	config := NewDefaultConfig()
	buf := make([]byte, 20)
	for i := int16(-128); i <= 127; i++ {
		n := valint82cbor(int8(i), buf)
		val, m := cbor2value(buf, config)
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
		n := value2cbor(test[0], buf, config)
		val, m := cbor2value(buf, config)
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
		value2cbor(uint64(9223372036854775808), buf, config)
		buf[0] = (buf[0] & 0x1f) | cborType1 // fix as negative integer
		cbor2value(buf, config)
	}()
}

func TestCborFloat16(t *testing.T) {
	config := NewDefaultConfig()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic while decoding float16")
		}
	}()
	cbor2value([]byte{0xf9, 0, 0, 0, 0}, config)
}

func TestCborFloat32(t *testing.T) {
	buf, ref := make([]byte, 10), float32(10.11)
	config := NewDefaultConfig()
	n := value2cbor(ref, buf, config)
	t.Logf("%v", buf)
	val, m := cbor2value(buf, config)
	if n != 5 || n != m || !reflect.DeepEqual(val, ref) {
		t.Errorf("fail code float32: %v %v %T(%v)", n, m, val, val)
	}
}

func TestCborFloat64(t *testing.T) {
	buf, ref := make([]byte, 10), float64(10.11)
	config := NewDefaultConfig()
	n := value2cbor(ref, buf, config)
	val, m := cbor2value(buf, config)
	if n != 9 || n != m || !reflect.DeepEqual(val, ref) {
		t.Errorf("fail code float32: %v %v %T(%v)", n, m, val, val)
	}
}

func TestCborTagBytes(t *testing.T) {
	buf, ref := make([]byte, 200), make([]uint8, 100)
	config := NewDefaultConfig()
	for i := 0; i < len(ref); i++ {
		ref[i] = uint8(i)
	}
	n := value2cbor(ref, buf, config)
	val, m := cbor2value(buf, config)
	if n != 102 || n != m || !reflect.DeepEqual(val, ref) {
		t.Errorf("fail code bytes: %v %v %T(%v)", n, m, val, val)
	}
	// test byte-start
	if n := bytesStart(buf); n != 1 {
		t.Errorf("fail code bytes-start len: %v wanted 1", n)
	} else if val, m := cbor2value(buf, config); m != n {
		t.Errorf("fail code bytes-start size: %v wanted %v", m, n)
	} else if !reflect.DeepEqual(val, CborIndefinite(0x5f)) {
		t.Errorf("fail code bytes-start: %v wanted 0x5f", buf[0])
	}
}

func TestCborText(t *testing.T) {
	buf, ref := make([]byte, 200), "hello world"
	config := NewDefaultConfig()
	n := value2cbor(ref, buf, config)
	val, m := cbor2value(buf, config)
	if n != 12 || n != m || !reflect.DeepEqual(val, ref) {
		t.Errorf("fail code text: %v %v %T(%v)", n, m, val, val)
	}

	// test text-start
	if n := textStart(buf); n != 1 {
		t.Errorf("fail code text-start len: %v wanted 1", n)
	} else if val, m := cbor2value(buf, config); m != n {
		t.Errorf("fail code text-start size: %v wanted %v", m, n)
	} else if !reflect.DeepEqual(val, CborIndefinite(0x7f)) {
		t.Errorf("fail code text-start: %x wanted 0x7f", buf[0])
	}
}

func TestCborArray(t *testing.T) {
	buf := make([]byte, 1024)
	ref := []interface{}{10.2, "hello world"}

	// encoding use LengthPrefix
	config := NewDefaultConfig()
	config = config.SetNumberKind(FloatNumber).SetSpaceKind(UnicodeSpace)
	config = config.SetContainerEncoding(LengthPrefix)
	n := value2cbor(ref, buf, config)
	val, m := cbor2value(buf, config)
	if n != 22 || n != m || !reflect.DeepEqual(ref, val) {
		t.Errorf("fail code text: %v %v %T(%v)", n, m, val, val)
	}
	// encoding use Stream
	config = NewDefaultConfig()
	config = config.SetNumberKind(FloatNumber).SetSpaceKind(UnicodeSpace)
	n = value2cbor(ref, buf, config)
	val, m = cbor2value(buf, config)
	if n != 23 || n != m || !reflect.DeepEqual(ref, val) {
		t.Errorf("fail code text: %v %v %T(%v)", n, m, val, val)
	}
}

func TestCborMapSlice(t *testing.T) {
	buf := make([]byte, 1024)
	ref := [][2]interface{}{
		[2]interface{}{"10.2", "hello world"},
		[2]interface{}{"hello world", 10.2},
	}
	refm := CborMap2golangMap(ref)
	// encoding use LengthPrefix
	config := NewDefaultConfig()
	config = config.SetNumberKind(FloatNumber).SetSpaceKind(UnicodeSpace)
	config = config.SetContainerEncoding(LengthPrefix)
	n := value2cbor(ref, buf, config)
	val, m := cbor2value(buf, config)
	if n != 39 || n != m || !reflect.DeepEqual(refm, val) {
		t.Errorf("fail code text: %v %v %v %v", n, m, refm, val)
	}
	// encoding use Stream
	config = NewDefaultConfig()
	config = config.SetNumberKind(FloatNumber).SetSpaceKind(UnicodeSpace)
	n = value2cbor(ref, buf, config)
	val, m = cbor2value(buf, config)
	if n != 40 || n != m || !reflect.DeepEqual(refm, val) {
		t.Errorf("fail code text: %v %v %v %v", n, m, refm, val)
	}
}

func TestCborMap(t *testing.T) {
	buf := make([]byte, 1024)
	ref := map[string]interface{}{
		"10.2":        "hello world",
		"hello world": float64(10.2),
	}
	// encoding use LengthPrefix
	config := NewDefaultConfig()
	config = config.SetNumberKind(FloatNumber).SetSpaceKind(UnicodeSpace)
	config = config.SetContainerEncoding(LengthPrefix)
	n := value2cbor(ref, buf, config)
	val, m := cbor2value(buf, config)
	val = CborMap2golangMap(val)
	if n != 40 || n != m || !reflect.DeepEqual(ref, val) {
		t.Errorf("fail code text: %v %v %v %v", n, m, ref, val)
	}
	// encoding use Stream
	config = NewDefaultConfig()
	config = config.SetNumberKind(FloatNumber).SetSpaceKind(UnicodeSpace)
	n = value2cbor(ref, buf, config)
	val, m = cbor2value(buf, config)
	val = CborMap2golangMap(val)
	if n != 40 || n != m || !reflect.DeepEqual(ref, val) {
		t.Errorf("fail code text: %v %v %v %v", n, m, ref, val)
	}
}

func TestCborBreakStop(t *testing.T) {
	buf := make([]byte, 10)
	config := NewDefaultConfig()
	if n := breakStop(buf); n != 1 {
		t.Errorf("fail code text-start len: %v wanted 1", n)
	} else if val, m := cbor2value(buf, config); m != n {
		t.Errorf("fail code text-start: %v wanted %v", m, n)
	} else if !reflect.DeepEqual(val, CborBreakStop(0xff)) {
		t.Errorf("fail code text-start: %x wanted 0xff", buf[0])
	}
}

func TestCborUndefined(t *testing.T) {
	buf := make([]byte, 10)
	config := NewDefaultConfig()
	if n := valundefined2cbor(buf); n != 1 {
		t.Errorf("fail value2cbor undefined: %v want 1", n)
	} else if item, m := cbor2value(buf, config); m != 1 {
		t.Errorf("fail cbor2value on undefined len: %v want 1", m)
	} else if item.(CborUndefined) != CborUndefined(cborSimpleUndefined) {
		t.Errorf("fail cbor2value on undefined: %T %v", item, item)
	}
}

func TestCborReserved(t *testing.T) {
	config := NewDefaultConfig()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic while decoding reserved")
		}
	}()
	cbor2value([]byte{cborHdr(cborType0, 28)}, config)
}

func TestCborMaster(t *testing.T) {
	var outval, ref interface{}

	testcases := append(scanvalid, []string{
		string(mapValue),
		string(allValueIndent),
		string(allValueCompact),
		string(pallValueIndent),
		string(pallValueCompact),
	}...)

	config := NewDefaultConfig()
	jsn := config.NewJson(nil, 0)
	cbr := config.NewCbor(make([]byte, 1024*1024), 0)
	jsn1 := config.NewJson(make([]byte, 1024*1024), 0)
	cbr1 := config.NewCbor(make([]byte, 1024*1024), 0)
	jsn2 := config.NewJson(make([]byte, 1024*1024), 0)

	for _, tcase := range testcases {
		t.Logf("%v", tcase)
		json.Unmarshal([]byte(tcase), &ref)
		jsn.Reset([]byte(tcase))

		// test JsonToCbor/CborToJson
		jsn.Tocbor(cbr.Reset(nil))
		cbr.Tojson(jsn1.Reset(nil))
		if err := json.Unmarshal(jsn1.Bytes(), &outval); err != nil {
			t.Fatalf("error parsing %v: %v", string(jsn1.Bytes()), err)
		} else if !reflect.DeepEqual(outval, ref) {
			t.Fatalf("expected '%v', got '%v'", ref, outval)
		}

		val := config.NewValue(cbr.Tovalue())
		val.Tocbor(cbr1.Reset(nil))
		cbr1.Tojson(jsn2.Reset(nil))
		if err := json.Unmarshal(jsn2.Bytes(), &outval); err != nil {
			t.Fatalf("error parsing %q: %v", string(jsn2.Bytes()), err)
		} else if !reflect.DeepEqual(outval, ref) {
			fmsg := "expected {%T,%v}, got {%T,%v}"
			t.Fatalf(fmsg, val.data, val.data, outval, outval)
		}
	}
}

func TestCborSmartnum(t *testing.T) {
	var outval, ref interface{}

	data := testdataFile("testdata/smartnum")
	json.Unmarshal(data, &ref)

	config := NewDefaultConfig()
	jsn := config.NewJson(data, -1)
	cbr := config.NewCbor(make([]byte, 1024*1024), 0)
	jsnback := config.NewJson(make([]byte, 1024*1024), 0)

	// test JsonToCbor/CborToJson
	jsn.Tocbor(cbr)
	cbr.Tojson(jsnback)
	if err := json.Unmarshal(jsnback.Bytes(), &outval); err != nil {
		t.Logf("%v", string(jsnback.Bytes()))
		t.Fatalf("error parsing code.json.gz: %v", err)
	} else if !reflect.DeepEqual(ref, outval) {
		t.Errorf("expected %v", ref)
		t.Errorf("got-json %v", string(jsnback.Bytes()))
		t.Fatalf("got %v", outval)
	}

	val := config.NewValue(cbr.Tovalue())

	jsn = config.NewJson(make([]byte, 1024*1024), 0)
	cbr = config.NewCbor(make([]byte, 1024*1024), 0)

	val.Tocbor(cbr)
	cbr.Tojson(jsn)
	if err := json.Unmarshal(jsn.Bytes(), &outval); err != nil {
		t.Fatalf("error parsing %v", err)
	} else if err := json.Unmarshal(data, &val.data); err != nil {
		t.Fatalf("error parsing code.json: %v", err)
	} else if !reflect.DeepEqual(outval, val.data) {
		t.Errorf("expected %v", val.data)
		t.Fatalf("got %v", outval)
	}
}

func TestCborMalformed(t *testing.T) {
	for _, tcase := range scaninvalid {
		func() {
			config := NewDefaultConfig()
			config = config.SetNumberKind(IntNumber).SetSpaceKind(AnsiSpace)
			jsn := config.NewJson(make([]byte, 1024), 0)
			cbr := config.NewCbor(make([]byte, 1024), 0)

			defer func() {
				if r := recover(); r == nil {
					t.Fatalf("expected panic")
				}
			}()
			t.Logf("%v", tcase)
			jsn.Tocbor(cbr)
		}()
	}
}

func TestCborCodeJSON(t *testing.T) {
	var ref, outval interface{}

	data := testdataFile("testdata/code.json.gz")
	json.Unmarshal(data, &ref)

	config := NewDefaultConfig()
	jsn := config.NewJson(data, -1)
	cbr := config.NewCbor(make([]byte, 10*1024*1024), 0)
	jsnback := config.NewJson(make([]byte, 10*1024*1024), 0)

	// json->cbor->json
	jsn.Tocbor(cbr)
	cbr.Tojson(jsnback)
	t.Logf("%v %v", len(data), len(jsnback.Bytes()))
	if err := json.Unmarshal(jsnback.Bytes(), &outval); err != nil {
		t.Logf("%v", string(jsnback.Bytes()))
		t.Fatalf("error parsing code.json.gz: %v", err)
	} else {
		if !reflect.DeepEqual(ref, outval) {
			t.Errorf("expected %v", ref)
			t.Fatalf("got %v", outval)
		}
	}

	jsn = config.NewJson(make([]byte, 10*1024*1024), 0)
	cbrback := config.NewCbor(make([]byte, 10*1024*1024), 0)

	// cbor->golang->cbor->json->golang
	val := config.NewValue(cbr.Tovalue())
	val.Tocbor(cbrback)
	cbrback.Tojson(jsn)
	if err := json.Unmarshal(jsn.Bytes(), &outval); err != nil {
		t.Fatalf("error parsing %v", err)
	} else {
		if !reflect.DeepEqual(outval, ref) {
			t.Errorf("expected %v", val.data)
			t.Fatalf("got %v", outval)
		}
	}
}

func TestCborTypical(t *testing.T) {
	var ref, out interface{}

	data := testdataFile("testdata/typical.json")
	json.Unmarshal(data, &ref)

	config := NewDefaultConfig()
	jsn := config.NewJson(data, -1)
	cbr := config.NewCbor(make([]byte, 1024*1024), 0)
	jsnback := config.NewJson(make([]byte, 1024*1024), 0)

	jsn.Tocbor(cbr)
	cbr.Tojson(jsnback)
	if err := json.Unmarshal(jsnback.Bytes(), &out); err != nil {
		t.Errorf("error parsing typical.json: %v", err)
	} else if !reflect.DeepEqual(ref, out) {
		t.Errorf("expected %v", ref)
		t.Errorf("got      %v", out)
	}
}

//---- test cases for tag function

func TestDateTime(t *testing.T) {
	ref, err := time.Parse(time.RFC3339, "2006-01-02T15:04:05+07:00")
	if err != nil {
		t.Errorf("time.Parse() failed: %v", err)
	}

	config := NewDefaultConfig()
	cbr := config.NewCbor(make([]byte, 64), 0)

	val := config.NewValue(ref)
	val.Tocbor(cbr)
	if item := cbr.Tovalue(); !ref.Equal(item.(time.Time)) {
		t.Errorf("expected %v got %v", ref, item.(time.Time))
	}

	// malformed.
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		cbr.data[5] = 'a'
		cbr.Tovalue()
	}()
}

func TestTagEpoch(t *testing.T) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(make([]byte, 64), 0)

	// positive and negative epoch
	for _, v := range [2]int64{1000000, -100000} {
		val := config.NewValue(CborTagEpoch(v))
		val.Tocbor(cbr.Reset(nil))
		if item := cbr.Tovalue(); !reflect.DeepEqual(val.data, item) {
			t.Errorf("expected %v got %v", val.data, item)
		}
	}

	// malformed epoch
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		cbr.data[1] = 0x5a // instead of 0x3a
		cbr.Tovalue()
	}()

	// malformed epoch
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		buf := make([]byte, 16)
		n := tag2cbor(tagEpoch, buf)
		n += valbytes2cbor([]byte{1, 2}, buf[n:])
		config.NewCbor(buf, n).Tovalue()
	}()
}

func TestTagEpochMicro(t *testing.T) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(make([]byte, 64), 0)

	// positive and negative epoch in uS.
	for _, v := range [2]float64{1000000.123456, -100000.123456} {
		val := config.NewValue(CborTagEpochMicro(v))
		val.Tocbor(cbr.Reset(nil))
		if item := cbr.Tovalue(); !reflect.DeepEqual(val.data, item) {
			t.Errorf("expected %v got %v", val.data, item)
		}
	}
}

func TestBigNum(t *testing.T) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(make([]byte, 64), 0)

	// positive and negative bignums
	for _, v := range [2]int64{1000, -1000} {
		z := big.NewInt(0).Mul(big.NewInt(9223372036854775807), big.NewInt(v))
		val := config.NewValue(z)
		val.Tocbor(cbr.Reset(nil))
		if item := cbr.Tovalue(); z.Cmp(item.(*big.Int)) != 0 {
			t.Errorf("expected %v got %v", z, item.(*big.Int))
		}
	}
}

func TestDecimalFraction(t *testing.T) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(make([]byte, 64), 0)

	// for positive
	refs := []CborTagFraction{
		CborTagFraction([2]int64{int64(-10), int64(-23)}),
		CborTagFraction([2]int64{int64(-10), int64(23)}),
		CborTagFraction([2]int64{int64(10), int64(-23)}),
		CborTagFraction([2]int64{int64(10), int64(23)}),
	}

	for _, ref := range refs {
		val := config.NewValue(ref)
		val.Tocbor(cbr.Reset(nil))
		if item := cbr.Tovalue(); !reflect.DeepEqual(ref, item) {
			t.Errorf("expected %v got %v", ref, item)
		}
	}
}

func TestBigFloat(t *testing.T) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(make([]byte, 64), 0)

	refs := []CborTagFloat{
		CborTagFloat([2]int64{int64(-10), int64(-23)}),
		CborTagFloat([2]int64{int64(-10), int64(23)}),
		CborTagFloat([2]int64{int64(10), int64(-23)}),
		CborTagFloat([2]int64{int64(10), int64(23)}),
	}
	for _, ref := range refs {
		val := config.NewValue(ref)
		val.Tocbor(cbr.Reset(nil))
		if item := cbr.Tovalue(); !reflect.DeepEqual(ref, item) {
			t.Errorf("expected %v got %v", ref, item)
		}
	}
}

func TestCbor(t *testing.T) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(make([]byte, 64), 0)

	val := config.NewValue(CborTagBytes([]byte("hello world")))
	val.Tocbor(cbr)
	if item := cbr.Tovalue(); !reflect.DeepEqual(val.data, item) {
		t.Errorf("exptected %v got %v", val.data, item)
	}
}

func TestRegexp(t *testing.T) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(make([]byte, 64), 0)

	ref, _ := regexp.Compile(`a([0-9]t*)+`)
	val := config.NewValue(ref)
	val.Tocbor(cbr)
	item := cbr.Tovalue()
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
		buf := make([]byte, 1024)
		n := tag2cbor(tagRegexp, buf)
		n += valtext2cbor(`a([0-9]t*+`, buf[n:])
		config.NewCbor(buf, n).Tovalue()
	}()
}

func TestCborTagPrefix(t *testing.T) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(make([]byte, 64), 0)

	val := config.NewValue(CborTagPrefix([]byte("hello world")))
	val.Tocbor(cbr)
	if item := cbr.Tovalue(); !reflect.DeepEqual(val.data, item) {
		t.Errorf("exptected %v got %v", val.data, item)
	}
}

//---- benchmarks

func BenchmarkVal2CborNull(b *testing.B) {
	buf := make([]byte, 10)
	config := NewDefaultConfig()
	for i := 0; i < b.N; i++ {
		value2cbor(nil, buf, config)
	}
}

func BenchmarkCbor2ValNull(b *testing.B) {
	buf := make([]byte, 10)
	config := NewDefaultConfig()
	n := value2cbor(nil, buf, config)
	for i := 0; i < b.N; i++ {
		cbor2value(buf[:n], config)
	}
}

func BenchmarkVal2CborTrue(b *testing.B) {
	buf := make([]byte, 64)
	config := NewDefaultConfig()
	val := interface{}(true)
	for i := 0; i < b.N; i++ {
		value2cbor(val, buf, config)
	}
}

func BenchmarkCbor2ValTrue(b *testing.B) {
	buf := make([]byte, 10)
	config := NewDefaultConfig()
	n := value2cbor(true, buf, config)
	for i := 0; i < b.N; i++ {
		cbor2value(buf[:n], config)
	}
}

func BenchmarkVal2CborFalse(b *testing.B) {
	buf := make([]byte, 10)
	config := NewDefaultConfig()
	val := interface{}(false)
	for i := 0; i < b.N; i++ {
		value2cbor(val, buf, config)
	}
}

func BenchmarkCbor2ValFalse(b *testing.B) {
	buf := make([]byte, 10)
	config := NewDefaultConfig()
	n := value2cbor(false, buf, config)
	for i := 0; i < b.N; i++ {
		cbor2value(buf[:n], config)
	}
}

func BenchmarkVal2CborUint8(b *testing.B) {
	buf := make([]byte, 64)
	config := NewDefaultConfig()
	val := interface{}(uint8(255))
	for i := 0; i < b.N; i++ {
		value2cbor(val, buf, config)
	}
}

func BenchmarkCbor2ValUint8(b *testing.B) {
	buf := make([]byte, 64)
	config := NewDefaultConfig()
	n := value2cbor(uint8(255), buf, config)
	for i := 0; i < b.N; i++ {
		cbor2value(buf[:n], config)
	}
}

func BenchmarkVal2CborInt8(b *testing.B) {
	buf := make([]byte, 64)
	config := NewDefaultConfig()
	val := interface{}(int8(-128))
	for i := 0; i < b.N; i++ {
		value2cbor(val, buf, config)
	}
}

func BenchmarkCbor2ValInt8(b *testing.B) {
	buf := make([]byte, 64)
	config := NewDefaultConfig()
	n := value2cbor(int8(-128), buf, config)
	for i := 0; i < b.N; i++ {
		cbor2value(buf[:n], config)
	}
}

func BenchmarkVal2CborUint16(b *testing.B) {
	buf := make([]byte, 64)
	config := NewDefaultConfig()
	val := interface{}(uint16(65535))
	for i := 0; i < b.N; i++ {
		value2cbor(val, buf, config)
	}
}

func BenchmarkCbor2ValUint16(b *testing.B) {
	buf := make([]byte, 64)
	config := NewDefaultConfig()
	n := value2cbor(uint16(65535), buf, config)
	for i := 0; i < b.N; i++ {
		cbor2value(buf[:n], config)
	}
}

func BenchmarkVal2CborInt16(b *testing.B) {
	buf := make([]byte, 64)
	config := NewDefaultConfig()
	val := interface{}(int16(-32768))
	for i := 0; i < b.N; i++ {
		value2cbor(val, buf, config)
	}
}

func BenchmarkCbor2ValInt16(b *testing.B) {
	buf := make([]byte, 64)
	config := NewDefaultConfig()
	n := value2cbor(int16(-32768), buf, config)
	for i := 0; i < b.N; i++ {
		cbor2value(buf[:n], config)
	}
}

func BenchmarkVal2CborUint32(b *testing.B) {
	buf := make([]byte, 64)
	config := NewDefaultConfig()
	val := interface{}(uint32(4294967295))
	for i := 0; i < b.N; i++ {
		value2cbor(val, buf, config)
	}
}

func BenchmarkCbor2ValUint32(b *testing.B) {
	buf := make([]byte, 64)
	config := NewDefaultConfig()
	n := value2cbor(uint32(4294967295), buf, config)
	for i := 0; i < b.N; i++ {
		cbor2value(buf[:n], config)
	}
}

func BenchmarkVal2CborInt32(b *testing.B) {
	buf := make([]byte, 64)
	config := NewDefaultConfig()
	val := interface{}(int32(-2147483648))
	for i := 0; i < b.N; i++ {
		value2cbor(val, buf, config)
	}
}

func BenchmarkCbor2ValInt32(b *testing.B) {
	buf := make([]byte, 64)
	config := NewDefaultConfig()
	n := value2cbor(int32(-2147483648), buf, config)
	for i := 0; i < b.N; i++ {
		cbor2value(buf[:n], config)
	}
}

func BenchmarkVal2CborUint64(b *testing.B) {
	buf := make([]byte, 64)
	config := NewDefaultConfig()
	val := interface{}(uint64(18446744073709551615))
	for i := 0; i < b.N; i++ {
		value2cbor(val, buf, config)
	}
}

func BenchmarkCbor2ValUint64(b *testing.B) {
	buf := make([]byte, 64)
	config := NewDefaultConfig()
	n := value2cbor(uint64(18446744073709551615), buf, config)
	for i := 0; i < b.N; i++ {
		cbor2value(buf[:n], config)
	}
}

func BenchmarkVal2CborInt64(b *testing.B) {
	buf := make([]byte, 64)
	config := NewDefaultConfig()
	val := interface{}(int64(-2147483648))
	for i := 0; i < b.N; i++ {
		value2cbor(val, buf, config)
	}
}

func BenchmarkCbor2ValInt64(b *testing.B) {
	buf := make([]byte, 64)
	config := NewDefaultConfig()
	n := value2cbor(int64(-2147483648), buf, config)
	for i := 0; i < b.N; i++ {
		cbor2value(buf[:n], config)
	}
}

func BenchmarkVal2CborFlt32(b *testing.B) {
	buf := make([]byte, 64)
	config := NewDefaultConfig()
	val := interface{}(float32(10.2))
	for i := 0; i < b.N; i++ {
		value2cbor(val, buf, config)
	}
}

func BenchmarkCbor2ValFlt32(b *testing.B) {
	buf := make([]byte, 64)
	config := NewDefaultConfig()
	n := value2cbor(float32(10.2), buf, config)
	for i := 0; i < b.N; i++ {
		cbor2value(buf[:n], config)
	}
}

func BenchmarkVal2CborFlt64(b *testing.B) {
	buf := make([]byte, 64)
	config := NewDefaultConfig()
	val := interface{}(float64(10.2))
	for i := 0; i < b.N; i++ {
		value2cbor(val, buf, config)
	}
}

func BenchmarkCbor2ValFlt64(b *testing.B) {
	buf := make([]byte, 64)
	config := NewDefaultConfig()
	n := value2cbor(float64(10.2), buf, config)
	for i := 0; i < b.N; i++ {
		cbor2value(buf[:n], config)
	}
}

func BenchmarkVal2CborTBytes(b *testing.B) {
	buf := make([]byte, 64)
	config := NewDefaultConfig()
	val := interface{}([]byte("hello world"))
	for i := 0; i < b.N; i++ {
		value2cbor(val, buf, config)
	}
}

func BenchmarkCbor2ValBytes(b *testing.B) {
	buf := make([]byte, 64)
	config := NewDefaultConfig()
	n := value2cbor([]byte("hello world"), buf, config)
	for i := 0; i < b.N; i++ {
		cbor2value(buf[:n], config)
	}
}

func BenchmarkVal2CborText(b *testing.B) {
	buf := make([]byte, 64)
	config := NewDefaultConfig()
	val := interface{}("hello world")
	for i := 0; i < b.N; i++ {
		value2cbor(val, buf, config)
	}
}

func BenchmarkCbor2ValText(b *testing.B) {
	buf := make([]byte, 64)
	config := NewDefaultConfig()
	n := value2cbor("hello world", buf, config)
	for i := 0; i < b.N; i++ {
		cbor2value(buf[:n], config)
	}
}

func BenchmarkVal2CborArr0(b *testing.B) {
	buf, arr := make([]byte, 1024), make([]interface{}, 0)
	config := NewDefaultConfig()
	val := interface{}(arr)
	for i := 0; i < b.N; i++ {
		value2cbor(val, buf, config)
	}
}

func BenchmarkCbor2ValArr0(b *testing.B) {
	buf, arr := make([]byte, 1024), make([]interface{}, 0)
	config := NewDefaultConfig()
	n := value2cbor(arr, buf, config)
	for i := 0; i < b.N; i++ {
		cbor2value(buf[:n], config)
	}
}

func BenchmarkVal2CborArr5(b *testing.B) {
	buf := make([]byte, 1024)
	config := NewDefaultConfig()
	arr := interface{}([]interface{}{5, 5.0, "hello world", true, nil})
	for i := 0; i < b.N; i++ {
		value2cbor(arr, buf, config)
	}
}

func BenchmarkCbor2ValArr5(b *testing.B) {
	buf := make([]byte, 1024)
	config := NewDefaultConfig()
	arr := []interface{}{5, 5.0, "hello world", true, nil}
	n := value2cbor(arr, buf, config)
	for i := 0; i < b.N; i++ {
		cbor2value(buf[:n], config)
	}
}

func BenchmarkVal2CborMap0(b *testing.B) {
	buf := make([]byte, 1024)
	config := NewDefaultConfig()
	m := interface{}(make([][2]interface{}, 0))
	for i := 0; i < b.N; i++ {
		value2cbor(m, buf, config)
	}
}

func BenchmarkCbor2ValMap0(b *testing.B) {
	buf := make([]byte, 1024)
	config := NewDefaultConfig()
	m := make([][2]interface{}, 0)
	n := value2cbor(m, buf, config)
	for i := 0; i < b.N; i++ {
		cbor2value(buf[:n], config)
	}
}

func BenchmarkVal2CborMap5(b *testing.B) {
	buf := make([]byte, 1024)
	config := NewDefaultConfig()
	m := interface{}([][2]interface{}{
		[2]interface{}{"key0", 5}, [2]interface{}{"key1", 5.0},
		[2]interface{}{"key2", "hello world"},
		[2]interface{}{"key3", true}, [2]interface{}{"key4", nil},
	})
	for i := 0; i < b.N; i++ {
		value2cbor(m, buf, config)
	}
}

func BenchmarkCbor2ValMap5(b *testing.B) {
	buf := make([]byte, 1024)
	config := NewDefaultConfig()
	m := [][2]interface{}{
		[2]interface{}{"key0", 5}, [2]interface{}{"key1", 5.0},
		[2]interface{}{"key2", "hello world"},
		[2]interface{}{"key3", true}, [2]interface{}{"key4", nil},
	}
	n := value2cbor(m, buf, config)
	for i := 0; i < b.N; i++ {
		cbor2value(buf[:n], config)
	}
}

func BenchmarkVal2CborTyp(b *testing.B) {
	config := NewDefaultConfig()
	jsn := config.NewJson(testdataFile("testdata/typical.json"), -1)
	_, value := jsn.Tovalue()

	buf := make([]byte, 10*1024)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		value2cbor(value, buf, config)
	}
}

func BenchmarkCbor2ValTyp(b *testing.B) {
	config := NewDefaultConfig()
	jsn := config.NewJson(testdataFile("testdata/typical.json"), -1)
	cbr := config.NewCbor(make([]byte, 10*1024), 0)
	jsn.Tocbor(cbr)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cbor2value(cbr.Bytes(), config)
	}
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
