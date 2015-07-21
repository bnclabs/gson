package cbor

import "testing"
import "reflect"
import "time"
import "fmt"
import "regexp"
import "math/big"

var _ = fmt.Sprintf("dummy")

func TestCborMajor(t *testing.T) {
	if typ := major(0xff); typ != 0xe0 {
		t.Errorf("fail major() got %v wanted 0xe0", typ)
	}
}

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
	} else if item.(Undefined) != Undefined(simpleUndefined) {
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
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		Decode([]byte{0xf8, 31})
	}()

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
		n := Encode(test[0], buf)
		val, m := Decode(buf)
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
		Encode(uint64(9223372036854775808), buf)
		buf[0] = (buf[0] & 0x1f) | type1 // fix as negative integer
		Decode(buf)
	}()
}

func TestCborFloat16(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic while decoding float16")
		}
	}()
	Decode([]byte{0xf9, 0, 0, 0, 0})
}

func TestCborFloat32(t *testing.T) {
	buf, ref := make([]byte, 10), float32(10.11)
	n := Encode(ref, buf)
	t.Logf("%v", buf)
	val, m := Decode(buf)
	if n != 5 || n != m || !reflect.DeepEqual(val, ref) {
		t.Errorf("fail code float32: %v %v %T(%v)", n, m, val, val)
	}
}

func TestCborFloat64(t *testing.T) {
	buf, ref := make([]byte, 10), float64(10.11)
	n := Encode(ref, buf)
	val, m := Decode(buf)
	if n != 9 || n != m || !reflect.DeepEqual(val, ref) {
		t.Errorf("fail code float32: %v %v %T(%v)", n, m, val, val)
	}
}

func TestCborBytes(t *testing.T) {
	buf, ref := make([]byte, 200), make([]uint8, 100)
	for i := 0; i < len(ref); i++ {
		ref[i] = uint8(i)
	}
	n := Encode(ref, buf)
	val, m := Decode(buf)
	if n != 102 || n != m || !reflect.DeepEqual(val, ref) {
		t.Errorf("fail code bytes: %v %v %T(%v)", n, m, val, val)
	}
	// test byte-start
	if n := encodeBytesStart(buf); n != 1 {
		t.Errorf("fail code bytes-start len: %v wanted 1", n)
	} else if val, m := Decode(buf); m != n {
		t.Errorf("fail code bytes-start size: %v wanted %v", m, n)
	} else if !reflect.DeepEqual(val, Indefinite(0x5f)) {
		t.Errorf("fail code bytes-start: %v wanted 0x5f", buf[0])
	}
}

func TestCborText(t *testing.T) {
	buf, ref := make([]byte, 200), "hello world"
	n := Encode(ref, buf)
	val, m := Decode(buf)
	if n != 12 || n != m || !reflect.DeepEqual(val, ref) {
		t.Errorf("fail code text: %v %v %T(%v)", n, m, val, val)
	}

	// test text-start
	if n := encodeTextStart(buf); n != 1 {
		t.Errorf("fail code text-start len: %v wanted 1", n)
	} else if val, m := Decode(buf); m != n {
		t.Errorf("fail code text-start size: %v wanted %v", m, n)
	} else if !reflect.DeepEqual(val, Indefinite(0x7f)) {
		t.Errorf("fail code text-start: %x wanted 0x7f", buf[0])
	}
}

//func TestCborIndefinite(t *testing.T) {
//    defer func() {
//        if r := recover(); r == nil {
//            t.Errorf("expected panic while encoding rune")
//        }
//    }()
//
//    buf := make([]byte, 10)
//    Encode(Indefinite(10), buf) // cannot encode rune
//}

func TestCborArray(t *testing.T) {
	buf := make([]byte, 1024)
	ref := []interface{}{10.2, "hello world"}
	n := Encode(ref, buf)
	val, m := Decode(buf)
	if n != 22 || n != m || !reflect.DeepEqual(ref, val) {
		t.Errorf("fail code text: %v %v %T(%v)", n, m, val, val)
	}
	// test text-start
	if n := encodeArrayStart(buf); n != 1 {
		t.Errorf("fail code text-start len: %v wanted 1", n)
	} else if val, m := Decode(buf); m != n {
		t.Errorf("fail code text-start size : %v wanted %v", m, n)
	} else if !reflect.DeepEqual(val, Indefinite(0x9f)) {
		t.Errorf("fail code text-start: %x wanted 0x9f", buf[0])
	}
}

func TestCborMap(t *testing.T) {
	buf := make([]byte, 1024)
	ref := [][2]interface{}{
		[2]interface{}{10.2, "hello world"},
		[2]interface{}{"hello world", 10.2},
	}
	n := Encode(ref, buf)
	val, m := Decode(buf)
	if n != 43 || n != m || !reflect.DeepEqual(ref, val) {
		t.Errorf("fail code text: %v %v %T(%v)", n, m, val, val)
	}
	// test text-start
	if n := encodeMapStart(buf); n != 1 {
		t.Errorf("fail code text-start len: %v wanted 1", n)
	} else if val, m := Decode(buf); m != n {
		t.Errorf("fail code text-start size : %v wanted %v", m, n)
	} else if !reflect.DeepEqual(val, Indefinite(0xbf)) {
		t.Errorf("fail code text-start: %x wanted 0xbf", buf[0])
	}
}

func TestCborBreakStop(t *testing.T) {
	buf := make([]byte, 10)
	if n := encodeBreakStop(buf); n != 1 {
		t.Errorf("fail code text-start len: %v wanted 1", n)
	} else if val, m := Decode(buf); m != n {
		t.Errorf("fail code text-start: %v wanted %v", m, n)
	} else if !reflect.DeepEqual(val, BreakStop(0xff)) {
		t.Errorf("fail code text-start: %x wanted 0xff", buf[0])
	}
}

func TestCborReserved(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic while decoding reserved")
		}
	}()
	Decode([]byte{hdr(type0, 28)})
}

func TestDateTime(t *testing.T) {
	buf := make([]byte, 64)

	ref, err := time.Parse(time.RFC3339, "2006-01-02T15:04:05+07:00")
	if err != nil {
		t.Errorf("time.Parse() failed: %v", err)
	}

	n := Encode(ref, buf)
	item, m := Decode(buf[:n])
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
		Decode(buf[:n])
	}()
}

func TestEpoch(t *testing.T) {
	buf := make([]byte, 64)

	// positive and negative epoch
	for _, val := range [2]int64{1000000, -100000} {
		ref := Epoch(val)
		n := Encode(ref, buf)
		item, m := Decode(buf[:n])
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
		Decode(buf)
	}()
}

func TestEpochMicro(t *testing.T) {
	buf := make([]byte, 64)
	// positive and negative epoch in uS.
	for _, val := range [2]float64{1000000.123456, -100000.123456} {
		ref := EpochMicro(val)
		n := Encode(ref, buf)
		item, m := Decode(buf[:n])
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
	// positive and negative bignums
	for _, val := range [2]int64{1000, -1000} {
		bigx := big.NewInt(9223372036854775807)
		bigy := big.NewInt(val)
		bigz := big.NewInt(0).Mul(bigx, bigy)
		n := Encode(bigz, buf)
		item, m := Decode(buf[:n])
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
	// for positive
	refs := []DecimalFraction{
		DecimalFraction([2]interface{}{int64(-10), int64(-23)}),
		DecimalFraction([2]interface{}{int64(-10), int64(23)}),
		DecimalFraction([2]interface{}{int64(10), int64(-23)}),
		DecimalFraction([2]interface{}{int64(10), int64(23)}),
	}
	for _, ref := range refs {
		n := Encode(ref, buf)
		item, m := Decode(buf[:n])
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
	refs := []BigFloat{
		BigFloat([2]interface{}{int64(-10), int64(-23)}),
		BigFloat([2]interface{}{int64(-10), int64(23)}),
		BigFloat([2]interface{}{int64(10), int64(-23)}),
		BigFloat([2]interface{}{int64(10), int64(23)}),
	}
	for _, ref := range refs {
		n := Encode(ref, buf)
		item, m := Decode(buf[:n])
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
	ref := Cbor([]byte("hello world"))
	n := Encode(ref, buf)
	item, m := Decode(buf[:n])
	if n != 14 || n != m {
		t.Errorf("expected %v got %v %v", 14, n, m)
	}
	if !reflect.DeepEqual(ref, item) {
		t.Errorf("exptected %v got %v", ref, item)
	}
}

func TestRegexp(t *testing.T) {
	buf := make([]byte, 64)
	ref, err := regexp.Compile(`a([0-9]t*)+`)
	if err != nil {
		t.Errorf("compiling regex")
	}
	n := Encode(ref, buf)
	item, m := Decode(buf[:n])
	if n != 14 || n != m {
		t.Errorf("expected %v got %v %v", 14, n, m)
	}
	if ref.String() != (item.(*regexp.Regexp)).String() {
		t.Errorf("expected %v got %v", ref, item)
	}
}

func TestCborPrefix(t *testing.T) {
	buf := make([]byte, 64)
	ref := CborPrefix([]byte("hello world"))
	n := Encode(ref, buf)
	item, m := Decode(buf[:n])
	if n != 15 || n != m {
		t.Errorf("expected %v got %v %v", 15, n, m)
	}
	if !reflect.DeepEqual(ref, item) {
		t.Errorf("exptected %v got %v", ref, item)
	}
}
