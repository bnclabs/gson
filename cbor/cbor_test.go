package cbor

import "testing"
import "reflect"
import "fmt"

var _ = fmt.Sprintf("dummy")

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

	// test decoding typ7/simpletype extended byte
	for i := 32; i < 255; i++ {
		n := EncodeSimpleType(byte(i), buf)
		item, m := Decode(buf)
		if n != m || item.(byte) != byte(i) {
			t.Errorf("fail codec simpletype extended: %v %v %v", n, m, item)
		}
	}
}

type testLocal byte

func TestUndefined(t *testing.T) {
	ref, buf := Undefined(simpleUndefined), make([]byte, 16)
	n := Encode(ref, buf)
	val, m := Decode(buf[:n])
	if n != m {
		t.Errorf("expected %v got %v", n, m)
	} else if !reflect.DeepEqual(ref, val) {
		t.Errorf("expected %v got %v", ref, val)
	}
	// test unknown type.
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		Encode(testLocal(10), buf)
	}()
}

func TestIndefinite(t *testing.T) {
	buf := make([]byte, 16)

	encodeBytesStart(buf)
	if IsIndefiniteBytes(Indefinite(buf[0])) == false {
		t.Errorf("failed")
	}

	encodeTextStart(buf)
	if IsIndefiniteText(Indefinite(buf[0])) == false {
		t.Errorf("failed")
	}

	encodeArrayStart(buf)
	if IsIndefiniteArray(Indefinite(buf[0])) == false {
		t.Errorf("failed")
	}

	encodeMapStart(buf)
	if IsIndefiniteMap(Indefinite(buf[0])) == false {
		t.Errorf("failed")
	}
}
