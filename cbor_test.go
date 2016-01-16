//  Copyright (c) 2015 Couchbase, Inc.

// +build ignore

package gson

import "fmt"
import "testing"

var _ = fmt.Sprintf("dummy")

func TestCborMajor(t *testing.T) {
	if typ := cborMajor(0xff); typ != 0xe0 {
		t.Errorf("fail major() got %v wanted 0xe0", typ)
	}
}

func TestCborSmallInt(t *testing.T) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(make([]byte, 10), 0)

	for i := int8(-24); i < 24; i++ { // SmallInt is -24..23
		cbr.EncodeSmallint(i)
		if item, m := cbr.Tovalue(); m != 1 {
			t.Errorf("fail decode on SmallInt len: %v want 1", m)
		} else if val1, ok := item.(int64); ok && val1 != int64(i) {
			t.Errorf("fail decode on SmallInt: %x, want %x", val1, i)
		} else if val2, ok := item.(uint64); ok && val2 != uint64(i) {
			t.Errorf("fail decode on SmallInt: %x, want %x", val2, i)
		}
		cbr.Reset()
	}
}

func TestCborSimpleType(t *testing.T) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(make([]byte, 10), 0)

	// test encoding type7/simpletype < 20
	for i := 0; i < 20; i++ {
		cbr.EncodeSimpletype(byte(i))
		if item, m := cbr.Tovalue(); m != 1 {
			t.Errorf("fail decode on simple-type: %v want 1", m)
		} else if item.(byte) != byte(i) {
			t.Errorf("fail decode on simple-type: %v want %v", item, i)
		}
		cbr.Reset()
	}

	// test decoding typ7/simpletype extended byte
	for i := 32; i < 255; i++ {
		cbr.EncodeSimpletype(byte(i))
		if item, _ := cbr.Tovalue(); item.(byte) != byte(i) {
			t.Errorf("fail codec simpletype extended: %v", item)
		}
		cbr.Reset()
	}
}

func TestIndefinite(t *testing.T) {
	buf := make([]byte, 16)
	config := NewDefaultConfig()
	cbr := config.NewCbor(buf, 0)

	cbr.n = bytesStart(cbr.data)
	if cbr.IsIndefiniteBytes() == false {
		t.Errorf("failed")
	}
	cbr.Reset()

	cbr.n = textStart(cbr.data)
	if cbr.IsIndefiniteText() == false {
		t.Errorf("failed")
	}
	cbr.Reset()

	cbr.n = arrayStart(cbr.data)
	if cbr.IsIndefiniteArray() == false {
		t.Errorf("failed")
	}
	cbr.Reset()

	cbr.n = mapStart(cbr.data)
	if cbr.IsIndefiniteMap() == false {
		t.Errorf("failed")
	}
	cbr.Reset()
}
