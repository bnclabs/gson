package gson

import "reflect"
import "unsafe"
import "fmt"
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

// CborMap2golangMap used by validation tools.
// Transforms [][2]interface{} to map[string]interface{} that is required for
// converting golang to cbor and vice-versa.
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

// GolangMap2cborMap used by validation tools.
// Transforms map[string]interface{} to [][2]interface{}
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

// Fixtojson used by validation tools.
func Fixtojson(config *Config, val interface{}) interface{} {
	var err error

	if val == nil {
		return nil
	}

	if s, ok := val.(json.Number); ok {
		val, err = strconv.ParseFloat(string(s), 64)
		if err != nil {
			panic(err)
		}
	}

	switch v := val.(type) {
	case bool:
		return v
	case string:
		return v
	case int8:
		return float64(v)
	case uint8:
		return float64(v)
	case int16:
		return float64(v)
	case uint16:
		return float64(v)
	case int32:
		return float64(v)
	case uint32:
		return float64(v)
	case int64:
		if config.nk == FloatNumber {
			return float64(v)
		} else if config.nk == SmartNumber {
			return v
		}
	case uint64:
		if config.nk == FloatNumber {
			return float64(v)
		} else if config.nk == SmartNumber {
			return v
		}
	case float32:
		return float64(v)
	case float64:
		return v
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
	panic(fmt.Errorf("unreachable code, unexpected %T", val))
}

func collateFloat64(value float64, code []byte) int {
	var num [64]byte
	bs := strconv.AppendFloat(num[:0], value, 'e', -1, 64)
	return collateFloat(bs, code)
}

func collateUint64(value uint64, code []byte) int {
	var num, fltx [64]byte
	if value > 9007199254740991 {
		intx := strconv.AppendUint(num[:0], value, 10)
		fltx[0], fltx[1] = intx[0], '.'
		n := 2 + copy(fltx[2:], intx[1:])
		fltx[n], fltx[n+1], n = 'e', '+', n+2
		tmp := strconv.AppendInt(fltx[n:n], int64(len(intx[1:])), 10)
		ln := n + len(tmp)
		return collateFloat(fltx[:ln], code)

	}
	bs := strconv.AppendFloat(num[:0], float64(value), 'e', -1, 64)
	return collateFloat(bs, code)
}

func collateInt64(value int64, code []byte) int {
	var num, fltx [64]byte
	if value > 9007199254740991 {
		intx := strconv.AppendInt(num[:0], value, 10)
		fltx[0], fltx[1] = intx[0], '.'
		n := 2 + copy(fltx[2:], intx[1:])
		fltx[n], fltx[n+1], n = 'e', '+', n+2
		tmp := strconv.AppendInt(fltx[n:n], int64(len(intx[1:])), 10)
		ln := n + len(tmp)
		return collateFloat(fltx[:ln], code)

	} else if value < -9007199254740992 {
		intx := strconv.AppendInt(num[:0], value, 10)
		fltx[0], fltx[1], fltx[2] = intx[0], intx[1], '.'
		n := 3 + copy(fltx[3:], intx[2:])
		fltx[n], fltx[n+1], n = 'e', '+', n+2
		tmp := strconv.AppendInt(fltx[n:n], int64(len(intx[2:])), 10)
		ln := n + len(tmp)
		return collateFloat(fltx[:ln], code)

	} else {
		bs := strconv.AppendFloat(num[:0], float64(value), 'e', -1, 64)
		return collateFloat(bs, code)
	}
}

func collated2Number(code []byte, nk NumberKind) (uint64, int64, float64, int) {
	var mantissa, scratch [64]byte
	_, y := collated2Float(code, scratch[:])
	if nk == SmartNumber {
		dotat, _, exp, mant := parseFloat(scratch[:y], mantissa[:0])
		if exp >= 15 {
			x := len(mant) - dotat
			for i := 0; i < (exp - x); i++ {
				mant = append(mant, '0')
			}
			if mant[0] == '+' {
				mant = mant[1:]
			}
			ui, err := strconv.ParseUint(bytes2str(mant), 10, 64)
			if err == nil {
				return ui, 0, 0, 1
			}
			i, err := strconv.ParseInt(bytes2str(mant), 10, 64)
			if err == nil {
				return 0, i, 0, 2
			}
			panic(fmt.Errorf("unexpected number %v", string(scratch[:y])))
		}
	}
	f, err := strconv.ParseFloat(bytes2str(scratch[:y]), 64)
	if err != nil {
		panic(err)
	}
	return 0, 0, f, 3

}

func parseFloat(text []byte, m []byte) (int, int, int, []byte) {
	var err error
	var exp int

	dotat, expat := -1, -1
	for i, ch := range text {
		if ch == '.' {
			dotat = i
		} else if ch == 'e' || ch == 'E' {
			expat = i
		} else if expat > -1 && ch == '+' {
			expat = i
		} else if expat == -1 {
			m = append(m, ch)
		}
	}
	if expat > -1 {
		exp, err = strconv.Atoi(bytes2str(text[expat+1:]))
		if err != nil {
			panic(err)
		}
	}
	return dotat, expat, exp, m
}

func collated2Json(code []byte, text []byte, nk NumberKind) int {
	var num [64]byte

	ui, i, f, what := collated2Number(code, nk)
	switch what {
	case 1:
		nm := strconv.AppendUint(num[:0], ui, 10)
		return copy(text, nm)
	case 2:
		nm := strconv.AppendInt(num[:0], i, 10)
		return copy(text, nm)
	case 3:
		nm := strconv.AppendFloat(num[:0], f, 'e', -1, 64)
		return copy(text, nm)
	}
	panic("unreachable code")
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
