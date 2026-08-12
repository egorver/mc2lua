package pipeline

import (
	"testing"

	"mc2lua/internal/model"
	"mc2lua/internal/stateful"
)

func BenchmarkFaceCuller_Run(b *testing.B) {
	occ := stateful.NewOccupancyIndex()
	occ.FillRegion(-32, -4, -32, 64, 8, 64, true)
	regions := []model.Cuboid{
		{ID: "minecraft:stone", X: 0, Y: 0, Z: 0, Width: 8, Height: 4, Depth: 8},
	}
	svc := NewFaceCuller(&mockMergerPropsKeyBuilder{}, NewCuboidHelper())
	styles := styleIndex()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchSink = svc.Run(occ, regions, nil, nil, styles)
	}
}
