package gson

import "testing"
import "strings"
import "sort"

func TestSgmtsSort(t *testing.T) {
	txt := string(testdataFile("testdata/typical_pointers"))
	config := NewDefaultConfig()
	s := make(sgmtls, 0)
	for _, x := range strings.Split(txt, "\n") {
		s = append(s, config.ParsePointer(x, []string{}))
	}

	sort.Sort(s)
	l := len(s[0])
	for _, x := range s[1:] {
		if len(x) < l {
			t.Errorf("sort failure")
		}
	}
}

func TestFilterAppend(t *testing.T) {
	txt := string(testdataFile("testdata/typical_pointers"))
	config := NewDefaultConfig()
	s := make(sgmtls, 0)
	for _, x := range strings.Split(txt, "\n") {
		s = append(s, config.ParsePointer(x, []string{}))
	}

	s = s.filterAppend()
	for _, x := range s {
		if len(x) > 1 && x[len(x)-1] == "-" {
			t.Errorf("unexpected %v", x)
		}
	}
}
