package gson

import "testing"
import "bytes"
import "fmt"

var _ = fmt.Sprintf("dummy")

func TestCbor2CollateNil(t *testing.T) {
	inp, ref, config := "null", `\f\x00`, NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 1024), 0)
	cbr := config.NewCbor(make([]byte, 1024), 0)
	cltback := config.NewCollate(make([]byte, 1024), 0)

	config.NewJson([]byte(inp), -1).Tocollate(clt).Tocbor(cbr).Tocollate(cltback)

	seqn := fmt.Sprintf("%q", cltback.Bytes())
	if seqn = seqn[1 : len(seqn)-1]; seqn != ref {
		t.Errorf("expected %q, got %q", ref, seqn)
	}
}

func TestCbor2CollateTrue(t *testing.T) {
	inp, ref, config := "true", `\x0e\x00`, NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 1024), 0)
	cbr := config.NewCbor(make([]byte, 1024), 0)
	cltback := config.NewCollate(make([]byte, 1024), 0)

	config.NewJson([]byte(inp), -1).Tocollate(clt).Tocbor(cbr).Tocollate(cltback)

	seqn := fmt.Sprintf("%q", cltback.Bytes())
	if seqn = seqn[1 : len(seqn)-1]; seqn != ref {
		t.Errorf("expected %v, got %v", ref, seqn)
	}
}

func TestCbor2CollateFalse(t *testing.T) {
	inp, ref, config := "false", `\r\x00`, NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 1024), 0)
	cbr := config.NewCbor(make([]byte, 1024), 0)
	cltback := config.NewCollate(make([]byte, 1024), 0)

	config.NewJson([]byte(inp), -1).Tocollate(clt).Tocbor(cbr).Tocollate(cltback)

	seqn := fmt.Sprintf("%q", cltback.Bytes())
	if seqn = seqn[1 : len(seqn)-1]; seqn != ref {
		t.Errorf("expected %v, got %v", ref, seqn)
	}
}

func TestCbor2CollateNumber(t *testing.T) {
	testcases := [][3]interface{}{
		{"10.2", `\x0f>>2102-\x00`, FloatNumber},
		{"10", `\x0f>>21-\x00`, FloatNumber},
		{"10.2", `\x0f>>2102-\x00`, SmartNumber},
		{"10", `\x0f>>21-\x00`, SmartNumber},
		{"10", `\x0f>>21-\x00`, FloatNumber},
		{"-10", `\x0f--78>\x00`, SmartNumber},
		{"-10", `\x0f--78>\x00`, FloatNumber},
		{"200", `\x0f>>32-\x00`, SmartNumber},
		{"200", `\x0f>>32-\x00`, FloatNumber},
		{"-200", `\x0f--67>\x00`, SmartNumber},
		{"-200", `\x0f--67>\x00`, FloatNumber},
		{
			"4294967297", `\x0f>>>2104294967297-\x00`, FloatNumber},
		{
			"-4294967297", `\x0f---7895705032702>\x00`, FloatNumber},
		{
			"4294967297", `\x0f>>>2104294967297-\x00`, SmartNumber},
		{
			"-4294967297", `\x0f---7895705032702>\x00`, SmartNumber},
		{
			"9007199254740992", `\x0f>>>2169007199254740992-\x00`, FloatNumber},
		{
			"-9007199254740993", `\x0f---7830992800745259007>\x00`, FloatNumber},
		{
			"9007199254740992", `\x0f>>>2169007199254740992-\x00`, SmartNumber},

		{
			"-9007199254740993", `\x0f---7830992800745259006>\x00`, SmartNumber},
	}

	for _, tcase := range testcases {
		inp, refcode := tcase[0].(string), tcase[1].(string)
		nk := tcase[2].(NumberKind)

		t.Logf("%v", inp)

		config := NewDefaultConfig().SetNumberKind(nk)
		clt := config.NewCollate(make([]byte, 1024), 0)
		cbr := config.NewCbor(make([]byte, 1024), 0)
		cltback := config.NewCollate(make([]byte, 1024), 0)

		config.NewJson(
			[]byte(inp), -1).Tocollate(clt).Tocbor(cbr).Tocollate(cltback)

		seqn := fmt.Sprintf("%q", cltback.Bytes())
		if seqn = seqn[1 : len(seqn)-1]; seqn != refcode {
			t.Errorf("expected %v, got %v", refcode, seqn)
		}
	}
}

func TestCbor2CollateString(t *testing.T) {
	testcases := [][2]string{
		{`""`, `\x10\x00\x00`},
		{`"hello world"`, `\x10hello world\x00\x00`},
		{fmt.Sprintf(`"%s"`, MissingLiteral), `\v\x00`},
	}

	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 1024), 0)
	cbr := config.NewCbor(make([]byte, 1024), 0)
	cltback := config.NewCollate(make([]byte, 1024), 0)

	for _, tcase := range testcases {
		inp, refcode := tcase[0], tcase[1]

		t.Logf("%v", inp)

		config.NewJson(
			[]byte(inp), -1).Tocollate(
			clt.Reset(nil)).Tocbor(
			cbr.Reset(nil)).Tocollate(cltback.Reset(nil))

		seqn := fmt.Sprintf("%q", cltback.Bytes())
		if seqn = seqn[1 : len(seqn)-1]; seqn != refcode {
			t.Errorf("expected %v, got %v", refcode, seqn)
		}
	}

	// missing string without doMissing configuration
	inp := fmt.Sprintf(`"%s"`, MissingLiteral)
	refcode := `\x10~[]{}falsenilNA~\x00\x00`
	config = NewDefaultConfig().UseMissing(false)
	clt = config.NewCollate(make([]byte, 1024), 0)
	cbr = config.NewCbor(make([]byte, 1024), 0)
	cltback = config.NewCollate(make([]byte, 1024), 0)

	config.NewJson([]byte(inp), -1).Tocollate(clt).Tocbor(cbr).Tocollate(cltback)
	seqn := fmt.Sprintf("%q", cltback.Bytes())
	if seqn = seqn[1 : len(seqn)-1]; seqn != refcode {
		t.Errorf("expected %v, got %v", refcode, seqn)
	}

	// utf8 string
	inp = `"汉语 / 漢語; Hàn\b \t\uef24yǔ "`

	config = NewDefaultConfig()
	clt = config.NewCollate(make([]byte, 1024), 0)
	cbr = config.NewCbor(make([]byte, 1024), 0)
	cltback = config.NewCollate(make([]byte, 1024), 0)

	config.NewJson([]byte(inp), -1).Tocollate(clt).Tocbor(cbr).Tocollate(cltback)

	if bytes.Compare(clt.Bytes(), cltback.Bytes()) != 0 {
		t.Errorf("expected %v, got %v", clt.Bytes(), cltback.Bytes())
	}
}

func TestCbor2CollateBytes(t *testing.T) {
	inp, refcode := []byte("hello world"), `\x14hello world\x00`
	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 1024), 0)
	cbr := config.NewCbor(make([]byte, 1024), 0)
	cltback := config.NewCollate(make([]byte, 1024), 0)

	config.NewValue(inp).Tocollate(clt).Tocbor(cbr).Tocollate(cltback)
	seqn := fmt.Sprintf("%q", cltback.Bytes())
	if seqn = seqn[1 : len(seqn)-1]; seqn != refcode {
		t.Errorf("expected %v, got %v", refcode, seqn)
	}
}

func TestCbor2CollateArray(t *testing.T) {

	// without length prefix
	testcases := [][4]string{
		{`[]`,
			`\x12\x00`,
			`\x12\x110\x00\x00`,
			`[]`},
		{`[null,true,false,10.0,"hello"]`,
			`\x12\f\x00\x0e\x00\r\x00\x0f>>21-\x00\x10hello\x00\x00\x00`,
			`\x12\x11>5\x00\f\x00\x0e\x00\r\x00\x0f>>21-\x00\x10hello\x00` +
				`\x00\x00`,
			`[null,true,false,+0.1e+2,"hello"]`},
		{`[null,true,10.0,10.2,[],{"key":{}}]`,
			`\x12\f\x00\x0e\x00\x0f>>21-\x00\x0f>>2102-\x00\x12\x00` +
				`\x13\x11>1\x00\x10key\x00\x00\x13\x110\x00\x00\x00\x00`,
			`\x12\x11>6\x00\f\x00\x0e\x00\x0f>>21-\x00\x0f>>2102-\x00` +
				`\x12\x110\x00\x00\x13\x11>1\x00\x10key\x00\x00\x13\x110` +
				`\x00\x00\x00\x00`,
			`[null,true,+0.1e+2,+0.102e+2,[],{"key":{}}]`},
	}

	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 1024), 0)
	cbr := config.NewCbor(make([]byte, 1024), 0)
	cltback := config.NewCollate(make([]byte, 1024), 0)

	for _, tcase := range testcases {
		inp, refcode := tcase[0], tcase[1]

		config.NewJson(
			[]byte(inp), -1).Tocollate(
			clt.Reset(nil)).Tocbor(
			cbr.Reset(nil)).Tocollate(cltback.Reset(nil))

		seqn := fmt.Sprintf("%q", cltback.Bytes())
		if seqn = seqn[1 : len(seqn)-1]; seqn != refcode {
			t.Logf("%v", tcase[0])
			t.Errorf("expected %v, got %v", refcode, seqn)
		}
	}

	// with sort by length and length prefix
	config = config.SortbyArrayLen(true).SetContainerEncoding(LengthPrefix)
	clt = config.NewCollate(make([]byte, 1024), 0)
	cbr = config.NewCbor(make([]byte, 1024), 0)
	cltback = config.NewCollate(make([]byte, 1024), 0)

	for _, tcase := range testcases {
		inp, refcode := tcase[0], tcase[2]
		config.NewJson(
			[]byte(inp), -1).Tocollate(
			clt.Reset(nil)).Tocbor(
			cbr.Reset(nil)).Tocollate(cltback.Reset(nil))

		seqn := fmt.Sprintf("%q", cltback.Bytes())
		if seqn = seqn[1 : len(seqn)-1]; seqn != refcode {
			t.Logf("%v", tcase[0])
			t.Errorf("expected %v, got %v", refcode, seqn)
		}
	}

	// with sort by length and stream encoding
	config = config.SortbyArrayLen(true).SetContainerEncoding(Stream)
	clt = config.NewCollate(make([]byte, 1024), 0)
	cbr = config.NewCbor(make([]byte, 1024), 0)
	cltback = config.NewCollate(make([]byte, 1024), 0)

	for _, tcase := range testcases {
		inp, refcode := tcase[0], tcase[2]
		config.NewJson(
			[]byte(inp), -1).Tocollate(
			clt.Reset(nil)).Tocbor(
			cbr.Reset(nil)).Tocollate(cltback.Reset(nil))

		seqn := fmt.Sprintf("%q", cltback.Bytes())
		if seqn = seqn[1 : len(seqn)-1]; seqn != refcode {
			t.Logf("%v", tcase[0])
			t.Errorf("expected %v, got %v", refcode, seqn)
		}
	}
}

func TestCbor2CollateMap(t *testing.T) {
	// with length prefix
	testcases := [][4]string{
		{
			`{}`,
			`\x13\x110\x00\x00`,
			`\x13\x00`,
			`{}`},
		{
			`{"a":null,"b":true,"c":false,"d":10.0,"e":"hello","f":["wo"]}`,
			`\x13\x11>6\x00\x10a\x00\x00\f\x00\x10b\x00\x00\x0e\x00\x10c` +
				`\x00\x00\r\x00\x10d\x00\x00\x0f>>21-\x00\x10e\x00\x00\x10hello` +
				`\x00\x00\x10f\x00\x00\x12\x10wo\x00\x00\x00\x00`,
			`\x13\x10a\x00\x00\f\x00\x10b\x00\x00\x0e\x00\x10c\x00\x00` +
				`\r\x00\x10d\x00\x00\x0f>>21-\x00\x10e\x00\x00\x10hello` +
				`\x00\x00\x10f\x00\x00\x12\x10wo\x00\x00\x00\x00`,
			`{"a":null,"b":true,"c":false,"d":+0.1e+2,"e":"hello","f":["wo"]}`},
	}

	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 1024), 0)
	cbr := config.NewCbor(make([]byte, 1024), 0)
	cltback := config.NewCollate(make([]byte, 1024), 0)

	for _, tcase := range testcases {
		inp, refcode := tcase[0], tcase[1]
		config.NewJson(
			[]byte(inp), -1).Tocollate(
			clt.Reset(nil)).Tocbor(
			cbr.Reset(nil)).Tocollate(cltback.Reset(nil))

		seqn := fmt.Sprintf("%q", cltback.Bytes())
		if seqn = seqn[1 : len(seqn)-1]; seqn != refcode {
			t.Logf("%v", tcase[0])
			t.Errorf("expected %v, got %v", refcode, seqn)
		}
	}

	// without length prefix, and different length for keys
	config = NewDefaultConfig().SetMaxkeys(10).SortbyPropertyLen(false)
	config = config.SetContainerEncoding(LengthPrefix)
	clt = config.NewCollate(make([]byte, 1024), 0)
	cbr = config.NewCbor(make([]byte, 1024), 0)
	cltback = config.NewCollate(make([]byte, 1024), 0)

	for _, tcase := range testcases {
		inp, refcode := tcase[0], tcase[2]
		config.NewJson(
			[]byte(inp), -1).Tocollate(
			clt.Reset(nil)).Tocbor(
			cbr.Reset(nil)).Tocollate(cltback.Reset(nil))

		seqn := fmt.Sprintf("%q", cltback.Bytes())
		if seqn = seqn[1 : len(seqn)-1]; seqn != refcode {
			t.Logf("%v", tcase[0])
			t.Errorf("expected %v, got %v", refcode, seqn)
		}
	}
}

func BenchmarkColl2CborNil(b *testing.B) {
	config := NewDefaultConfig()
	jsn := config.NewJson([]byte("null"), -1)
	clt := config.NewCollate(make([]byte, 1024), 0)
	cbr := config.NewCbor(make([]byte, 1024), 0)

	jsn.Tocollate(clt)

	for i := 0; i < b.N; i++ {
		clt.Tocbor(cbr.Reset(nil))
	}
}

func BenchmarkColl2CborTrue(b *testing.B) {
	config := NewDefaultConfig()
	jsn := config.NewJson([]byte("true"), -1)
	cbr := config.NewCbor(make([]byte, 1024), 0)
	clt := config.NewCollate(make([]byte, 1024), 0)

	jsn.Tocollate(clt)

	for i := 0; i < b.N; i++ {
		clt.Tocbor(cbr.Reset(nil))
	}
}

func BenchmarkColl2CborFalse(b *testing.B) {
	config := NewDefaultConfig()
	jsn := config.NewJson([]byte("false"), -1)
	clt := config.NewCollate(make([]byte, 1024), 0)
	cbr := config.NewCbor(make([]byte, 1024), 0)

	jsn.Tocollate(clt)

	for i := 0; i < b.N; i++ {
		clt.Tocbor(cbr.Reset(nil))
	}
}

func BenchmarkColl2CborF64(b *testing.B) {
	config := NewDefaultConfig()
	jsn := config.NewJson([]byte("10.121312213123123"), -1)
	clt := config.NewCollate(make([]byte, 1024), 0)
	cbr := config.NewCbor(make([]byte, 1024), 0)

	jsn.Tocollate(clt)

	for i := 0; i < b.N; i++ {
		clt.Tocbor(cbr.Reset(nil))
	}
}

func BenchmarkColl2CborI64(b *testing.B) {
	config := NewDefaultConfig()
	jsn := config.NewJson([]byte("123456789"), -1)
	clt := config.NewCollate(make([]byte, 1024), 0)
	cbr := config.NewCbor(make([]byte, 1024), 0)

	jsn.Tocollate(clt)

	for i := 0; i < b.N; i++ {
		clt.Tocbor(cbr.Reset(nil))
	}
}

func BenchmarkColl2CborMiss(b *testing.B) {
	config := NewDefaultConfig()
	jsn := config.NewJson([]byte(fmt.Sprintf(`"%s"`, MissingLiteral)), -1)
	clt := config.NewCollate(make([]byte, 1024), 0)
	cbr := config.NewCbor(make([]byte, 1024), 0)

	jsn.Tocollate(clt)

	for i := 0; i < b.N; i++ {
		clt.Tocbor(cbr.Reset(nil))
	}
}

func BenchmarkColl2CborStr(b *testing.B) {
	config := NewDefaultConfig()
	jsn := config.NewJson([]byte(`"hello world"`), -1)
	clt := config.NewCollate(make([]byte, 1024), 0)
	cbr := config.NewCbor(make([]byte, 1024), 0)

	jsn.Tocollate(clt)

	for i := 0; i < b.N; i++ {
		clt.Tocbor(cbr.Reset(nil))
	}
}

func BenchmarkColl2CborArr(b *testing.B) {
	in := []byte(`[null,true,false,"hello world",10.23122312]`)

	config := NewDefaultConfig()
	jsn := config.NewJson(in, -1)
	clt := config.NewCollate(make([]byte, 1024), 0)
	cbr := config.NewCbor(make([]byte, 1024), 0)

	jsn.Tocollate(clt)

	for i := 0; i < b.N; i++ {
		clt.Tocbor(cbr.Reset(nil))
	}
}

func BenchmarkColl2CborMap(b *testing.B) {
	inp := `{"key1":null,"key2":true,"key3":false,"key4":"hello world",` +
		`"key5":10.23122312}`

	config := NewDefaultConfig()
	jsn := config.NewJson([]byte(inp), -1)
	clt := config.NewCollate(make([]byte, 1024), 0)
	cbr := config.NewCbor(make([]byte, 1024), 0)

	jsn.Tocollate(clt)

	for i := 0; i < b.N; i++ {
		clt.Tocbor(cbr.Reset(nil))
	}
}

func BenchmarkColl2CborTyp(b *testing.B) {
	data := testdataFile("testdata/typical.json")

	config := NewDefaultConfig().SetMaxkeys(100)
	jsn := config.NewJson(data, -1)
	clt := config.NewCollate(make([]byte, 10*1024), 0)
	cbr := config.NewCbor(make([]byte, 10*1024), 0)

	jsn.Tocollate(clt)

	for i := 0; i < b.N; i++ {
		clt.Tocbor(cbr.Reset(nil))
	}
}
