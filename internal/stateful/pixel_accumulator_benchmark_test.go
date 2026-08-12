package stateful

import "testing"

var benchSink any

func BenchmarkPixelAccumulator(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		acc := NewPixelAccumulator()
		for j := 0; j < 1000000; j++ {
			acc.Add(128, 128, 128, 255)
		}
		result, err := acc.Result()
		if err != nil {
			b.Fatal(err)
		}
		benchSink = result
	}
}
