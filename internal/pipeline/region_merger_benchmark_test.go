package pipeline

import (
	"fmt"
	"math/rand"
	"testing"

	"mc2lua/internal/model"
)

var benchSink any

func BenchmarkRegionMerger_FindLargestRect(b *testing.B) {
	svc := NewRegionMerger(NewCuboidHelper())
	sizes := []int{16, 64, 128}
	for _, n := range sizes {
		b.Run(fmt.Sprintf("%dx%d", n, n), func(b *testing.B) {
			grid := make([][]bool, n)
			for i := range grid {
				grid[i] = make([]bool, n)
				for j := range grid[i] {
					grid[i][j] = true
				}
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				row, col, rows, cols := svc.findLargestRect(grid)
				benchSink = row + col + rows + cols
			}
		})
	}
}

func BenchmarkRegionMerger_MaxRectInHistogram(b *testing.B) {
	svc := NewRegionMerger(NewCuboidHelper())
	lengths := []int{16, 64, 128}
	for _, n := range lengths {
		b.Run(fmt.Sprintf("len_%d", n), func(b *testing.B) {
			heights := make([]int, n)
			for i := range heights {
				heights[i] = i%5 + 1
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				area, rowStart, colStart, cols, rows := svc.maxRectInHistogram(heights, n)
				benchSink = area + rowStart + colStart + cols + rows
			}
		})
	}
}

func BenchmarkRegionMerger_DecomposeLayer(b *testing.B) {
	svc := NewRegionMerger(NewCuboidHelper())

	b.Run("dense", func(b *testing.B) {
		var blocks []*model.MergedBlock
		for x := 0; x < 64; x++ {
			for z := 0; z < 64; z++ {
				blocks = append(blocks, &model.MergedBlock{ID: "minecraft:stone", X: x, Y: 0, Z: z})
			}
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			benchSink = svc.decomposeLayer(blocks)
		}
	})

	b.Run("L-shape", func(b *testing.B) {
		var blocks []*model.MergedBlock
		for x := 0; x < 32; x++ {
			blocks = append(blocks, &model.MergedBlock{ID: "minecraft:stone", X: x, Y: 0, Z: 0})
		}
		for z := 1; z < 32; z++ {
			blocks = append(blocks, &model.MergedBlock{ID: "minecraft:stone", X: 31, Y: 0, Z: z})
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			benchSink = svc.decomposeLayer(blocks)
		}
	})
}

func BenchmarkRegionMerger_Run(b *testing.B) {
	pkb := &mockMergerPropsKeyBuilder{
		runFn: func(props map[string]string) string { return "" },
	}
	styles := styleIndex(struct {
		id        string
		prop      string
		alignment model.GridAlignment
	}{"minecraft:stone", "", model.GridFullBlock})

	b.Run("dense cube 32x32x32", func(b *testing.B) {
		blocks := make([]model.RawBlock, 0, 32*32*32)
		for x := 0; x < 32; x++ {
			for y := 0; y < 32; y++ {
				for z := 0; z < 32; z++ {
					blocks = append(blocks, fullBlockAt("minecraft:stone", "", x, y, z))
				}
			}
		}
		indexer := NewBlockVoxelIndexer(pkb)
		svc := NewRegionMerger(NewCuboidHelper())
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			grid := indexer.Run(blocks, styles)
			benchSink = svc.Run(grid)
		}
	})

	b.Run("sparse world 64x64x16", func(b *testing.B) {
		rng := rand.New(rand.NewSource(42))
		blocks := make([]model.RawBlock, 0, 20000)
		for i := 0; i < 20000; i++ {
			blocks = append(blocks, fullBlockAt("minecraft:stone", "",
				rng.Intn(64), rng.Intn(16), rng.Intn(64)))
		}
		indexer := NewBlockVoxelIndexer(pkb)
		svc := NewRegionMerger(NewCuboidHelper())
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			grid := indexer.Run(blocks, styles)
			benchSink = svc.Run(grid)
		}
	})
}
