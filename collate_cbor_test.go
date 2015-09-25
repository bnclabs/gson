//  Copyright (c) 2015 Couchbase, Inc.

package gson

import "testing"
import "fmt"

var _ = fmt.Sprintf("dummy")

func TestCbor2CollateNil(t *testing.T) {
	inp, ref := "null", `\x02\x00`
	code, config := make([]byte, 1024), NewDefaultConfig()
	out, coll := make([]byte, 1024), make([]byte, 1024)
	_, n := json2collate(inp, coll, config)

	_, n = collate2cbor(coll[:n], code, config)
	_, n = cbor2collate(code[:n], out, config)
	seqn := fmt.Sprintf("%q", out[:n])
	seqn = seqn[1 : len(seqn)-1]
	if seqn != ref {
		t.Errorf("expected %q, got %q", ref, seqn)
	}
}

func TestCbor2CollateTrue(t *testing.T) {
	inp, ref := "true", `\x04\x00`
	code, config := make([]byte, 1024), NewDefaultConfig()
	out, coll := make([]byte, 1024), make([]byte, 1024)
	_, n := json2collate(inp, coll, config)

	_, n = collate2cbor(coll[:n], code, config)
	_, n = cbor2collate(code[:n], out, config)
	seqn := fmt.Sprintf("%q", out[:n])
	seqn = seqn[1 : len(seqn)-1]
	if seqn != ref {
		t.Errorf("expected %v, got %v", ref, seqn)
	}
}

func TestCbor2CollateFalse(t *testing.T) {
	inp, ref := "false", `\x03\x00`
	code, config := make([]byte, 1024), NewDefaultConfig()
	out, coll := make([]byte, 1024), make([]byte, 1024)
	_, n := json2collate(inp, coll, config)

	_, n = collate2cbor(coll[:n], code, config)
	_, n = cbor2collate(code[:n], out, config)
	seqn := fmt.Sprintf("%q", out[:n])
	seqn = seqn[1 : len(seqn)-1]
	if seqn != ref {
		t.Errorf("expected %v, got %v", ref, seqn)
	}
}

func TestCbor2CollateNumber(t *testing.T) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	out, coll := make([]byte, 1024), make([]byte, 1024)
	testcases := [][3]interface{}{
		[3]interface{}{"10.2", `\x05>>2102-\x00`, FloatNumber},
		[3]interface{}{"10.2", `\x05>>2102-\x00`, FloatNumber32},
		[3]interface{}{"10", `\x05>>21-\x00`, FloatNumber},
		[3]interface{}{"10.2", `\x05>>210\x00`, IntNumber},
		[3]interface{}{"10", `\x05>>210\x00`, IntNumber},
		[3]interface{}{"-10", `\x05--789\x00`, IntNumber},
		[3]interface{}{"25", `\x05>>225\x00`, IntNumber},
		[3]interface{}{"-25", `\x05--774\x00`, IntNumber},
		[3]interface{}{"200", `\x05>>3200\x00`, IntNumber},
		[3]interface{}{"-200", `\x05--6799\x00`, IntNumber},
		[3]interface{}{"32767", `\x05>>532767\x00`, IntNumber},
		[3]interface{}{"-32767", `\x05--467232\x00`, IntNumber},
		[3]interface{}{"2147483647", `\x05>>>2102147483647\x00`, IntNumber},
		[3]interface{}{"-2147483648", `\x05---7897852516351\x00`, IntNumber},
		[3]interface{}{
			"4294967297",
			`\x05>>>2104294967297\x00`,
			IntNumber},
		[3]interface{}{
			"-4294967297",
			`\x05---7895705032702\x00`,
			IntNumber},
		[3]interface{}{"0.2", `\x05>2-\x00`, Decimal},
	}
	for _, tcase := range testcases {
		inp, refcode := tcase[0].(string), tcase[1].(string)
		t.Logf("%v", inp)
		nk := tcase[2].(NumberKind)
		config = config.NumberKind(nk)
		_, n := json2collate(inp, coll, config)
		_, n = collate2cbor(coll[:n], code, config)
		_, n = cbor2collate(code[:n], out, config)
		seqn := fmt.Sprintf("%q", out[:n])
		seqn = seqn[1 : len(seqn)-1]
		if seqn != refcode {
			t.Errorf("expected %v, got %v", refcode, seqn)
		}
	}
	testcases = [][3]interface{}{
		[3]interface{}{"-10", `\x05--78>\x00`, JsonNumber},
	}
	// skip collate2cbor
	for _, tcase := range testcases {
		inp, refcode := tcase[0].(string), tcase[1].(string)
		t.Logf("%v", inp)
		nk := tcase[2].(NumberKind)
		config = config.NumberKind(nk)
		_, n := json2cbor(inp, code, config)
		_, n = cbor2collate(code[:n], out, config)
		seqn := fmt.Sprintf("%q", out[:n])
		seqn = seqn[1 : len(seqn)-1]
		if seqn != refcode {
			t.Errorf("expected %v, got %v", refcode, seqn)
		}
	}
}

func TestCbor2CollateString(t *testing.T) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	out, coll := make([]byte, 1024), make([]byte, 1024)
	testcases := [][2]interface{}{
		[2]interface{}{`""`, `\x06\x00\x00`},
		[2]interface{}{`"hello world"`, `\x06hello world\x00\x00`},
		[2]interface{}{fmt.Sprintf(`"%s"`, MissingLiteral), `\x01\x00`},
	}
	for _, tcase := range testcases {
		inp, refcode := tcase[0].(string), tcase[1].(string)
		_, n := json2collate(inp, coll, config)
		_, n = collate2cbor(coll[:n], code, config)
		_, n = cbor2collate(code[:n], out, config)
		seqn := fmt.Sprintf("%q", out[:n])
		seqn = seqn[1 : len(seqn)-1]
		if seqn != refcode {
			t.Errorf("expected %v, got %v", refcode, seqn)
		}
	}
	// missing string without doMissing configuration
	inp := fmt.Sprintf(`"%s"`, MissingLiteral)
	refcode := `\x06~[]{}falsenilNA~\x00\x00`
	config = config.UseMissing(false)
	_, n := json2collate(inp, coll, config)

	_, n = collate2cbor(coll[:n], code, config)
	_, n = cbor2collate(code[:n], out, config)
	seqn := fmt.Sprintf("%q", out[:n])
	seqn = seqn[1 : len(seqn)-1]
	if seqn != refcode {
		t.Errorf("expected %v, got %v", refcode, seqn)
	}
}

func TestCbor2CollateJsonString(t *testing.T) {
	config := NewDefaultConfig().JsonString(true)
	buf, out := make([]byte, 64), make([]byte, 64)

	ref := `"汉语 / 漢語; Hàn\b \t\uef24yǔ "`
	n := tag2cbor(uint64(tagJsonString), buf)
	_, x := json2cbor(ref, buf[n:], config)
	n += x

	cbor2collate(buf[:n], out, config)
	//fmt.Printf("%q\n", out[:m])
}

func TestCbor2CollateBytes(t *testing.T) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	out, coll := make([]byte, 1024), make([]byte, 1024)
	inp, refcode := []byte("hello world"), `\nhello world\x00`
	n := gson2collate(inp, coll, config)
	_, n = collate2cbor(coll[:n], code, config)
	_, n = cbor2collate(code[:n], out, config)
	seqn := fmt.Sprintf("%q", out[:n])
	seqn = seqn[1 : len(seqn)-1]
	if seqn != refcode {
		t.Errorf("expected %v, got %v", refcode, seqn)
	}
}

func TestCbor2CollateArray(t *testing.T) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	out, coll := make([]byte, 1024), make([]byte, 1024)
	// without length prefix
	testcases := [][4]interface{}{
		[4]interface{}{`[]`,
			`\b\x00`,
			`\b\a0\x00\x00`,
			`[]`},
		[4]interface{}{`[null,true,false,10.0,"hello"]`,
			`\b\x02\x00\x04\x00\x03\x00\x05>>21-\x00\x06hello\x00\x00\x00`,
			`\b\a>5\x00\x02\x00\x04\x00\x03\x00\x05>>21-\x00` +
				`\x06hello\x00\x00\x00`,
			`[null,true,false,+0.1e+2,"hello"]`},
		[4]interface{}{`[null,true,10.0,10.2,[],{"key":{}}]`,
			`\b\x02\x00\x04\x00\x05>>21-\x00\x05>>2102-\x00\b\x00` +
				`\t\a>1\x00\x06key\x00\x00\t\a0\x00\x00\x00\x00`,
			`\b\a>6\x00\x02\x00\x04\x00\x05>>21-\x00\x05>>2102-\x00` +
				`\b\a0\x00\x00\t\a>1\x00\x06key\x00\x00\t\a0\x00\x00\x00\x00`,
			`[null,true,+0.1e+2,+0.102e+2,[],{"key":{}}]`},
	}
	for _, tcase := range testcases {
		t.Logf("%v", tcase[0])
		inp, refcode := tcase[0].(string), tcase[1].(string)
		_, n := json2collate(inp, coll, config)

		_, n = collate2cbor(coll[:n], code, config)
		_, n = cbor2collate(code[:n], out, config)
		seqn := fmt.Sprintf("%q", out[:n])
		seqn = seqn[1 : len(seqn)-1]
		if seqn != refcode {
			t.Errorf("expected %v, got %v", refcode, seqn)
		}
	}
	// with sort by length and length prefix
	config = config.SortbyArrayLen(true).ContainerEncoding(LengthPrefix)
	for _, tcase := range testcases {
		t.Logf("%v", tcase[0])
		inp, refcode := tcase[0].(string), tcase[2].(string)
		_, n := json2collate(inp, coll, config)

		_, n = collate2cbor(coll[:n], code, config)
		_, n = cbor2collate(code[:n], out, config)
		seqn := fmt.Sprintf("%q", out[:n])
		seqn = seqn[1 : len(seqn)-1]
		if seqn != refcode {
			t.Errorf("expected %v, got %v", refcode, seqn)
		}
	}
	// with sort by length and stream encoding
	config = config.SortbyArrayLen(true).ContainerEncoding(Stream)
	for _, tcase := range testcases {
		t.Logf("%v", tcase[0])
		inp, refcode := tcase[0].(string), tcase[2].(string)
		_, n := json2collate(inp, coll, config)

		_, n = collate2cbor(coll[:n], code, config)
		_, n = cbor2collate(code[:n], out, config)
		seqn := fmt.Sprintf("%q", out[:n])
		seqn = seqn[1 : len(seqn)-1]
		if seqn != refcode {
			t.Errorf("expected %v, got %v", refcode, seqn)
		}
	}
}

func TestCbor2CollateMap(t *testing.T) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	out, coll := make([]byte, 1024), make([]byte, 1024)
	// with length prefix
	testcases := [][4]interface{}{
		[4]interface{}{
			`{}`,
			`\t\a0\x00\x00`,
			`\t\x00`,
			`{}`},
		[4]interface{}{
			`{"a":null,"b":true,"c":false,"d":10.0,"e":"hello","f":["wo"]}`,
			`\t\a>6\x00\x06a\x00\x00\x02\x00\x06b\x00\x00\x04\x00\x06c` +
				`\x00\x00\x03\x00\x06d\x00\x00\x05>>21-\x00\x06e\x00\x00` +
				`\x06hello\x00\x00\x06f\x00\x00\b\x06wo\x00\x00\x00\x00`,
			`\t\x06a\x00\x00\x02\x00\x06b\x00\x00\x04\x00\x06c\x00\x00` +
				`\x03\x00\x06d\x00\x00\x05>>21-\x00\x06e\x00\x00` +
				`\x06hello\x00\x00\x06f\x00\x00\b\x06wo\x00\x00\x00\x00`,
			`{"a":null,"b":true,"c":false,"d":+0.1e+2,"e":"hello","f":["wo"]}`},
	}
	for _, tcase := range testcases {
		t.Logf("%v", tcase[0])
		inp, refcode := tcase[0].(string), tcase[1].(string)
		_, n := json2collate(inp, coll, config)

		_, n = collate2cbor(coll[:n], code, config)
		_, n = cbor2collate(code[:n], out, config)
		seqn := fmt.Sprintf("%q", out[:n])
		seqn = seqn[1 : len(seqn)-1]
		if seqn != refcode {
			t.Errorf("expected %v, got %v", refcode, seqn)
		}
	}
	// without length prefix, and different length for keys
	config = NewDefaultConfig().SetMaxkeys(10).SortbyPropertyLen(false)
	config = config.ContainerEncoding(LengthPrefix)
	for _, tcase := range testcases {
		t.Logf("%v", tcase[0])
		inp, refcode := tcase[0].(string), tcase[2].(string)
		_, n := json2collate(inp, coll, config)

		_, n = collate2cbor(coll[:n], code, config)
		_, n = cbor2collate(code[:n], out, config)
		seqn := fmt.Sprintf("%q", out[:n])
		seqn = seqn[1 : len(seqn)-1]
		if seqn != refcode {
			t.Errorf("expected %v, got %v", refcode, seqn)
		}
	}
}

//func BenchmarkJsonCollNil(b *testing.B) {
//    code, config := make([]byte, 1024), NewDefaultConfig()
//    for i := 0; i < b.N; i++ {
//        json2collate("null", code, config)
//    }
//}
//
//func BenchmarkCollJsonNil(b *testing.B) {
//    code, config := make([]byte, 1024), NewDefaultConfig()
//    txt := make([]byte, 1024)
//    _, n := json2collate("null", code, config)
//    for i := 0; i < b.N; i++ {
//        collate2json(code[:n], txt, config)
//    }
//}
//
//func BenchmarkJsonCollTrue(b *testing.B) {
//    code, config := make([]byte, 1024), NewDefaultConfig()
//    for i := 0; i < b.N; i++ {
//        json2collate("true", code, config)
//    }
//}
//
//func BenchmarkCollJsonTrue(b *testing.B) {
//    code, config := make([]byte, 1024), NewDefaultConfig()
//    txt := make([]byte, 1024)
//    _, n := json2collate("true", code, config)
//    for i := 0; i < b.N; i++ {
//        collate2json(code[:n], txt, config)
//    }
//}
//
//func BenchmarkJsonCollFalse(b *testing.B) {
//    code, config := make([]byte, 1024), NewDefaultConfig()
//    for i := 0; i < b.N; i++ {
//        json2collate("false", code, config)
//    }
//}
//
//func BenchmarkCollJsonFalse(b *testing.B) {
//    code, config := make([]byte, 1024), NewDefaultConfig()
//    txt := make([]byte, 1024)
//    _, n := json2collate("false", code, config)
//    for i := 0; i < b.N; i++ {
//        collate2json(code[:n], txt, config)
//    }
//}
//
//func BenchmarkJsonCollF64(b *testing.B) {
//    code, config := make([]byte, 1024), NewDefaultConfig()
//    for i := 0; i < b.N; i++ {
//        json2collate("10.121312213123123", code, config)
//    }
//}
//
//func BenchmarkCollJsonF64(b *testing.B) {
//    code, config := make([]byte, 1024), NewDefaultConfig()
//    txt := make([]byte, 1024)
//    _, n := json2collate("10.121312213123123", code, config)
//    for i := 0; i < b.N; i++ {
//        collate2json(code[:n], txt, config)
//    }
//}
//
//func BenchmarkJsonCollI64(b *testing.B) {
//    code, config := make([]byte, 1024), NewDefaultConfig()
//    for i := 0; i < b.N; i++ {
//        json2collate("123456789", code, config)
//    }
//}
//
//func BenchmarkCollJsonI64(b *testing.B) {
//    code, config := make([]byte, 1024), NewDefaultConfig()
//    txt := make([]byte, 1024)
//    _, n := json2collate("123456789", code, config)
//    for i := 0; i < b.N; i++ {
//        collate2json(code[:n], txt, config)
//    }
//}
//
//func BenchmarkJsonCollMiss(b *testing.B) {
//    code, config := make([]byte, 1024), NewDefaultConfig()
//    inp := fmt.Sprintf(`"%s"`, MissingLiteral)
//    for i := 0; i < b.N; i++ {
//        json2collate(inp, code, config)
//    }
//}
//
//func BenchmarkCollJsonMiss(b *testing.B) {
//    code, config := make([]byte, 1024), NewDefaultConfig()
//    txt := make([]byte, 1024)
//    inp := fmt.Sprintf(`"%s"`, MissingLiteral)
//    _, n := json2collate(inp, code, config)
//    for i := 0; i < b.N; i++ {
//        collate2json(code[:n], txt, config)
//    }
//}
//
//func BenchmarkJsonCollStr(b *testing.B) {
//    code, config := make([]byte, 1024), NewDefaultConfig()
//    for i := 0; i < b.N; i++ {
//        json2collate(`"hello world"`, code, config)
//    }
//}
//
//func BenchmarkCollJsonStr(b *testing.B) {
//    code, config := make([]byte, 1024), NewDefaultConfig()
//    txt := make([]byte, 1024)
//    _, n := json2collate(`"hello world"`, code, config)
//    for i := 0; i < b.N; i++ {
//        collate2json(code[:n], txt, config)
//    }
//}
//
//func BenchmarkJsonCollArr(b *testing.B) {
//    code, config := make([]byte, 1024), NewDefaultConfig()
//    inp := `[null,true,false,"hello world",10.23122312]`
//    for i := 0; i < b.N; i++ {
//        json2collate(inp, code, config)
//    }
//}
//
//func BenchmarkCollJsonArr(b *testing.B) {
//    code, config := make([]byte, 1024), NewDefaultConfig()
//    txt := make([]byte, 1024)
//    inp := `[null,true,false,"hello world",10.23122312]`
//    _, n := json2collate(inp, code, config)
//    for i := 0; i < b.N; i++ {
//        collate2json(code[:n], txt, config)
//    }
//}
//
//func BenchmarkJsonCollMap(b *testing.B) {
//    code, config := make([]byte, 1024), NewDefaultConfig().SetMaxkeys(10)
//    inp := `{"key1":null,"key2":true,"key3":false,"key4":"hello world",` +
//        `"key5":10.23122312}`
//    for i := 0; i < b.N; i++ {
//        json2collate(inp, code, config)
//    }
//}
//
//func BenchmarkCollJsonMap(b *testing.B) {
//    code, config := make([]byte, 1024), NewDefaultConfig()
//    txt := make([]byte, 1024)
//    inp := `{"key1":null,"key2":true,"key3":false,"key4":"hello world",` +
//        `"key5":10.23122312}`
//    _, n := json2collate(inp, code, config)
//    for i := 0; i < b.N; i++ {
//        collate2json(code[:n], txt, config)
//    }
//}
