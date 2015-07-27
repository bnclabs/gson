package cbor

import "testing"
import "reflect"
import "time"
import "fmt"
import "regexp"
import "math/big"

var _ = fmt.Sprintf("dummy")

func TestDateTime(t *testing.T) {
	buf := make([]byte, 64)

	ref, err := time.Parse(time.RFC3339, "2006-01-02T15:04:05+07:00")
	if err != nil {
		t.Errorf("time.Parse() failed: %v", err)
	}

	n := Encode(ref, buf)
	item, m := Decode(buf[:n])
	if n != 28 || n != m {
		t.Errorf("expected %v got %v %v", 28, n, m)
	}
	if !ref.Equal(item.(time.Time)) {
		t.Errorf("expected %v got %v", ref, item.(time.Time))
	}

	// malformed.
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		buf[5] = 'a'
		Decode(buf[:n])
	}()
}

func TestEpoch(t *testing.T) {
	buf := make([]byte, 64)

	// positive and negative epoch
	for _, val := range [2]int64{1000000, -100000} {
		ref := Epoch(val)
		n := Encode(ref, buf)
		item, m := Decode(buf[:n])
		if n != 6 || n != m {
			t.Errorf("expected %v got %v %v", 6, n, m)
		}
		if !reflect.DeepEqual(ref, item) {
			t.Errorf("expected %v got %v", ref, item)
		}
	}
	// malformed epoch
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		buf[1] = 0x5a // instead of 0x3a
		Decode(buf)
	}()
	// malformed epoch
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		buf := make([]byte, 16)
		n := encodeTag(tagEpoch, buf)
		n += encodeBytes([]byte{1, 2}, buf[n:])
		Decode(buf)
	}()
}

func TestEpochMicro(t *testing.T) {
	buf := make([]byte, 64)
	// positive and negative epoch in uS.
	for _, val := range [2]float64{1000000.123456, -100000.123456} {
		ref := EpochMicro(val)
		n := Encode(ref, buf)
		item, m := Decode(buf[:n])
		if n != 10 || n != m {
			t.Errorf("expected %v got %v %v", 10, n, m)
		}
		if !reflect.DeepEqual(ref, item) {
			t.Errorf("expected %v got %v", ref, item)
		}
	}
}

func TestBigNum(t *testing.T) {
	buf := make([]byte, 64)
	// positive and negative bignums
	for _, val := range [2]int64{1000, -1000} {
		bigx := big.NewInt(9223372036854775807)
		bigy := big.NewInt(val)
		bigz := big.NewInt(0).Mul(bigx, bigy)
		n := Encode(bigz, buf)
		item, m := Decode(buf[:n])
		if n != 12 || n != m {
			t.Errorf("expected %v got %v %v", 12, n, m)
		}
		if bigz.Cmp(item.(*big.Int)) != 0 {
			t.Errorf("expected %v got %v", bigz, item.(*big.Int))
		}
	}
}

func TestDecimalFraction(t *testing.T) {
	buf := make([]byte, 64)
	// for positive
	refs := []DecimalFraction{
		DecimalFraction([2]interface{}{int64(-10), int64(-23)}),
		DecimalFraction([2]interface{}{int64(-10), int64(23)}),
		DecimalFraction([2]interface{}{int64(10), int64(-23)}),
		DecimalFraction([2]interface{}{int64(10), int64(23)}),
	}
	for _, ref := range refs {
		n := Encode(ref, buf)
		item, m := Decode(buf[:n])
		if n != 3 || n != m {
			t.Errorf("expected %v got %v %v", 3, n, m)
		}
		if !reflect.DeepEqual(ref, item) {
			t.Errorf("expected %v got %v", ref, item)
		}
	}
}

func TestBigFloat(t *testing.T) {
	buf := make([]byte, 64)
	refs := []BigFloat{
		BigFloat([2]interface{}{int64(-10), int64(-23)}),
		BigFloat([2]interface{}{int64(-10), int64(23)}),
		BigFloat([2]interface{}{int64(10), int64(-23)}),
		BigFloat([2]interface{}{int64(10), int64(23)}),
	}
	for _, ref := range refs {
		n := Encode(ref, buf)
		item, m := Decode(buf[:n])
		if n != 3 || n != m {
			t.Errorf("expected %v got %v %v", 3, n, m)
		}
		if !reflect.DeepEqual(ref, item) {
			t.Errorf("expected %v got %v", ref, item)
		}
	}
}

func TestCbor(t *testing.T) {
	buf := make([]byte, 64)
	ref := Cbor([]byte("hello world"))
	n := Encode(ref, buf)
	item, m := Decode(buf[:n])
	if n != 14 || n != m {
		t.Errorf("expected %v got %v %v", 14, n, m)
	}
	if !reflect.DeepEqual(ref, item) {
		t.Errorf("exptected %v got %v", ref, item)
	}
}

func TestRegexp(t *testing.T) {
	buf := make([]byte, 64)
	ref, err := regexp.Compile(`a([0-9]t*)+`)
	if err != nil {
		t.Errorf("compiling regex")
	}
	n := Encode(ref, buf)
	item, m := Decode(buf[:n])
	if n != 14 || n != m {
		t.Errorf("expected %v got %v %v", 14, n, m)
	}
	if ref.String() != (item.(*regexp.Regexp)).String() {
		t.Errorf("expected %v got %v", ref, item)
	}
	// malformed reg-ex
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		n := encodeTag(tagRegexp, buf)
		n += encodeText(`a([0-9]t*+`, buf[n:])
		Decode(buf)
	}()
}

func TestCborPrefix(t *testing.T) {
	buf := make([]byte, 64)
	ref := CborPrefix([]byte("hello world"))
	n := Encode(ref, buf)
	item, m := Decode(buf[:n])
	if n != 15 || n != m {
		t.Errorf("expected %v got %v %v", 15, n, m)
	}
	if !reflect.DeepEqual(ref, item) {
		t.Errorf("exptected %v got %v", ref, item)
	}
}
