package collate

import "testing"
import "bytes"
import "fmt"

var _ = fmt.Sprintf("dummy")

func TestGson2Collate(t *testing.T) {
	testcases := [][2]interface{}{
		[2]interface{}{nil, []byte{TypeNull, Terminator}},
		[2]interface{}{true, []byte{TypeTrue, Terminator}},
		[2]interface{}{false, []byte{TypeFalse, Terminator}},
		[2]interface{}{10000000000000,
			[]byte{5, 62, 62, 62, 50, 49, 52, 49, 45, 0}}, // ">>>2141-"
		[2]interface{}{Length(10),
			[]byte{7, 62, 62, 50, 49, 48, 0}}, // ">>210"
		[2]interface{}{string(MissingLiteral), []byte{1, 0}},
		[2]interface{}{"hello \x00world",
			[]byte{6, 104, 101, 108, 108, 111, 32, 0, 1, 119, 111, 114, 108,
				100, 0, 0}},
		[2]interface{}{[]interface{}{nil, true, false, 10, "hello"},
			[]byte{8, 2, 0, 4, 0, 3, 0, 5, 62, 62, 50, 49, 45, 0,
				6, 104, 101, 108, 108, 111, 0, 0, 0}},
		[2]interface{}{
			map[string]interface{}{"a": nil, "b": true, "c": false, "d": 10, "e": "hello"},
			[]byte{9, 7, 62, 53, 0, 6, 97, 0, 0, 2, 0, 6, 98, 0, 0, 4, 0,
				6, 99, 0, 0, 3, 0, 6, 100, 0, 0, 5, 62, 62, 50, 49, 45, 0, 6, 101,
				0, 0, 6, 104, 101, 108, 108, 111, 0, 0, 0}},
	}
	config := NewDefaultConfig()
	code := make([]byte, 1024)
	for _, tcase := range testcases {
		t.Logf("%v", tcase[0])
		n := gson2collate(tcase[0], code, config)
		if bytes.Compare(code[:n], tcase[1].([]byte)) != 0 {
			fmt.Println(string(code[:n]))
			t.Errorf("expected %v, got %v", tcase[1], code[:n])
		}
	}
}
