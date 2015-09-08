package gson

import "testing"
import "reflect"
import "fmt"

var _ = fmt.Sprintf("dummy")

func TestConfig(t *testing.T) {
	config := NewDefaultConfig()
	ref, buf := 10.2, make([]byte, 16)
	n := config.ValueToCbor(ref, buf)
	val, m := config.CborToValue(buf[:n])
	if n != m {
		t.Errorf("expected %v got %v", n, m)
	} else if !reflect.DeepEqual(ref, val) {
		t.Errorf("expected %v got %v", ref, val)
	}
}

func TestCborSmallInt(t *testing.T) {
	buf := make([]byte, 10)
	config := NewDefaultConfig()

	for i := int8(-24); i < 24; i++ { // SmallInt is -24..23
		if n := config.SmallintToCbor(i, buf); n != 1 {
			t.Errorf("fail Encode SmallInt: %v want 1", n)
		} else if item, m := config.CborToValue(buf); m != 1 {
			t.Errorf("fail decode on SmallInt len: %v want 1", m)
		} else if val1, ok := item.(int64); ok && val1 != int64(i) {
			t.Errorf("fail decode on SmallInt: %x, want %x", val1, i)
		} else if val2, ok := item.(uint64); ok && val2 != uint64(i) {
			t.Errorf("fail decode on SmallInt: %x, want %x", val2, i)
		}
	}
}

func TestCborSimpleType(t *testing.T) {
	buf := make([]byte, 10)
	config := NewDefaultConfig()

	// test encoding type7/simpletype < 20
	for i := 0; i < 20; i++ {
		if n := config.SimpletypeToCbor(byte(i), buf); n != 1 {
			t.Errorf("fail Encode simple-type: %v want 1", n)
		} else if item, m := config.CborToValue(buf); m != 1 {
			t.Errorf("fail decode on simple-type: %v want 1", m)
		} else if item.(byte) != byte(i) {
			//t.Logf("buf: %x", buf[0])
			t.Errorf("fail decode on simple-type: %v want %v", item, i)
		}

	}

	// test decoding typ7/simpletype extended byte
	for i := 32; i < 255; i++ {
		n := config.SimpletypeToCbor(byte(i), buf)
		item, m := config.CborToValue(buf)
		if n != m || item.(byte) != byte(i) {
			t.Errorf("fail codec simpletype extended: %v %v %v", n, m, item)
		}
	}
}

type testLocal byte

func TestUndefined(t *testing.T) {
	config := NewDefaultConfig()
	ref, buf := CborUndefined(cborSimpleUndefined), make([]byte, 16)
	n := config.ValueToCbor(ref, buf)
	val, m := config.CborToValue(buf[:n])
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
		config.ValueToCbor(testLocal(10), buf)
	}()
}

func TestIndefinite(t *testing.T) {
	buf := make([]byte, 16)
	config := NewDefaultConfig()

	bytesStart(buf)
	if config.IsIndefiniteBytes(CborIndefinite(buf[0])) == false {
		t.Errorf("failed")
	}

	textStart(buf)
	if config.IsIndefiniteText(CborIndefinite(buf[0])) == false {
		t.Errorf("failed")
	}

	arrayStart(buf)
	if config.IsIndefiniteArray(CborIndefinite(buf[0])) == false {
		t.Errorf("failed")
	}

	mapStart(buf)
	if config.IsIndefiniteMap(CborIndefinite(buf[0])) == false {
		t.Errorf("failed")
	}
}
