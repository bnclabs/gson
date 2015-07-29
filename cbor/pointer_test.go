package cbor

import "testing"
import "fmt"

var _ = fmt.Sprintf("dummy")

var tcasesJSONPointers = []string{
	``,
	`/`,
	"/foo",
	"/foo/0",
	"/a~1b",
	"/c%d",
	"/e^f",
	"/g|h",
	"/i\\j",
	"/k\"l",
	"/ ",
	"/m~0n",
	"/g~1n~1r",
	"/g/n/r",
}

func TestCborPointer(t *testing.T) {
	buf, out := make([]byte, 1024), make([]byte, 1024)
	for _, tcase := range tcasesJSONPointers {
		t.Logf(tcase)
		n := FromJsonPointer([]byte(tcase), buf)
		m := ToJsonPointer(buf[:n], out)
		if result := string(out[:m]); tcase != result {
			t.Errorf("expected %q, got %q", tcase, result)
		}
	}
}
