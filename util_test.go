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
	cborcode := make([]byte, 1024)
	config := NewDefaultConfig()
	_, value := config.JsonToValue(ref)
	n := config.ValueToCbor(GolangMap2cborMap(value), cborcode)
	value1, _ := config.CborToValue(cborcode[:n])
	data, err := json.Marshal(CborMap2golangMap(value1))
	if err != nil {
		t.Fatalf("json parsing: %v\n	%v", value1, err)
	}
	if s := string(data); s != ref {
		t.Errorf("expected %q, got %q", ref, s)
	}
}

func TestNormalizeFloat(t *testing.T) {
	config := NewDefaultConfig()
	code := make([]byte, 1024)
	// test with SmartNumber32 and SmartNumber
	panicfn := func(nk NumberKind) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		config = config.NumberKind(nk)
		config.JsonToCollate("10.3", code)
	}
	panicfn(SmartNumber32)
	panicfn(SmartNumber)
}

func TestDenormalizeFloat(t *testing.T) {
	config := NewDefaultConfig()
	code, text := make([]byte, 1024), make([]byte, 1024)
	n := config.JsonToCollate("10.3", code)
	// test with SmartNumber32 and SmartNumber
	panicfn := func(nk NumberKind) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		config = config.NumberKind(nk)
		config.CollateToCbor(code[:n], text)
	}
	panicfn(SmartNumber32)
	panicfn(SmartNumber)
}

func TestDenormalizeFloatTojson(t *testing.T) {
	config := NewDefaultConfig()
	code, text := make([]byte, 1024), make([]byte, 1024)
	n := config.JsonToCollate("10.3", code)
	// test with SmartNumber32 and SmartNumber
	panicfn := func(nk NumberKind) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		config = config.NumberKind(nk)
		config.CollateToJson(code[:n], text)
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
