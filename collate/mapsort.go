package collate

import "sort"

// sort JSON property objects based on property names.
func sortProps(props map[string]interface{}) []string {
	keys := make([]string, 0, len(props))
	for k := range props {
		keys = append(keys, k)
	}
	ss := sort.StringSlice(keys)
	ss.Sort()
	return keys
}

type kvref struct {
	key  string
	code []byte
}

type kvrefs []kvref

func (kv kvrefs) Len() int {
	return len(kv)
}

func (kv kvrefs) Less(i, j int) bool {
	return kv[i].key < kv[j].key
}

func (kv kvrefs) Swap(i, j int) {
	tmp := kv[i]
	kv[i] = kv[j]
	kv[j] = tmp
}
