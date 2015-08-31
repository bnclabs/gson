package collate

import "reflect"
import "unsafe"
import "strconv"

func bytes2str(bytes []byte) string {
	if bytes == nil {
		return ""
	}
	sl := (*reflect.SliceHeader)(unsafe.Pointer(&bytes))
	st := &reflect.StringHeader{Data: sl.Data, Len: sl.Len}
	return *(*string)(unsafe.Pointer(st))
}

func str2bytes(str string) []byte {
	if str == "" {
		return nil
	}
	st := (*reflect.StringHeader)(unsafe.Pointer(&str))
	sl := &reflect.SliceHeader{Data: st.Data, Len: st.Len, Cap: st.Len}
	return *(*[]byte)(unsafe.Pointer(sl))
}

// numbers can be encoded as integers, or as small-decimal,
// or as floating-point - normalizeFloat() takes the number as
// float64 or int64 and based on the configuration encodes it
// integer or small-decimal or floating-point.
func normalizeFloat(value interface{}, code []byte, nt NumberKind) int {
	var num [64]byte
	switch nt {
	case Float64:
		v, ok := value.(float64)
		if !ok {
			v = float64(value.(int64))
		}
		bs := strconv.AppendFloat(num[:0], v, 'e', -1, 64)
		return encodeFloat(bs, code)

	case Int64:
		v, ok := value.(int64)
		if !ok {
			v = int64(value.(float64))
		}
		bs := strconv.AppendInt(num[:0], v, 10)
		return encodeInt(bs, code)

	case Decimal:
		v, ok := value.(float64)
		if !ok {
			v = float64(value.(int64))
		}
		if -1 >= v || v <= 1 {
			bs := strconv.AppendFloat(num[:0], v, 'f', -1, 64)
			return encodeSD(bs, code)
		}
		panic("collate invalid decimal")
	}
	panic("collate invalid number configuration")
}

func denormalizeFloat(code []byte, nt NumberKind) interface{} {
	var scratch [64]byte
	switch nt {
	case Float64:
		_, y := decodeFloat(code, scratch[:])
		res, err := strconv.ParseFloat(bytes2str(scratch[:y]), 64)
		if err != nil {
			panic(err)
		}
		return res

	case Int64:
		_, y := decodeInt(code, scratch[:])
		i, err := strconv.Atoi(bytes2str(scratch[:y]))
		if err != nil {
			panic(err)
		}
		return int64(i)

	case Decimal:
		_, y := decodeSD(code, scratch[:])
		res, err := strconv.ParseFloat(bytes2str(scratch[:y]), 64)
		if err != nil {
			panic(err)
		}
		return res
	}
	panic("collate gson denormalizeFloat bad configuration")
}
