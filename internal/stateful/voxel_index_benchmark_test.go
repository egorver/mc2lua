package stateful

import (
	"testing"

	"mc2lua/internal/model"
)

func BenchmarkVoxelIndex_AddAndGet(b *testing.B) {
	const n = 1000000
	blocks := make([]*model.MergedBlock, n)
	for i := range blocks {
		blocks[i] = &model.MergedBlock{
			ID: "minecraft:stone",
			X:  i % 100,
			Y:  (i / 100) % 100,
			Z:  i / 10000,
		}
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vg := NewVoxelIndex()
		for _, blk := range blocks {
			vg.AddBlock(blk)
		}
		sum := 0
		for _, blk := range blocks {
			if vg.GetBlock(blk.X, blk.Y, blk.Z) != nil {
				sum++
			}
		}
		benchSink = sum
	}
}
