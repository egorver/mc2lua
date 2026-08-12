package pipeline

import (
	"testing"

	"mc2lua/internal/model"
	"mc2lua/internal/runtime"
)

func BenchmarkLuaGenerator_Run(b *testing.B) {
	parts := make([]model.Part, 10000)
	for i := range parts {
		if i%2 == 0 {
			parts[i] = makePart("stone", "", 0, "minecraft:stone", "",
				model.Vector3{4, 4, 4}, model.Vector3{0, 0, 0}, model.Color{128, 128, 128}, "Slate")
		} else {
			parts[i] = makePart("elem 1", "oak_stairs", i/2+1, "minecraft:oak_stairs", "facing=north,half=bottom",
				model.Vector3{4, 2, 4}, model.Vector3{0, 0, 2}, model.Color{131, 84, 50}, "Wood")
		}
	}

	svc := NewLuaGenerator(runtime.NewFSMock())

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := svc.Run(parts, 4, "out.lua"); err != nil {
			b.Fatal(err)
		}
	}
}
