package gson

import "reflect"
import "unsafe"

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

func absInt8(val int8) uint8 {
	mask := val >> 7
	return uint8((mask + val) ^ mask)
}

func absInt16(val int16) uint16 {
	mask := val >> 15
	return uint16((mask + val) ^ mask)
}

func absInt32(val int32) uint32 {
	mask := val >> 31
	return uint32((mask + val) ^ mask)
}

func absInt64(val int64) uint64 {
	mask := val >> 63
	return uint64((mask + val) ^ mask)
}
