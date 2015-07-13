package gson

import "testing"
import "fmt"

var _ = fmt.Sprintln("dummy")

func TestCborSmallInt(t *testing.T) {
	buf := make([]byte, 10)

	for i := int8(-24); i < 24; i++ { // SmallInt is -24..23
		if n := EncodeSmallInt(i, buf[:]); n != 1 {
			t.Errorf("fail EncodeSmall: %v want 1", n)
		} else if item, m := Decode(buf); n != 1 {
			t.Errorf("fail Decode on SmallInt len: %v want 1", m)
		} else if item.(int64) != int64(i) {
			t.Errorf("fail Decode on SmallInt: %x, want %x", item.(int64), i)
		}
	}
}
