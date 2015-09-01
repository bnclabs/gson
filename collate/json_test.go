package collate

import "testing"
import "fmt"

//import "reflect"

var _ = fmt.Sprintf("dummy")

func TestJson2CollateNil(t *testing.T) {
	inp, ref := "null", `\x02\x00`
	code, config := make([]byte, 1024), NewDefaultConfig()
	txt := make([]byte, 1024)
	_, n := scanToken(inp, code, config)
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
	_, n := scanToken(inp, code, config)
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
	_, n := scanToken(inp, code, config)
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
	// as float64 using Float64 configuration
	inp, refcode, reftxt := "10.2", `\x05>>2102-\x00`, "+0.102e+2"
	_, n := scanToken(inp, code, config.NumberKind(Float64))
	out := fmt.Sprintf("%q", code[:n])
	out = out[1 : len(out)-1]
	if out != refcode {
		t.Errorf("expected %v, got %v", refcode, out)
	}
	x, y := collate2json(code[:n], txt, config)
	if s := string(txt[:y]); s != reftxt || n != x {
		t.Errorf("expected {%v,%v}, got {%v,%v}", reftxt, n, s, x)
	}
	// as int64 using Float64 configuration
	inp, refcode, reftxt = "10", `\x05>>21-\x00`, "+0.1e+2"
	_, n = scanToken(inp, code, config.NumberKind(Float64))
	out = fmt.Sprintf("%q", code[:n])
	out = out[1 : len(out)-1]
	if out != refcode {
		t.Errorf("expected %v, got %v", refcode, out)
	}
	x, y = collate2json(code[:n], txt, config)
	if s := string(txt[:y]); s != reftxt || n != x {
		t.Errorf("expected {%v,%v}, got {%v,%v}", reftxt, n, s, x)
	}

	// as float64 using Int64 configuration
	inp, refcode, reftxt = "10.2", `\x05>>210\x00`, "+10"
	_, n = scanToken(inp, code, config.NumberKind(Int64))
	out = fmt.Sprintf("%q", code[:n])
	out = out[1 : len(out)-1]
	if out != refcode {
		t.Errorf("expected %v, got %v", refcode, out)
	}
	x, y = collate2json(code[:n], txt, config)
	if s := string(txt[:y]); s != reftxt || n != x {
		t.Errorf("expected {%v,%v}, got {%v,%v}", reftxt, n, s, x)
	}
	// as int64 using Int64 configuration
	inp, refcode, reftxt = "10", `\x05>>210\x00`, "+10"
	_, n = scanToken(inp, code, config.NumberKind(Int64))
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
	_, n = scanToken(inp, code, config.NumberKind(Decimal))
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
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		scanToken("10", code, config.NumberKind(Decimal))
	}()
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		scanToken("10.2", code, config.NumberKind(Decimal))
	}()
}

func TestJson2CollateString(t *testing.T) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	txt := make([]byte, 1024)
	// empty string
	inp, refcode, reftxt := `""`, `\x06\x00\x00`, `""`
	_, n := scanToken(inp, code, config)
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
	_, n = scanToken(inp, code, config)
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
	_, n = scanToken(inp, code, config)
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
	refcode, reftxt = `\x06~[]{}falsenilNA~\x00\x00`, inp
	_, n = scanToken(inp, code, config.UseMissing(false))
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
		_, n := scanToken(inp, code, config)
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
		_, n := scanToken(inp, code, config)
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
		_, n := scanToken(inp, code, config)
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
	// without length prefix
	for _, tcase := range testcases {
		t.Logf("%v", tcase[0])
		inp, refcode := tcase[0].(string), tcase[2].(string)
		reftxt := tcase[3].(string)
		_, n := scanToken(inp, code, config.SortbyPropertyLen(false))
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

func BenchmarkJsonCollNil(b *testing.B) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	for i := 0; i < b.N; i++ {
		scanToken("null", code, config)
	}
}

func BenchmarkCollJsonNil(b *testing.B) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	txt := make([]byte, 1024)
	_, n := scanToken("null", code, config)
	for i := 0; i < b.N; i++ {
		collate2json(code[:n], txt, config)
	}
}

func BenchmarkJsonCollTrue(b *testing.B) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	for i := 0; i < b.N; i++ {
		scanToken("true", code, config)
	}
}

func BenchmarkCollJsonTrue(b *testing.B) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	txt := make([]byte, 1024)
	_, n := scanToken("true", code, config)
	for i := 0; i < b.N; i++ {
		collate2json(code[:n], txt, config)
	}
}

func BenchmarkJsonCollFalse(b *testing.B) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	for i := 0; i < b.N; i++ {
		scanToken("false", code, config)
	}
}

func BenchmarkCollJsonFalse(b *testing.B) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	txt := make([]byte, 1024)
	_, n := scanToken("false", code, config)
	for i := 0; i < b.N; i++ {
		collate2json(code[:n], txt, config)
	}
}

func BenchmarkJsonCollF64(b *testing.B) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	for i := 0; i < b.N; i++ {
		scanToken("10.121312213123123", code, config)
	}
}

func BenchmarkCollJsonF64(b *testing.B) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	txt := make([]byte, 1024)
	_, n := scanToken("10.121312213123123", code, config)
	for i := 0; i < b.N; i++ {
		collate2json(code[:n], txt, config)
	}
}

func BenchmarkJsonCollI64(b *testing.B) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	for i := 0; i < b.N; i++ {
		scanToken("123456789", code, config)
	}
}

func BenchmarkCollJsonI64(b *testing.B) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	txt := make([]byte, 1024)
	_, n := scanToken("123456789", code, config)
	for i := 0; i < b.N; i++ {
		collate2json(code[:n], txt, config)
	}
}

func BenchmarkJsonCollMiss(b *testing.B) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	inp := fmt.Sprintf(`"%s"`, MissingLiteral)
	for i := 0; i < b.N; i++ {
		scanToken(inp, code, config)
	}
}

func BenchmarkCollJsonMiss(b *testing.B) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	txt := make([]byte, 1024)
	inp := fmt.Sprintf(`"%s"`, MissingLiteral)
	_, n := scanToken(inp, code, config)
	for i := 0; i < b.N; i++ {
		collate2json(code[:n], txt, config)
	}
}

func BenchmarkJsonCollStr(b *testing.B) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	for i := 0; i < b.N; i++ {
		scanToken(`"hello world"`, code, config)
	}
}

func BenchmarkCollJsonStr(b *testing.B) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	txt := make([]byte, 1024)
	_, n := scanToken(`"hello world"`, code, config)
	for i := 0; i < b.N; i++ {
		collate2json(code[:n], txt, config)
	}
}

func BenchmarkJsonCollArr(b *testing.B) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	inp := `[null,true,false,"hello world",10.23122312]`
	for i := 0; i < b.N; i++ {
		scanToken(inp, code, config)
	}
}

func BenchmarkCollJsonArr(b *testing.B) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	txt := make([]byte, 1024)
	inp := `[null,true,false,"hello world",10.23122312]`
	_, n := scanToken(inp, code, config)
	for i := 0; i < b.N; i++ {
		collate2json(code[:n], txt, config)
	}
}

func BenchmarkJsonCollMap(b *testing.B) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	inp := `{"key1":null,"key2":true,"key3":false,"key4":"hello world",` +
		`"key5":10.23122312}`
	for i := 0; i < b.N; i++ {
		scanToken(inp, code, config)
	}
}

func BenchmarkCollJsonMap(b *testing.B) {
	code, config := make([]byte, 1024), NewDefaultConfig()
	txt := make([]byte, 1024)
	inp := `{"key1":null,"key2":true,"key3":false,"key4":"hello world",` +
		`"key5":10.23122312}`
	_, n := scanToken(inp, code, config)
	for i := 0; i < b.N; i++ {
		collate2json(code[:n], txt, config)
	}
}
