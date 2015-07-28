package cbor

import "testing"
import "reflect"

func TestConfig(t *testing.T) {
	config := NewDefaultConfig()
	ref, buf := 10.2, make([]byte, 16)
	n := config.Encode(ref, buf)
	val, m := config.Decode(buf[:n])
	if n != m {
		t.Errorf("expected %v got %v", n, m)
	} else if !reflect.DeepEqual(ref, val) {
		t.Errorf("expected %v got %v", ref, val)
	}
}
