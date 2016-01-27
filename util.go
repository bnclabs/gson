//  Copyright (c) 2015 Couchbase, Inc.

package gson

import "reflect"
import "unsafe"
import "encoding/json"
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
// that is required for converting golang to cbor and vice-versa.
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
// that is required for converting golang to cbor and vice-versa.
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

func Fixtojson(config *Config, val interface{}) interface{} {
	var err error
	if s, ok := val.(json.Number); ok {
		val, err = strconv.ParseFloat(string(s), 64)
		if err != nil {
			panic(err)
		}
	}
	switch v := val.(type) {
	case int8:
		if config.nk == IntNumber {
			return int64(v)
		} else if config.nk == FloatNumber32 || config.nk == SmartNumber32 {
			return float32(v)
		} else {
			return float64(v)
		}
	case uint8:
		if config.nk == IntNumber {
			return int64(v)
		} else if config.nk == FloatNumber32 || config.nk == SmartNumber32 {
			return float32(v)
		} else {
			return float64(v)
		}
	case int16:
		if config.nk == IntNumber {
			return int64(v)
		} else if config.nk == FloatNumber32 || config.nk == SmartNumber32 {
			return float32(v)
		} else {
			return float64(v)
		}
	case uint16:
		if config.nk == IntNumber {
			return int64(v)
		} else if config.nk == FloatNumber32 || config.nk == SmartNumber32 {
			return float32(v)
		} else {
			return float64(v)
		}
	case int32:
		if config.nk == IntNumber {
			return int64(v)
		} else if config.nk == FloatNumber32 || config.nk == SmartNumber32 {
			return float32(v)
		} else {
			return float64(v)
		}
	case uint32:
		if config.nk == IntNumber {
			return int64(v)
		} else if config.nk == FloatNumber32 || config.nk == SmartNumber32 {
			return float32(v)
		} else {
			return float64(v)
		}
	case int64:
		if config.nk == IntNumber {
			return int64(v)
		} else if config.nk == FloatNumber32 || config.nk == SmartNumber32 {
			return float32(v)
		} else {
			return float64(v)
		}
	case uint64:
		if config.nk == IntNumber {
			return int64(v)
		} else if config.nk == FloatNumber32 || config.nk == SmartNumber32 {
			return float32(v)
		} else {
			return float64(v)
		}
	case float32:
		if config.nk == IntNumber {
			return int64(v)
		} else if config.nk == FloatNumber32 || config.nk == SmartNumber32 {
			return float32(v)
		} else {
			return float64(v)
		}
	case float64:
		if config.nk == IntNumber {
			return int64(v)
		} else if config.nk == FloatNumber32 || config.nk == SmartNumber32 {
			return float32(v)
		} else {
			return float64(v)
		}
	case map[string]interface{}:
		for key, x := range v {
			v[key] = Fixtojson(config, x)
		}
		return v
	case [][2]interface{}:
		m := make(map[string]interface{})
		for _, item := range v {
			m[item[0].(string)] = Fixtojson(config, item[1])
		}
		return m
	case []interface{}:
		for i, x := range v {
			v[i] = Fixtojson(config, x)
		}
		return v
	}
	return val
}

// numbers can be encoded as integers, or as small-decimal,
// or as floating-point - normalizeFloat() takes the number as
// float64 and based on the configuration encodes it integer or
// small-decimal or floating-point.
func normalizeFloat(value float64, code []byte, nt NumberKind) int {
	var num [64]byte
	switch nt {
	case FloatNumber:
		bs := strconv.AppendFloat(num[:0], value, 'e', -1, 64)
		return collateFloat(bs, code)

	case FloatNumber32:
		bs := strconv.AppendFloat(num[:0], value, 'e', -1, 32)
		return collateFloat(bs, code)

	case IntNumber:
		v := int64(value)
		bs := strconv.AppendInt(num[:0], v, 10)
		return collateInt(bs, code)

	case Decimal:
		if -1 >= value || value <= 1 {
			bs := strconv.AppendFloat(num[:0], value, 'f', -1, 64)
			return collateSD(bs, code)
		}
		panic("collate invalid decimal")
	}
	panic("SmartNumber32 or SmartNumber not supported for collation")
}

// numbers can be encoded as integers, or as small-decimal,
// or as floating-point - normalizeFloat() takes the number as
// int64 and based on the configuration encodes it integer or
// small-decimal or floating-point.
func normalizeInt64(value int64, code []byte, nt NumberKind) int {
	var num [64]byte
	switch nt {
	case FloatNumber:
		v := float64(value)
		bs := strconv.AppendFloat(num[:0], v, 'e', -1, 64)
		return collateFloat(bs, code)

	case FloatNumber32:
		v := float64(value)
		bs := strconv.AppendFloat(num[:0], v, 'e', -1, 32)
		return collateFloat(bs, code)

	case IntNumber:
		bs := strconv.AppendInt(num[:0], value, 10)
		return collateInt(bs, code)

	case Decimal:
		v := float64(value)
		if -1 >= v || v <= 1 {
			bs := strconv.AppendFloat(num[:0], v, 'f', -1, 64)
			return collateSD(bs, code)
		}
		panic("collate invalid decimal")
	}
	panic("SmartNumber32 or SmartNumber not supported for collation")
}

func denormalizeFloat(code []byte, nt NumberKind) float64 {
	var scratch [64]byte
	switch nt {
	case FloatNumber:
		_, y := collated2Float(code, scratch[:])
		res, err := strconv.ParseFloat(bytes2str(scratch[:y]), 64)
		if err != nil {
			panic(err)
		}
		return res

	case FloatNumber32:
		_, y := collated2Float(code, scratch[:])
		f, err := strconv.ParseFloat(bytes2str(scratch[:y]), 64)
		if err != nil {
			panic(err)
		}
		return f

	case Decimal:
		_, y := collated2SD(code, scratch[:])
		res, err := strconv.ParseFloat(bytes2str(scratch[:y]), 64)
		if err != nil {
			panic(err)
		}
		return res
	}
	panic("SmartNumber32/SmartNumber/IntNumber not configured for collation")
}

func denormalizeInt64(code []byte, nt NumberKind) int64 {
	var scratch [64]byte
	switch nt {
	case IntNumber:
		_, y := collated2Int(code, scratch[:])
		i, err := strconv.Atoi(bytes2str(scratch[:y]))
		if err != nil {
			panic(err)
		}
		return int64(i)
	}
	panic("only IntNumber is configured for collation")
}

func denormalizeFloatTojson(code []byte, text []byte, nt NumberKind) int {
	switch nt {
	case FloatNumber, FloatNumber32:
		_, y := collated2Float(code, text[:])
		return y

	case IntNumber:
		_, y := collated2Int(code, text[:])
		return y

	case Decimal:
		_, y := collated2SD(code, text[:])
		return y
	}
	panic("SmartNumber32 or SmartNumber not supported for collation")
}

// sort JSON property objects based on property names.
func sortProps(props map[string]interface{}, keys []string) []string {
	for k := range props {
		keys = append(keys, k)
	}

	return sortStrings(keys)
}

// bubble sort, moving to qsort should be atleast 40% faster.
func sortStrings(strs []string) []string {
	for ln := len(strs) - 1; ; ln-- {
		changed := false
		for i := 0; i < ln; i++ {
			if strs[i] > strs[i+1] {
				strs[i], strs[i+1] = strs[i+1], strs[i]
				changed = true
			}
		}
		if changed == false {
			break
		}
	}
	return strs
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

// bubble sort, moving to qsort should be atleast 40% faster.
func (kv kvrefs) sort() {
	for ln := len(kv) - 1; ; ln-- {
		changed := false
		for i := 0; i < ln; i++ {
			if kv[i].key > kv[i+1].key {
				kv[i], kv[i+1] = kv[i+1], kv[i]
				changed = true
			}
		}
		if changed == false {
			break
		}
	}
}
