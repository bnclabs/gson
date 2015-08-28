package collate

import "strconv"

// collate golang representation of a json object.
func gson2collate(obj interface{}, code []byte, config *Config) int {
	if obj == nil {
		code[0], code[1] = TypeNull, Terminator
		return 2
	}

	var num [64]byte
	n := 0

	switch value := obj.(type) {
	case bool:
		if value {
			code[n] = TypeTrue
		} else {
			code[n] = TypeFalse
		}
		code[n+1] = Terminator
		n += 2

	case float64:
		code[n] = TypeNumber
		n++
		n += normalizeFloat(value, code[n:], config.nt)
		code[n] = Terminator
		n++

	case int:
		switch config.nt {
		case Float64:
			n += gson2collate(float64(value), code[n:], config)
		case Int64:
			bs := strconv.AppendInt(num[:0], int64(value), 10)
			code[n] = TypeNumber
			n++
			n += encodeInt(bs, code[n:])
			code[n] = Terminator
			n++
		default:
			panic("collate decimal not configured")
		}

	case Length:
		bs := strconv.AppendInt(num[:0], int64(value), 10)
		code[n] = TypeLength
		n++
		n += encodeInt(bs, code[n:])
		code[n] = Terminator
		n++

	case Missing:
		if config.doMissing && MissingLiteral.Equal(string(value)) {
			code[n], code[n+1] = TypeMissing, Terminator
			n += 2
		}

	case string:
		if config.doMissing && MissingLiteral.Equal(value) {
			code[n], code[n+1] = TypeMissing, Terminator
			n += 2
		} else {
			code[n] = TypeString
			n++
			n += suffixEncodeString(str2bytes(value), code[n:])
			code[n] = Terminator
			n++
		}

	case []interface{}:
		code[n] = TypeArray
		n++
		if config.arrayLenPrefix {
			n += gson2collate(Length(len(value)), code[n:], config)
		}
		for _, val := range value {
			n += gson2collate(val, code[n:], config)
		}
		code[n] = Terminator
		n++

	case map[string]interface{}:
		code[n] = TypeObj
		n++
		if config.propertyLenPrefix {
			n += gson2collate(Length(len(value)), code[n:], config)
		}
		for _, key := range sortProps(value) {
			n += gson2collate(key, code[n:], config)        // encode key
			n += gson2collate(value[key], code[n:], config) // encode value
		}
		code[n] = Terminator
		n++

	default:
		panic("collate invalid golang type")
	}
	return n
}

// transform collated binary back to golang representation
// of a json object.
func collate2gson(code []byte, config *Config) (interface{}, int) {
	if len(code) == 0 {
		return nil, 0
	}

	var scratch [64]byte
	n := 1
	switch code[0] {
	case TypeMissing:
		return string(MissingLiteral), 2

	case TypeNull:
		return nil, 2

	case TypeTrue:
		return true, 2

	case TypeFalse:
		return false, 2

	case TypeLength:
		m := getDatum(code[n:])
		_, y := decodeInt(code[n:n+m-1], scratch[:]) // -1 to skip terminator
		length, err := strconv.Atoi(bytes2str(scratch[:y]))
		if err != nil {
			panic(err)
		}
		return Length(length), n + m

	case TypeNumber:
		m := getDatum(code[n:])
		f := denormalizeFloat(code[n:n+m-1], config.nt) // -1 to skip terminator
		return f, n + m

	case TypeString:
		s := make([]byte, encodedStringSize(code[n:]))
		x, y := suffixDecodeString(code[n:], s)
		return bytes2str(s[:y]), n + x

	case TypeArray:
		if config.arrayLenPrefix {
			m := getDatum(code[n:])
			_, y := decodeInt(code[n:], scratch[:])
			ln, err := strconv.Atoi(bytes2str(scratch[:y]))
			if err != nil {
				panic(err)
			}
			arr := make([]interface{}, ln)
			n += m
			for ; ln > 0; ln-- {
				item, y := collate2gson(code[n:], config)
				arr = append(arr, item)
				n += y
			}
			return arr, n
		}
		arr := make([]interface{}, 8)
		for code[n] != Terminator {
			item, y := collate2gson(code[n:], config)
			arr = append(arr, item)
			n += y
		}
		return arr, n

	case TypeObj:
		obj := make(map[string]interface{})
		if config.propertyLenPrefix {
			m := getDatum(code[n:])
			_, y := decodeInt(code[n:], scratch[:])
			ln, err := strconv.Atoi(bytes2str(scratch[:y]))
			if err != nil {
				panic(err)
			}
			n += m
			for ; ln > 0; ln-- {
				key, m := collate2gson(code[n:], config)
				n += m
				value, m := collate2gson(code[n:], config)
				obj[key.(string)] = value
				n += m
			}
			return obj, n
		}
		for code[n] != Terminator {
			key, m := collate2gson(code[n:], config)
			n += m
			value, m := collate2gson(code[n:], config)
			obj[key.(string)] = value
			n += m
		}
		return obj, n
	}
	panic("collate decode invalid binary")
}

func normalizeFloat(value float64, code []byte, nt NumberType) int {
	var num [64]byte
	switch nt {
	case Float64:
		bs := strconv.AppendFloat(num[:0], value, 'e', -1, 64)
		return encodeFloat(bs, code)

	case Int64:
		bs := strconv.AppendInt(num[:0], int64(value), 10)
		return encodeInt(bs, code)

	case Decimal:
		bs := strconv.AppendFloat(num[:0], value, 'f', -1, 64)
		return encodeSD(bs, code)
	}
	panic("collate invalid number configuration")
}

func denormalizeFloat(code []byte, nt NumberType) interface{} {
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
		return float64(i)

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

// get the collated datum based on Terminator and return the length
// of the datum.
func getDatum(code []byte) int {
	var i int
	var b byte
	for i, b = range code {
		if b == Terminator {
			return i + 1
		}
	}
	panic("collate decode terminator not found")
}
