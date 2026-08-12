package minecraft

import (
	"testing"

	"mc2lua/internal/model"
)

func BenchmarkBlockResolver_Run(b *testing.B) {
	bsp := &mockBlockstateParser{
		runFn: func(ns, blockID string, props map[string]string, namespaces map[string][]string) ([]blockstateMatch, error) {
			return []blockstateMatch{
				{Model: "minecraft:block/stone"},
				{Model: "minecraft:block/stone_alt", RotY: 90},
				{Model: "minecraft:block/stone_top"},
			}, nil
		},
	}
	mp := &mockModelParser{
		runFn: func(modelName string, namespaces map[string][]string) (*flattenedModel, error) {
			return &flattenedModel{
				Elements: []model.ModelElement{{From: model.Vector3{0, 0, 0}, To: model.Vector3{16, 16, 16}, Shade: true}},
				Textures: map[string]string{"particle": "block/stone", "all": "block/stone"},
			}, nil
		},
	}
	tr := &mockTextureResolver{
		runFn: func(textures map[string]string) map[string]string { return textures },
	}

	svc := NewBlockResolver(bsp, mp, tr, NewElementRotator())

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resolved, err := svc.Run("minecraft:stone", "", nil, nil)
		if err != nil {
			b.Fatal(err)
		}
		benchSink = resolved
	}
}
