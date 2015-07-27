package cbor

import "testing"
import "fmt"
import "reflect"
import "encoding/json"

var _ = fmt.Sprintf("dummy")

var testcases = []string{
	//// null
	//"null",
	//// boolean
	//"true",
	//"false",
	//// integers
	//"10",
	//"0.1",
	//"-0.1",
	//"10.1",
	//"-10.1",
	//"-10E-1",
	//"-10e+1",
	//"10E-1",
	//"10e+1",
	//// string
	//`"true"`,
	//`"tru\"e"`,
	//`"tru\\e"`,
	//`"tru\be"`,
	//`"tru\fe"`,
	//`"tru\ne"`,
	//`"tru\re"`,
	//`"tru\te"`,
	//`"tru\u0123e"`,
	//`"汉语 / 漢語; Hàn\b \t\uef24yǔ "`,
	//// array
	//`[]`,
	//` [null,true,false,10,"tru\"e"]`,
	// object
	//`{}`,
	`{"a":null,"b":true,"c":false,"d\"":10,"e":"tru\"e", "f":[1,2]}`,
}

func TestSkipWS(t *testing.T) {
	ref := "hello  "
	if got := skipWS("  hello  ", AnsiSpace); got != ref {
		t.Errorf("expected %v got %v", ref, got)
	}
}

func TestJson(t *testing.T) {
	cborout, jsonout := make([]byte, 1024), make([]byte, 1024)
	config := NewDefaultConfig()
	var ref1, ref2 interface{}
	for _, tcase := range testcases {
		t.Logf("testcase - %v", tcase)
		_, n := config.ParseJson(tcase, cborout)
		if err := json.Unmarshal([]byte(tcase), &ref1); err != nil {
			t.Errorf("json.Unmarshal() failed for tcase %v: %v", tcase, err)
		}
		t.Logf("%v %v", cborout[:n], n)
		p, m := config.ToJson(cborout, jsonout)
		if p != n {
			t.Errorf("expected %v, got %v", n, p)
		}
		if err := json.Unmarshal(jsonout[:m], &ref2); err != nil {
			t.Errorf("json.Unmarshal() failed for cbor %v: %v", tcase, err)
		}
		if !reflect.DeepEqual(ref1, ref2) {
			t.Errorf("mismatch %v, got %v", ref1, ref2)
		}
	}
}
