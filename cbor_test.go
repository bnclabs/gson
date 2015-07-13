package gson

import "testing"

func TestCborSmallInt(t *testing.T) {
	buf := make([]byte, 10)

	for i := int8(-24); i < 24; i++ { // SmallInt is -24..23
		if n := EncodeSmallInt(i, buf); n != 1 {
			t.Errorf("fail Encode SmallInt: %v want 1", n)
		} else if item, m := Decode(buf); m != 1 {
			t.Errorf("fail Decode on SmallInt len: %v want 1", m)
		} else if item.(int64) != int64(i) {
			t.Errorf("fail Decode on SmallInt: %x, want %x", item.(int64), i)
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
			t.Logf("buf: %x", buf[0])
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
