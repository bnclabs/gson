package gson

import "testing"
import "bytes"
import "sort"
import "io/ioutil"
import "strings"
import "path"
import "fmt"

var _ = fmt.Sprintf("dummy")

func TestValueCompare(t *testing.T) {
	config := NewDefaultConfig()
	testcases := [][3]interface{}{
		// numbers
		[3]interface{}{uint64(10), float64(10), 0},
		[3]interface{}{uint64(10), float64(10.1), -1},
		[3]interface{}{uint64(10), int64(-10), 1},
		[3]interface{}{uint64(10), int64(10), 0},
		[3]interface{}{uint64(11), int64(10), 1},
		[3]interface{}{uint64(9), uint64(10), -1},
		[3]interface{}{uint64(10), uint64(10), 0},
		[3]interface{}{int64(10), int(11), -1},
		[3]interface{}{int64(11), int(10), 1},
		[3]interface{}{int64(10), int64(10), 0},
		[3]interface{}{int64(10), uint64(11), -1},
		[3]interface{}{int64(10), uint64(10), 0},
		[3]interface{}{int64(10), uint64(9), 1},
		[3]interface{}{float64(10), uint64(10), 0},
		[3]interface{}{float64(10.1), uint64(10), 1},
		[3]interface{}{float64(10), uint64(9), 1},
		// all others
		[3]interface{}{nil, nil, 0},
		[3]interface{}{nil, false, -1},
		[3]interface{}{nil, true, -1},
		[3]interface{}{nil, 10, -1},
		[3]interface{}{nil, "hello", -1},
		[3]interface{}{nil, []interface{}{10, 20}, -1},
		[3]interface{}{nil, map[string]interface{}{"key1": 10}, -1},
		[3]interface{}{false, nil, 1},
		[3]interface{}{false, false, 0},
		[3]interface{}{false, true, -1},
		[3]interface{}{false, 10, -1},
		[3]interface{}{false, "hello", -1},
		[3]interface{}{false, []interface{}{10, 20}, -1},
		[3]interface{}{false, map[string]interface{}{"key1": 10}, -1},
		[3]interface{}{true, nil, 1},
		[3]interface{}{true, false, 1},
		[3]interface{}{true, true, 0},
		[3]interface{}{true, 10, -1},
		[3]interface{}{true, "hello", -1},
		[3]interface{}{true, []interface{}{10, 20}, -1},
		[3]interface{}{true, map[string]interface{}{"key1": 10}, -1},
		[3]interface{}{10, nil, 1},
		[3]interface{}{10, false, 1},
		[3]interface{}{10, true, 1},
		[3]interface{}{10, 10, 0},
		[3]interface{}{10, "hello", -1},
		[3]interface{}{10, []interface{}{10, 20}, -1},
		[3]interface{}{10, map[string]interface{}{"key1": 10}, -1},
		[3]interface{}{[]interface{}{10}, nil, 1},
		[3]interface{}{[]interface{}{10}, false, 1},
		[3]interface{}{[]interface{}{10}, true, 1},
		[3]interface{}{[]interface{}{10}, 10, 1},
		[3]interface{}{[]interface{}{10}, "hello", 1},
		[3]interface{}{[]interface{}{10}, []interface{}{10}, 0},
		[3]interface{}{[]interface{}{10}, map[string]interface{}{"key1": 10}, -1},
		[3]interface{}{map[string]interface{}{"key1": 10}, nil, 1},
		[3]interface{}{map[string]interface{}{"key1": 10}, false, 1},
		[3]interface{}{map[string]interface{}{"key1": 10}, true, 1},
		[3]interface{}{map[string]interface{}{"key1": 10}, 10, 1},
		[3]interface{}{map[string]interface{}{"key1": 10}, "hello", 1},
		[3]interface{}{map[string]interface{}{"key1": 10}, []interface{}{10}, 1},
		[3]interface{}{map[string]interface{}{"key1": 10},
			map[string]interface{}{"key1": 10}, 0},
	}
	for _, tcase := range testcases {
		val1 := config.NewValue(tcase[0])
		val2 := config.NewValue(tcase[1])
		ref, cmp := tcase[2].(int), val1.Compare(val2)
		if cmp != ref {
			t.Errorf("for nil expected %v, got %v", ref, cmp)
		}
	}
}

func TestValueCollate(t *testing.T) {
	dirname := "testdata/collate"
	entries, err := ioutil.ReadDir(dirname)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		file := path.Join(dirname, entry.Name())
		if !strings.HasSuffix(file, ".ref") {
			out := strings.Join(collatefile(file), "\n")
			ref, err := ioutil.ReadFile(file + ".ref")
			if err != nil {
				t.Fatal(err)
			}
			if strings.Trim(string(ref), "\n") != out {
				//fmt.Println(string(ref))
				//fmt.Println(string(out))
				t.Fatalf("sort mismatch in %v", file)
			}
		}
	}
}

func collatefile(filename string) (outs []string) {
	s, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err.Error())
	}
	config := NewDefaultConfig()
	if strings.Contains(filename, "numbers") {
		config = config.SetNumberKind(SmartNumber)
	}
	return collateLines(config, s)
}

func collateLines(config *Config, s []byte) []string {
	texts, values := lines(s), make(valueList, 0)
	for i, text := range texts {
		jsn := config.NewJson(text, -1)
		_, val := jsn.Tovalue()
		values = append(values, valObj{i, config.NewValue(val)})
	}
	outs := doSort(texts, values)
	return outs
}

func doSort(texts [][]byte, values valueList) (outs []string) {
	sort.Sort(values)
	outs = make([]string, 0)
	for _, value := range values {
		outs = append(outs, string(texts[value.off]))
	}
	return
}

func lines(content []byte) [][]byte {
	content = bytes.Trim(content, "\r\n")
	return bytes.Split(content, []byte("\n"))
}

type valObj struct {
	off int
	val *Value
}

type valueList []valObj

func (values valueList) Len() int {
	return len(values)
}

func (values valueList) Less(i, j int) bool {
	return values[i].val.Compare(values[j].val) < 0
}

func (values valueList) Swap(i, j int) {
	values[i], values[j] = values[j], values[i]
}
