package gson

import "testing"

func TestEncodeSmall(t *testing.T) {
	var buf [10]byte

	if EncodeSmall(byte(0), buf) != 1 {
		t.Errorf("fail EncodeSmall return expected as 1")
	} else if buf[0] != 0x00 {
		t.Errorf("fail EncodeSmall expected 0x00")
	}

	if EncodeSmall(byte(0), buf) != 1 {
		t.Errorf("fail EncodeSmall return expected as 1")
	} else if buf[0] != 0x17 {
		t.Errorf("fail EncodeSmall expected 0x17")
	}
}
