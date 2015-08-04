// this file is used by test code.

package gson

type sgmts []string

type sgmtls [][]string

func (s sgmtls) Len() int {
	return len(s)
}

func (s sgmtls) Less(i, j int) bool {
	return len(s[i]) < len(s[j])
}

func (s sgmtls) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s sgmtls) filterAppend() sgmtls {
	news := make(sgmtls, 0, len(s))
	for _, x := range s {
		if ln := len(x); ln == 0 || x[ln-1] != "-" { // skip this
			news = append(news, x)
		}
	}
	return news
}

func setcontainer(s sgmts, doc, replica interface{}) {
	switch get(s, doc).(type) {
	case []interface{}:
		set(s, replica, []interface{}{})
	case map[string]interface{}:
		set(s, replica, map[string]interface{}{})
	default:
		// does not point to a container
	}
}
