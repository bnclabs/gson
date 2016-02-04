//  Copyright (c) 2015 Couchbase, Inc.

package gson

import "testing"
import "fmt"
import "reflect"

var _ = fmt.Sprintf("dummy")

func TestCollateReset(t *testing.T) {
	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 1024), 0)
	cltr := config.NewCollate(make([]byte, 1024), 0)

	ref := []interface{}{"sound", "ok", "horn"}
	config.NewValue(ref).Tocollate(clt)
	cltr.Reset(clt.Bytes())
	if value := cltr.Tovalue(); !reflect.DeepEqual(value, ref) {
		t.Errorf("expected %v, got %v", ref, value)
	}
}

func TestCollateEmpty(t *testing.T) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(make([]byte, 10), 0)
	jsn := config.NewJson(make([]byte, 10), 0)
	clt := config.NewCollate(make([]byte, 10), 0)

	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		clt.Tovalue()
	}()
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		clt.Tojson(jsn)
	}()
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		clt.Tocbor(cbr)
	}()
}
