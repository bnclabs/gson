//  Copyright (c) 2015 Couchbase, Inc.

package gson

import "testing"
import "fmt"
import "reflect"

var _ = fmt.Sprintf("dummy")

func TestCborClone(t *testing.T) {
	testcases := [][2]interface{}{
		[2]interface{}{nil, nil},
		[2]interface{}{true, true},
		[2]interface{}{false, false},
		[2]interface{}{uint8(1), uint64(1)},
		[2]interface{}{int8(-1), int64(-1)},
		[2]interface{}{uint8(100), uint64(100)},
		[2]interface{}{int8(-100), int64(-100)},
		[2]interface{}{uint16(1024), uint64(1024)},
		[2]interface{}{int16(-1024), int64(-1024)},
		[2]interface{}{uint32(1048576), uint64(1048576)},
		[2]interface{}{int32(-1048576), int64(-1048576)},
		[2]interface{}{uint64(1099511627776), uint64(1099511627776)},
		[2]interface{}{int64(-1099511627776), int64(-1099511627776)},
		[2]interface{}{float32(10.2), float32(10.2)},
		[2]interface{}{float64(10.2), float64(10.2)},
		[2]interface{}{[]byte("hello world"), []byte("hello world")},
		[2]interface{}{"hello world", "hello world"},
		[2]interface{}{
			[]interface{}{12.0, nil, true, false, "hello world"},
			[]interface{}{12.0, nil, true, false, "hello world"},
		},
		[2]interface{}{
			map[string]interface{}{
				"a": 12.0, "b": nil, "c": true, "d": false, "e": "hello world"},
			map[string]interface{}{
				"a": 12.0, "b": nil, "c": true, "d": false, "e": "hello world"},
		},
	}

	config := NewDefaultConfig()
	cbr := config.NewCbor(make([]byte, 1024), 0)
	cloned := config.NewCbor(make([]byte, 1024), 0)
	out := make([]byte, 1024)

	for _, tcase := range testcases {
		t.Logf("%v", tcase[0])

		config.NewValue(tcase[0]).Tocbor(cbr.Reset(nil))
		in := cbr.Bytes()
		if n := cborclone(in, out, config); len(in) != n {
			t.Errorf("expected %v, got %v", len(in), n)
		}
		value := cloned.Reset(out).Tovalue()
		if !reflect.DeepEqual(value, tcase[1]) {
			t.Errorf("expected %v, got %v", tcase[1], value)
		}
	}

	// test CborUndefined
	config.NewValue(CborUndefined(1)).Tocbor(cbr.Reset(nil))
	in := cbr.Bytes()
	n := cborclone(in, out, config)
	if len(in) != n {
		t.Errorf("expected %v, got %v", len(in), n)
	}
	value := cloned.Reset(out).Tovalue()
	if _, ok := value.(CborUndefined); !ok {
		t.Errorf("expected CborUndefined, got %T", value)
	}
}
