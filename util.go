package gson

import "reflect"
import "unsafe"
import "sort"
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

// CborMap2golangMap transforms [][2]interface{} to map[string]interface{}
// that is required for converting golan->cbor.
func CborMap2golangMap(value interface{}) interface{} {
	switch items := value.(type) {
	case []interface{}:
		for i, item := range items {
			items[i] = CborMap2golangMap(item)
		}
		return items
	case [][2]interface{}:
		m := make(map[string]interface{})
		for _, item := range items {
			m[item[0].(string)] = CborMap2golangMap(item[1])
		}
		return m
	}
	return value
}

// GolangMap2cborMap transforms map[string]interface{} to [][2]interface{}
// that is required for converting golan->cbor.
func GolangMap2cborMap(value interface{}) interface{} {
	switch items := value.(type) {
	case []interface{}:
		for i, item := range items {
			items[i] = GolangMap2cborMap(item)
		}
		return items
	case map[string]interface{}:
		sl := make([][2]interface{}, 0, len(items))
		for k, v := range items {
			sl = append(sl, [2]interface{}{k, GolangMap2cborMap(v)})
		}
		return sl
	}
	return value
}

// numbers can be encoded as integers, or as small-decimal,
// or as floating-point - normalizeFloat() takes the number as
// float64 or int64 and based on the configuration encodes it
// integer or small-decimal or floating-point.
func normalizeFloat(value interface{}, code []byte, nt NumberKind) int {
	var num [64]byte
	switch nt {
	case FloatNumber:
		v, ok := value.(float64)
		if !ok {
			v = float64(value.(int64))
		}
		bs := strconv.AppendFloat(num[:0], v, 'e', -1, 64)
		return collateFloat(bs, code)

	case IntNumber:
		v, ok := value.(int64)
		if !ok {
			v = int64(value.(float64))
		}
		bs := strconv.AppendInt(num[:0], v, 10)
		return collateInt(bs, code)

	case Decimal:
		v, ok := value.(float64)
		if !ok {
			v = float64(value.(int64))
		}
		if -1 >= v || v <= 1 {
			bs := strconv.AppendFloat(num[:0], v, 'f', -1, 64)
			return collateSD(bs, code)
		}
		panic("collate invalid decimal")
	}
	panic("collate invalid number configuration")
}

func denormalizeFloat(code []byte, nt NumberKind) interface{} {
	var scratch [64]byte
	switch nt {
	case FloatNumber:
		_, y := collated2Float(code, scratch[:])
		res, err := strconv.ParseFloat(bytes2str(scratch[:y]), 64)
		if err != nil {
			panic(err)
		}
		return res

	case IntNumber:
		_, y := collated2Int(code, scratch[:])
		i, err := strconv.Atoi(bytes2str(scratch[:y]))
		if err != nil {
			panic(err)
		}
		return int64(i)

	case Decimal:
		_, y := collated2SD(code, scratch[:])
		res, err := strconv.ParseFloat(bytes2str(scratch[:y]), 64)
		if err != nil {
			panic(err)
		}
		return res
	}
	panic("collate gson denormalizeFloat bad configuration")
}

func denormalizeFloatTojson(code []byte, text []byte, nt NumberKind) int {
	switch nt {
	case FloatNumber:
		_, y := collated2Float(code, text[:])
		return y

	case IntNumber:
		_, y := collated2Int(code, text[:])
		return y

	case Decimal:
		_, y := collated2SD(code, text[:])
		return y
	}
	panic("collate gson denormalizeFloat bad configuration")
}

// sort JSON property objects based on property names.
func sortProps(props map[string]interface{}, keys []string) []string {
	for k := range props {
		keys = append(keys, k)
	}
	ss := sort.StringSlice(keys)
	ss.Sort()
	return keys
}

//---- data modelling to sort and collate JSON property items.

type kvref struct {
	key  string
	code []byte
}

type kvrefs []kvref

func (kv kvrefs) Len() int {
	return len(kv)
}

func (kv kvrefs) Less(i, j int) bool {
	return kv[i].key < kv[j].key
}

func (kv kvrefs) Swap(i, j int) {
	tmp := kv[i]
	kv[i] = kv[j]
	kv[j] = tmp
}
