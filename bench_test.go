//  Copyright (c) 2015 Couchbase, Inc.

package gson

import "encoding/json"
import "testing"

func BenchmarkMarshalJson(b *testing.B) {
	txt := string(testdataFile("testdata/typical.json"))
	config := NewDefaultConfig().NumberKind(FloatNumber).SpaceKind(AnsiSpace)
	_, val := config.JsonToValue(txt)
	b.SetBytes(int64(len(txt)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Marshal(val)
	}
}
