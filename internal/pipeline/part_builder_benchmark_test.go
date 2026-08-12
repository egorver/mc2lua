package pipeline

import (
	"testing"

	"mc2lua/internal/model"
	"mc2lua/internal/stateful"
)

func BenchmarkPartBuilder_Run(b *testing.B) {
	pkb := &mockGeneratorPropsKeyBuilder{
		runFn: func(props map[string]string) string { return "" },
	}

	idx := stateful.NewStyleIndex()
	idx.Add("minecraft:stone", "", makeStyledBlock("minecraft:stone", "", model.GridFullBlock,
		makeStyledElement(0, 0, 0, 16, 16, 16, model.Color{128, 128, 128}, "Slate"),
	))
	idx.Add("minecraft:oak_stairs", "", makeStyledBlock("minecraft:oak_stairs", "", model.GridNotAligned,
		makeStyledElement(0, 0, 0, 16, 8, 16, model.Color{131, 84, 50}, "Wood"),
		makeStyledElement(0, 8, 0, 16, 16, 8, model.Color{131, 84, 50}, "Wood"),
	))
	styles := *idx

	blockCuboids := make([]model.Cuboid, 1000)
	for i := range blockCuboids {
		blockCuboids[i] = model.Cuboid{
			ID: "minecraft:stone", X: float64(i), Y: 0, Z: 0,
			Width: 1, Height: 1, Depth: 1,
		}
	}
	blocks := make([]model.RawBlock, 500)
	for i := range blocks {
		blocks[i] = model.RawBlock{ID: "minecraft:oak_stairs", X: i, Y: 1, Z: 0}
	}

	visibility := makeVisibility(blockCuboids, nil, blocks)
	svc := NewPartBuilder(pkb)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parts, err := svc.Run(blocks, blockCuboids, nil, visibility, styles, 4)
		if err != nil {
			b.Fatal(err)
		}
		benchSink = parts
	}
}
