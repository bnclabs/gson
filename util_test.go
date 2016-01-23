//  Copyright (c) 2015 Couchbase, Inc.

package gson

import "testing"
import "encoding/json"
import "sort"
import "reflect"
import "fmt"

var _ = fmt.Sprintf("dummy")

func TestBytes2Str(t *testing.T) {
	if bytes2str(nil) != "" {
		t.Errorf("fail bytes2str(nil)")
	}
}

func TestStr2Bytes(t *testing.T) {
	if str2bytes("") != nil {
		t.Errorf(`fail str2bytes("")`)
	}
}

func TestCborMap2Golang(t *testing.T) {
	ref := `{"a":10,"b":[true,false,null]}`
	config := NewDefaultConfig()
	jsn := config.NewJson([]byte(ref), -1)
	cbr := config.NewCbor(make([]byte, 1024), 0)

	_, val1 := jsn.Tovalue()
	value := config.NewValue(GolangMap2cborMap(val1))
	value.Tocbor(cbr)
	val2 := cbr.Tovalue()
	data, err := json.Marshal(CborMap2golangMap(val2))
	if err != nil {
		t.Fatalf("json parsing: %v\n	%v", val2, err)
	}
	if s := string(data); s != ref {
		t.Errorf("expected %q, got %q", ref, s)
	}
}

func TestNormalizeFloat(t *testing.T) {
	// test with SmartNumber32 and SmartNumber
	panicfn := func(nk NumberKind) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		config := NewDefaultConfig().SetNumberKind(nk)
		jsn := config.NewJson([]byte("10.3"), -1)
		clt := config.NewCollate(make([]byte, 1024), 0)
		jsn.Tocollate(clt)
	}
	panicfn(SmartNumber32)
	panicfn(SmartNumber)
}

func TestDenormalizeFloat(t *testing.T) {
	// test with SmartNumber32 and SmartNumber
	panicfn := func(nk NumberKind) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		config := NewDefaultConfig().SetNumberKind(nk)
		jsn := config.NewJson([]byte("10.3"), -1)
		cbr := config.NewCbor(make([]byte, 1024), 0)
		clt := config.NewCollate(make([]byte, 1024), 0)
		jsn.Tocollate(clt)
		clt.Tocbor(cbr)
	}
	panicfn(SmartNumber32)
	panicfn(SmartNumber)
}

func TestDenormalizeFloatTojson(t *testing.T) {
	// test with SmartNumber32 and SmartNumber
	panicfn := func(nk NumberKind) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		config := NewDefaultConfig().SetNumberKind(nk)
		jsn := config.NewJson(make([]byte, 1024), 0)
		clt := config.NewCollate(make([]byte, 1024), 0)

		jsn.Reset([]byte("10.3"))
		jsn.Tocollate(clt)
		clt.Tojson(jsn.Reset(nil))
	}
	panicfn(SmartNumber32)
	panicfn(SmartNumber)
}

func TestKvrefs(t *testing.T) {
	items := make(kvrefs, 4)
	items[0] = kvref{"1", []byte("1")}
	items[2] = kvref{"3", []byte("3")}
	items[1] = kvref{"2", []byte("2")}
	items[3] = kvref{"0", []byte("0")}
	sort.Sort(items)
	ref := kvrefs{
		kvref{"0", []byte("0")},
		kvref{"1", []byte("1")},
		kvref{"2", []byte("2")},
		kvref{"3", []byte("3")},
	}
	if !reflect.DeepEqual(ref, items) {
		t.Errorf("expected %v, got %v", ref, items)
	}
}

func BenchmarkBytes2Str(b *testing.B) {
	bs := []byte("hello world")
	for i := 0; i < b.N; i++ {
		bytes2str(bs)
	}
}

func BenchmarkStr2Bytes(b *testing.B) {
	s := "hello world"
	for i := 0; i < b.N; i++ {
		str2bytes(s)
	}
}

func compare_jsons(t *testing.T, json1, json2 string) error {
	var m1, m2 interface{}
	err := json.Unmarshal(str2bytes(json1), &m1)
	if err != nil {
		return fmt.Errorf("parsing %v: %v", json1, err)
	}
	err = json.Unmarshal(str2bytes(json2), &m2)
	if err != nil {
		return fmt.Errorf("parsing %v: %v", json2, err)
	}
	if !reflect.DeepEqual(m1, m2) {
		return fmt.Errorf("expected %v, got %v", m1, m2)
	}
	return nil
}
