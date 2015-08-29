package collate

import "testing"
import "fmt"

var _ = fmt.Sprintf("dummy")

func TestGson2Collate(t *testing.T) {
	testcases := [][2]interface{}{
		[2]interface{}{nil,
			`\x02\x00`},
		[2]interface{}{true,
			`\x04\x00`},
		[2]interface{}{false,
			`\x03\x00`},
		[2]interface{}{10000000000000,
			`\x05>>>2141-\x00`},
		[2]interface{}{Length(10),
			`\a>>210\x00`},
		[2]interface{}{string(MissingLiteral),
			`\x01\x00`},
		[2]interface{}{"hello \x00world",
			`\x06hello \x00\x01world\x00\x00`},
		[2]interface{}{[]interface{}{nil, true, false, 10, "hello"},
			`\b\x02\x00\x04\x00\x03\x00\x05>>21-\x00\x06hello\x00\x00\x00`},
		[2]interface{}{
			map[string]interface{}{
				"a": nil, "b": true, "c": false, "d": 10, "e": "hello"},
			`\t\a>5\x00\x06a\x00\x00\x02\x00\x06b\x00\x00\x04\x00\x06c` +
				`\x00\x00\x03\x00\x06d\x00\x00\x05>>21-\x00\x06e\x00\x00` +
				`\x06hello\x00\x00\x00`},
	}
	config := NewDefaultConfig()
	code := make([]byte, 1024)
	for _, tcase := range testcases {
		t.Logf("%v", tcase[0])
		n := gson2collate(tcase[0], code, config)
		out := fmt.Sprintf("%q", code[:n])
		out = out[1 : len(out)-1]
		if out != tcase[1].(string) {
			t.Errorf("expected %v, got %v", tcase[1], out)
		}
	}
}
