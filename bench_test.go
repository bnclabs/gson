//  Copyright (c) 2015 Couchbase, Inc.

package gson

import "encoding/json"
import "testing"

func BenchmarkMarshalJson(b *testing.B) {
	config := NewDefaultConfig().NumberKind(FloatNumber).SpaceKind(AnsiSpace)

	jsn := config.NewJson(testdataFile("testdata/typical.json"), -1)
	_, val := jsn.Tovalue()
	b.SetBytes(int64(len(jsn.Bytes())))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Marshal(val)
	}
}
