package gson

import "encoding/json"
import "testing"

func BenchmarkMarshalJson(b *testing.B) {
	txt := string(testdataFile("testdata/typical.json"))
	config := NewConfig(FloatNumber, AnsiSpace)
	val, _ := config.Parse(txt)
	b.SetBytes(int64(len(txt)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Marshal(val)
	}
}
