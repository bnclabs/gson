package gson

import "testing"
import "bytes"
import "fmt"

var _ = fmt.Sprintf("dummy")

func TestCbor2CollateNil(t *testing.T) {
	inp, ref, config := "null", `\x02\x00`, NewDefaultConfig()
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
	inp, ref, config := "true", `\x04\x00`, NewDefaultConfig()
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
	inp, ref, config := "false", `\x03\x00`, NewDefaultConfig()
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
		[2]string{`""`, `\x06\x00\x00`},
		[2]string{`"hello world"`, `\x06hello world\x00\x00`},
		[2]string{fmt.Sprintf(`"%s"`, MissingLiteral), `\x01\x00`},
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
	refcode := `\x06~[]{}falsenilNA~\x00\x00`
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
	inp, refcode := []byte("hello world"), `\nhello world\x00`
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
		[4]string{`[]`,
			`\b\x00`,
			`\b\a0\x00\x00`,
			`[]`},
		[4]string{`[null,true,false,10.0,"hello"]`,
			`\b\x02\x00\x04\x00\x03\x00\x05>>21-\x00\x06hello\x00\x00\x00`,
			`\b\a>5\x00\x02\x00\x04\x00\x03\x00\x05>>21-\x00` +
				`\x06hello\x00\x00\x00`,
			`[null,true,false,+0.1e+2,"hello"]`},
		[4]string{`[null,true,10.0,10.2,[],{"key":{}}]`,
			`\b\x02\x00\x04\x00\x05>>21-\x00\x05>>2102-\x00\b\x00` +
				`\t\a>1\x00\x06key\x00\x00\t\a0\x00\x00\x00\x00`,
			`\b\a>6\x00\x02\x00\x04\x00\x05>>21-\x00\x05>>2102-\x00` +
				`\b\a0\x00\x00\t\a>1\x00\x06key\x00\x00\t\a0\x00\x00\x00\x00`,
			`[null,true,+0.1e+2,+0.102e+2,[],{"key":{}}]`},
	}

	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 1024), 0)
	cbr := config.NewCbor(make([]byte, 1024), 0)
	cltback := config.NewCollate(make([]byte, 1024), 0)

	for _, tcase := range testcases {
		inp, refcode := tcase[0], tcase[1]

		t.Logf("%v", tcase[0])

		config.NewJson(
			[]byte(inp), -1).Tocollate(
			clt.Reset(nil)).Tocbor(
			cbr.Reset(nil)).Tocollate(cltback.Reset(nil))

		seqn := fmt.Sprintf("%q", cltback.Bytes())
		if seqn = seqn[1 : len(seqn)-1]; seqn != refcode {
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

		t.Logf("%v", tcase[0])

		config.NewJson(
			[]byte(inp), -1).Tocollate(
			clt.Reset(nil)).Tocbor(
			cbr.Reset(nil)).Tocollate(cltback.Reset(nil))

		seqn := fmt.Sprintf("%q", cltback.Bytes())
		if seqn = seqn[1 : len(seqn)-1]; seqn != refcode {
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

		t.Logf("%v", tcase[0])

		config.NewJson(
			[]byte(inp), -1).Tocollate(
			clt.Reset(nil)).Tocbor(
			cbr.Reset(nil)).Tocollate(cltback.Reset(nil))

		seqn := fmt.Sprintf("%q", cltback.Bytes())
		if seqn = seqn[1 : len(seqn)-1]; seqn != refcode {
			t.Errorf("expected %v, got %v", refcode, seqn)
		}
	}
}

func TestCbor2CollateMap(t *testing.T) {
	// with length prefix
	testcases := [][4]string{
		[4]string{
			`{}`,
			`\t\a0\x00\x00`,
			`\t\x00`,
			`{}`},
		[4]string{
			`{"a":null,"b":true,"c":false,"d":10.0,"e":"hello","f":["wo"]}`,
			`\t\a>6\x00\x06a\x00\x00\x02\x00\x06b\x00\x00\x04\x00\x06c` +
				`\x00\x00\x03\x00\x06d\x00\x00\x05>>21-\x00\x06e\x00\x00` +
				`\x06hello\x00\x00\x06f\x00\x00\b\x06wo\x00\x00\x00\x00`,
			`\t\x06a\x00\x00\x02\x00\x06b\x00\x00\x04\x00\x06c\x00\x00` +
				`\x03\x00\x06d\x00\x00\x05>>21-\x00\x06e\x00\x00` +
				`\x06hello\x00\x00\x06f\x00\x00\b\x06wo\x00\x00\x00\x00`,
			`{"a":null,"b":true,"c":false,"d":+0.1e+2,"e":"hello","f":["wo"]}`},
	}

	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 1024), 0)
	cbr := config.NewCbor(make([]byte, 1024), 0)
	cltback := config.NewCollate(make([]byte, 1024), 0)

	for _, tcase := range testcases {
		inp, refcode := tcase[0], tcase[1]

		t.Logf("%v", tcase[0])

		config.NewJson(
			[]byte(inp), -1).Tocollate(
			clt.Reset(nil)).Tocbor(
			cbr.Reset(nil)).Tocollate(cltback.Reset(nil))

		seqn := fmt.Sprintf("%q", cltback.Bytes())
		if seqn = seqn[1 : len(seqn)-1]; seqn != refcode {
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

		t.Logf("%v", tcase[0])

		config.NewJson(
			[]byte(inp), -1).Tocollate(
			clt.Reset(nil)).Tocbor(
			cbr.Reset(nil)).Tocollate(cltback.Reset(nil))

		seqn := fmt.Sprintf("%q", cltback.Bytes())
		if seqn = seqn[1 : len(seqn)-1]; seqn != refcode {
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

	config := NewDefaultConfig().SetMaxkeys(10)
	jsn := config.NewJson(data, -1)
	clt := config.NewCollate(make([]byte, 10*1024), 0)
	cbr := config.NewCbor(make([]byte, 10*1024), 0)

	jsn.Tocollate(clt)

	for i := 0; i < b.N; i++ {
		clt.Tocbor(cbr.Reset(nil))
	}
}
