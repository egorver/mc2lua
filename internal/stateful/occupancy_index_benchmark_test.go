package stateful

import "testing"

func BenchmarkOccupancyIndex_FillRegion(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		occ := NewOccupancyIndex()
		occ.FillRegion(0, 0, 0, 16, 16, 16, true)
		benchSink = occ.Len()
	}
}
