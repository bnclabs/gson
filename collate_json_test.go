//  Copyright (c) 2015 Couchbase, Inc.

package gson

import "testing"
import "fmt"

var _ = fmt.Sprintf("dummy")

func TestJson2CollateNil(t *testing.T) {
	inp, ref := "null", `\x02\x00`
	code, config := make([]byte, 1024), NewDefaultConfig()
	txt := make([]byte, 1024)
	_, n := json2collate(inp, code, config)
	out := fmt.Sprintf("%q", code[:n])
	out = out[1 : len(out)-1]
	if out != ref {
		t.Errorf("expected %v, got %v", ref, out)
	}
	x, y := collate2json(code[:n], txt, config)
	if s := string(txt[:y]); s != "null" || x != n {
		t.Errorf("expected {%v,%v}, got {%v,%v}", inp, n, s, x)
	}
}

func TestJson2CollateTrue(t *testing.T) {
	inp, ref := "true", `\x04\x00`
	code, config := make([]byte, 1024), NewDefaultConfig()
	txt := make([]byte, 1024)
	_, n := json2collate(inp, code, config)
	out := fmt.Sprintf("%q", code[:n])
	out = out[1 : len(out)-1]
	if out != ref {
		t.Errorf("expected %v, got %v", ref, out)
	}
	x, y := collate2json(code[:n], txt, config)
	if s := string(txt[:y]); s != "true" || n != x {
		t.Errorf("expected {%v,%v}, got {%v,%v}", inp, n, s, x)
	}
}

func TestJson2CollateFalse(t *testing.T) {
	inp, ref := "false", `\x03\x00`
	code, config := make([]byte, 1024), NewDefaultConfig()
	txt := make([]byte, 1024)
	_, n := json2collate(inp, code, config)
	out := fmt.Sprintf("%q", code[:n])
	out = out[1 : len(out)-1]
	if out != ref {
		t.Errorf("expected %v, got %v", ref, out)
	}
	x, y := collate2json(code[:n], txt, config)
	if s := string(txt[:y]); s != "false" || n != x {
		t.Errorf("expected {%v,%v}, got {%v,%v}", inp, n, s, x)
	}
}

func TestJson2CollateNumber(t *testing.T) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	txt := make([]byte, 1024)
	// as float64 using FloatNumber configuration
	inp, refcode, reftxt := "10.2", `\x05>>2102-\x00`, "+0.102e+2"
	config = config.NumberKind(FloatNumber)
	_, n := json2collate(inp, code, config)
	out := fmt.Sprintf("%q", code[:n])
	out = out[1 : len(out)-1]
	if out != refcode {
		t.Errorf("expected %v, got %v", refcode, out)
	}
	x, y := collate2json(code[:n], txt, config)
	if s := string(txt[:y]); s != reftxt || n != x {
		t.Errorf("expected {%v,%v}, got {%v,%v}", reftxt, n, s, x)
	}
	// as int64 using FloatNumber configuration
	inp, refcode, reftxt = "10", `\x05>>21-\x00`, "+0.1e+2"
	_, n = json2collate(inp, code, config.NumberKind(FloatNumber))
	out = fmt.Sprintf("%q", code[:n])
	out = out[1 : len(out)-1]
	if out != refcode {
		t.Errorf("expected %v, got %v", refcode, out)
	}
	x, y = collate2json(code[:n], txt, config)
	if s := string(txt[:y]); s != reftxt || n != x {
		t.Errorf("expected {%v,%v}, got {%v,%v}", reftxt, n, s, x)
	}

	// as float64 using IntNumber configuration
	inp, refcode, reftxt = "10.2", `\x05>>210\x00`, "+10"
	config = config.NumberKind(IntNumber)
	_, n = json2collate(inp, code, config)
	out = fmt.Sprintf("%q", code[:n])
	out = out[1 : len(out)-1]
	if out != refcode {
		t.Errorf("expected %v, got %v", refcode, out)
	}
	x, y = collate2json(code[:n], txt, config)
	if s := string(txt[:y]); s != reftxt || n != x {
		t.Errorf("expected {%v,%v}, got {%v,%v}", reftxt, n, s, x)
	}
	// as int64 using IntNumber configuration
	inp, refcode, reftxt = "10", `\x05>>210\x00`, "+10"
	_, n = json2collate(inp, code, config.NumberKind(IntNumber))
	out = fmt.Sprintf("%q", code[:n])
	out = out[1 : len(out)-1]
	if out != refcode {
		t.Errorf("expected %v, got %v", refcode, out)
	}
	x, y = collate2json(code[:n], txt, config)
	if s := string(txt[:y]); s != reftxt || n != x {
		t.Errorf("expected {%v,%v}, got {%v,%v}", reftxt, n, s, x)
	}

	// as float64 using Decimal configuration
	inp, refcode, reftxt = "0.2", `\x05>2-\x00`, "+0.2"
	config = config.NumberKind(Decimal)
	_, n = json2collate(inp, code, config)
	out = fmt.Sprintf("%q", code[:n])
	out = out[1 : len(out)-1]
	if out != refcode {
		t.Errorf("expected %v, got %v", refcode, out)
	}
	x, y = collate2json(code[:n], txt, config)
	if s := string(txt[:y]); s != reftxt || n != x {
		t.Errorf("expected {%v,%v}, got {%v,%v}", reftxt, n, s, x)
	}
	// as int64 using Decimal configuration, expect a panic
	config = config.NumberKind(Decimal)
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		json2collate("10", code, config)
	}()
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		json2collate("10.2", code, config)
	}()
}

func TestJson2CollateString(t *testing.T) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	txt := make([]byte, 1024)
	// empty string
	inp, refcode, reftxt := `""`, `\x06\x00\x00`, `""`
	_, n := json2collate(inp, code, config)
	out := fmt.Sprintf("%q", code[:n])
	out = out[1 : len(out)-1]
	if out != refcode {
		t.Errorf("expected %v, got %v", refcode, out)
	}
	x, y := collate2json(code[:n], txt, config)
	if s := string(txt[:y]); s != reftxt || n != x {
		t.Errorf("expected {%v,%v}, got {%v,%v}", reftxt, n, s, x)
	}
	// normal string
	inp, refcode = `"hello world"`, `\x06hello world\x00\x00`
	reftxt = `"hello world"`
	_, n = json2collate(inp, code, config)
	out = fmt.Sprintf("%q", code[:n])
	out = out[1 : len(out)-1]
	if out != refcode {
		t.Errorf("expected %v, got %v", refcode, out)
	}
	x, y = collate2json(code[:n], txt, config)
	if s := string(txt[:y]); s != reftxt || n != x {
		t.Errorf("expected {%v,%v}, got {%v,%v}", reftxt, n, s, x)
	}
	// missing string
	inp, refcode = fmt.Sprintf(`"%s"`, MissingLiteral), `\x01\x00`
	reftxt = string(MissingLiteral)
	_, n = json2collate(inp, code, config)
	out = fmt.Sprintf("%q", code[:n])
	out = out[1 : len(out)-1]
	if out != refcode {
		t.Errorf("expected %v, got %v", refcode, out)
	}
	x, y = collate2json(code[:n], txt, config)
	if s := string(txt[:y]); s != reftxt || n != x {
		t.Errorf("expected {%v,%v}, got {%v,%v}", reftxt, n, s, x)
	}
	// missing string without doMissing configuration
	inp = fmt.Sprintf(`"%s"`, MissingLiteral)
	config = config.UseMissing(false)
	refcode, reftxt = `\x06~[]{}falsenilNA~\x00\x00`, inp
	_, n = json2collate(inp, code, config)
	out = fmt.Sprintf("%q", code[:n])
	out = out[1 : len(out)-1]
	if out != refcode {
		t.Errorf("expected %v, got %v", refcode, out)
	}
	x, y = collate2json(code[:n], txt, config)
	if s := string(txt[:y]); s != reftxt || n != x {
		t.Errorf("expected {%v,%v}, got {%v,%v}", reftxt, n, s, x)
	}
}

func TestJson2CollateArray(t *testing.T) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	txt := make([]byte, 1024)
	// without length prefix
	testcases := [][4]interface{}{
		[4]interface{}{`[null,true,false,10.0,"hello"]`,
			`\b\x02\x00\x04\x00\x03\x00\x05>>21-\x00\x06hello\x00\x00\x00`,
			`\b\a>5\x00\x02\x00\x04\x00\x03\x00\x05>>21-\x00` +
				`\x06hello\x00\x00\x00`,
			`[null,true,false,+0.1e+2,"hello"]`},
		[4]interface{}{`[]`,
			`\b\x00`,
			`\b\a0\x00\x00`,
			`[]`},
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
		reftxt := tcase[3].(string)
		_, n := json2collate(inp, code, config)
		out := fmt.Sprintf("%q", code[:n])
		out = out[1 : len(out)-1]
		if out != refcode {
			t.Errorf("expected %v, got %v", refcode, out)
		}
		_, y := collate2json(code[:n], txt, config)
		if s := string(txt[:y]); s != reftxt {
			t.Errorf("expected %v, got %v", reftxt, s)
		}
	}
	// with length prefix
	config = config.SortbyArrayLen(true)
	for _, tcase := range testcases {
		t.Logf("%v", tcase[0])
		inp, refcode := tcase[0].(string), tcase[2].(string)
		reftxt := tcase[3].(string)
		_, n := json2collate(inp, code, config)
		out := fmt.Sprintf("%q", code[:n])
		out = out[1 : len(out)-1]
		if out != refcode {
			t.Errorf("expected %v, got %v", refcode, out)
		}
		_, y := collate2json(code[:n], txt, config)
		if s := string(txt[:y]); s != reftxt {
			t.Errorf("expected %v, got %v", reftxt, s)
		}
	}
}

func TestJson2CollateMap(t *testing.T) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	txt := make([]byte, 1024)
	// with length prefix
	testcases := [][4]interface{}{
		[4]interface{}{
			`{"a":null,"b":true,"c":false,"d":10.0,"e":"hello","f":["world"]}`,
			`\t\a>6\x00\x06a\x00\x00\x02\x00\x06b\x00\x00\x04\x00\x06c` +
				`\x00\x00\x03\x00\x06d\x00\x00\x05>>21-\x00\x06e\x00\x00` +
				`\x06hello\x00\x00\x06f\x00\x00\b\x06world\x00\x00\x00\x00`,
			`\t\x06a\x00\x00\x02\x00\x06b\x00\x00\x04\x00\x06c\x00\x00` +
				`\x03\x00\x06d\x00\x00\x05>>21-\x00\x06e\x00\x00` +
				`\x06hello\x00\x00\x06f\x00\x00\b\x06world\x00\x00\x00\x00`,
			`{"a":null,"b":true,"c":false,"d":+0.1e+2,"e":"hello","f":["world"]}`},
	}
	for _, tcase := range testcases {
		t.Logf("%v", tcase[0])
		inp, refcode := tcase[0].(string), tcase[1].(string)
		reftxt := tcase[3].(string)
		_, n := json2collate(inp, code, config)
		out := fmt.Sprintf("%q", code[:n])
		out = out[1 : len(out)-1]
		if out != refcode {
			t.Errorf("expected %v, got %v", refcode, out)
		}
		_, y := collate2json(code[:n], txt, config)
		if s := string(txt[:y]); s != reftxt {
			t.Errorf("expected %v, got %v", reftxt, s)
		}
	}
	// without length prefix, and different length for keys
	config = NewDefaultConfig().SetMaxkeys(10).SortbyPropertyLen(false)
	for _, tcase := range testcases {
		t.Logf("%v", tcase[0])
		inp, refcode := tcase[0].(string), tcase[2].(string)
		reftxt := tcase[3].(string)
		_, n := json2collate(inp, code, config)
		out := fmt.Sprintf("%q", code[:n])
		out = out[1 : len(out)-1]
		if out != refcode {
			t.Errorf("expected %v, got %v", refcode, out)
		}
		_, y := collate2json(code[:n], txt, config)
		if s := string(txt[:y]); s != reftxt {
			t.Errorf("expected {%v,%v}, got {%v,%v}", reftxt, s)
		}
	}
}

func BenchmarkJson2CollNil(b *testing.B) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	for i := 0; i < b.N; i++ {
		json2collate("null", code, config)
	}
}

func BenchmarkColl2JsonNil(b *testing.B) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	txt := make([]byte, 1024)
	_, n := json2collate("null", code, config)
	for i := 0; i < b.N; i++ {
		collate2json(code[:n], txt, config)
	}
}

func BenchmarkJson2CollTrue(b *testing.B) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	for i := 0; i < b.N; i++ {
		json2collate("true", code, config)
	}
}

func BenchmarkColl2JsonTrue(b *testing.B) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	txt := make([]byte, 1024)
	_, n := json2collate("true", code, config)
	for i := 0; i < b.N; i++ {
		collate2json(code[:n], txt, config)
	}
}

func BenchmarkJson2CollFalse(b *testing.B) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	for i := 0; i < b.N; i++ {
		json2collate("false", code, config)
	}
}

func BenchmarkColl2JsonFalse(b *testing.B) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	txt := make([]byte, 1024)
	_, n := json2collate("false", code, config)
	for i := 0; i < b.N; i++ {
		collate2json(code[:n], txt, config)
	}
}

func BenchmarkJson2CollF64(b *testing.B) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	for i := 0; i < b.N; i++ {
		json2collate("10.121312213123123", code, config)
	}
}

func BenchmarkColl2JsonF64(b *testing.B) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	txt := make([]byte, 1024)
	_, n := json2collate("10.121312213123123", code, config)
	for i := 0; i < b.N; i++ {
		collate2json(code[:n], txt, config)
	}
}

func BenchmarkJson2CollI64(b *testing.B) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	for i := 0; i < b.N; i++ {
		json2collate("123456789", code, config)
	}
}

func BenchmarkColl2JsonI64(b *testing.B) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	txt := make([]byte, 1024)
	_, n := json2collate("123456789", code, config)
	for i := 0; i < b.N; i++ {
		collate2json(code[:n], txt, config)
	}
}

func BenchmarkJson2CollMiss(b *testing.B) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	inp := fmt.Sprintf(`"%s"`, MissingLiteral)
	for i := 0; i < b.N; i++ {
		json2collate(inp, code, config)
	}
}

func BenchmarkColl2JsonMiss(b *testing.B) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	txt := make([]byte, 1024)
	inp := fmt.Sprintf(`"%s"`, MissingLiteral)
	_, n := json2collate(inp, code, config)
	for i := 0; i < b.N; i++ {
		collate2json(code[:n], txt, config)
	}
}

func BenchmarkJson2CollStr(b *testing.B) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	for i := 0; i < b.N; i++ {
		json2collate(`"hello world"`, code, config)
	}
}

func BenchmarkColl2JsonStr(b *testing.B) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	txt := make([]byte, 1024)
	_, n := json2collate(`"hello world"`, code, config)
	for i := 0; i < b.N; i++ {
		collate2json(code[:n], txt, config)
	}
}

func BenchmarkJson2CollArr(b *testing.B) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	inp := `[null,true,false,"hello world",10.23122312]`
	for i := 0; i < b.N; i++ {
		json2collate(inp, code, config)
	}
}

func BenchmarkColl2JsonArr(b *testing.B) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	txt := make([]byte, 1024)
	inp := `[null,true,false,"hello world",10.23122312]`
	_, n := json2collate(inp, code, config)
	for i := 0; i < b.N; i++ {
		collate2json(code[:n], txt, config)
	}
}

func BenchmarkJson2CollMap(b *testing.B) {
	code, config := make([]byte, 1024), NewDefaultConfig().SetMaxkeys(10)
	inp := `{"key1":null,"key2":true,"key3":false,"key4":"hello world",` +
		`"key5":10.23122312}`
	for i := 0; i < b.N; i++ {
		json2collate(inp, code, config)
	}
}

func BenchmarkColl2JsonMap(b *testing.B) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	txt := make([]byte, 1024)
	inp := `{"key1":null,"key2":true,"key3":false,"key4":"hello world",` +
		`"key5":10.23122312}`
	_, n := json2collate(inp, code, config)
	for i := 0; i < b.N; i++ {
		collate2json(code[:n], txt, config)
	}
}

func BenchmarkJson2CollTyp(b *testing.B) {
	code, config := make([]byte, 10*1024), NewDefaultConfig().SetMaxkeys(10)
	txt := string(testdataFile("testdata/typical.json"))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json2collate(txt, code, config)
	}
}

func BenchmarkColl2JsonTyp(b *testing.B) {
	code, config := make([]byte, 10*1024), NewDefaultConfig()
	out := make([]byte, 10*1024)
	txt := string(testdataFile("testdata/typical.json"))
	_, n := json2collate(txt, code, config)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collate2json(code[:n], out, config)
	}
}
