package gson

import "testing"
import "reflect"

func TestCborSmallInt(t *testing.T) {
	buf := make([]byte, 10)

	for i := int8(-24); i < 24; i++ { // SmallInt is -24..23
		if n := EncodeSmallInt(i, buf); n != 1 {
			t.Errorf("fail Encode SmallInt: %v want 1", n)
		} else if item, m := Decode(buf); m != 1 {
			t.Errorf("fail Decode on SmallInt len: %v want 1", m)
		} else if val1, ok := item.(int64); ok && val1 != int64(i) {
			t.Errorf("fail Decode on SmallInt: %x, want %x", val1, i)
		} else if val2, ok := item.(uint64); ok && val2 != uint64(i) {
			t.Errorf("fail Decode on SmallInt: %x, want %x", val2, i)
		}
	}
}

func TestCborNil(t *testing.T) {
	buf := make([]byte, 10)

	if n := Encode(nil, buf); n != 1 {
		t.Errorf("fail Encode nil: %v want 1", n)
	} else if item, m := Decode(buf); m != 1 {
		t.Errorf("fail Decode on nil len: %v want 1", m)
	} else if item != nil {
		t.Errorf("fail Decode on nil: %x", item)
	}
}

func TestCborTrue(t *testing.T) {
	buf := make([]byte, 10)

	if n := Encode(true, buf); n != 1 {
		t.Errorf("fail Encode true: %v want 1", n)
	} else if item, m := Decode(buf); m != 1 {
		t.Errorf("fail Decode on true len: %v want 1", m)
	} else if item.(bool) != true {
		t.Errorf("fail Decode on true: %v", item)
	}
}

func TestCborFalse(t *testing.T) {
	buf := make([]byte, 10)

	if n := Encode(false, buf); n != 1 {
		t.Errorf("fail Encode false: %v want 1", n)
	} else if item, m := Decode(buf); m != 1 {
		t.Errorf("fail Decode on false len: %v want 1", m)
	} else if item.(bool) != false {
		t.Errorf("fail Decode on false: %v", item)
	}
}

func TestCborUndefined(t *testing.T) {
	buf := make([]byte, 10)

	if n := EncodeUndefined(buf); n != 1 {
		t.Errorf("fail Encode undefined: %v want 1", n)
	} else if item, m := Decode(buf); m != 1 {
		t.Errorf("fail Decode on undefined len: %v want 1", m)
	} else if item.(Undefined) != Undefined(SimpleUndefined) {
		t.Errorf("fail Decode on undefined: %T %v", item, item)
	}
}

func TestCborSimpleType(t *testing.T) {
	buf := make([]byte, 10)

	// test encoding type7/simpletype < 20
	for i := 0; i < 20; i++ {
		if n := EncodeSimpleType(byte(i), buf); n != 1 {
			t.Errorf("fail Encode simple-type: %v want 1", n)
		} else if item, m := Decode(buf); m != 1 {
			t.Errorf("fail Decode on simple-type: %v want 1", m)
		} else if item.(byte) != byte(i) {
			//t.Logf("buf: %x", buf[0])
			t.Errorf("fail Decode on simple-type: %v want %v", item, i)
		}

	}

	// test encodint type7/simpletype reserved
	for i := 20; i < 32; i++ {
		func() {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("expected panic")
				}
			}()
			EncodeSimpleType(byte(i), buf)
		}()
	}

	// test decoding typ7/simpletype extended byte
	for i := 32; i < 255; i++ {
		n := EncodeSimpleType(byte(i), buf)
		item, m := Decode(buf)
		if n != m || item.(byte) != byte(i) {
			t.Errorf("fail codec simpletype extended: %v %v %v", n, m, item)
		}
	}
}
func TestCborUint8(t *testing.T) {
	buf := make([]byte, 20)
	for i := uint16(0); i <= 255; i++ {
		n := encodeUint8(uint8(i), buf)
		val, m := Decode(buf)
		if n != m || val.(uint64) != uint64(i) {
			t.Errorf("fail codec uint8: %v %v %v", n, m, val)
		}
	}
}

func TestCborInt8(t *testing.T) {
	buf := make([]byte, 20)
	for i := int16(-128); i <= 127; i++ {
		n := encodeInt8(int8(i), buf)
		val, m := Decode(buf)
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
		[2]interface{}{int32(2147483647), uint64(2147483647)},
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
		[2]interface{}{int64(-9223372036854775807), int64(-9223372036854775807)},
	}
	for _, test := range tests {
		n := Encode(test[0], buf)
		val, m := Decode(buf)
		//t.Logf("executing test case %v", test)
		if n != m || !reflect.DeepEqual(val, test[1]) {
			t.Errorf(
				"fail codec Num: %v %v %T(%v) %T(%v)",
				n, m, val, val, test[1], test[1])
		}
	}
}
