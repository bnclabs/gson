//  Copyright (c) 2015 Couchbase, Inc.

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
		item := cbr.Tovalue()
		if val1, ok := item.(int64); ok && val1 != int64(i) {
			t.Errorf("fail decode on SmallInt: %x, want %x", val1, i)
		} else if val2, ok := item.(uint64); ok && val2 != uint64(i) {
			t.Errorf("fail decode on SmallInt: %x, want %x", val2, i)
		}
		cbr.Reset(nil)
	}
}

func TestCborSimpleType(t *testing.T) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(make([]byte, 10), 0)

	// test encoding type7/simpletype < 20
	for i := 0; i < 20; i++ {
		cbr.EncodeSimpletype(byte(i))
		item := cbr.Tovalue()
		if item.(byte) != byte(i) {
			t.Errorf("fail decode on simple-type: %v want %v", item, i)
		}
		cbr.Reset(nil)
	}

	// test decoding typ7/simpletype extended byte
	for i := 32; i < 255; i++ {
		cbr.EncodeSimpletype(byte(i))
		if item := cbr.Tovalue(); item.(byte) != byte(i) {
			t.Errorf("fail codec simpletype extended: %v", item)
		}
		cbr.Reset(nil)
	}
}
