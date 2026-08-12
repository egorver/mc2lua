package pipeline

import (
	"testing"

	"mc2lua/internal/model"
)

func BenchmarkPartStylizer_Run(b *testing.B) {
	red := model.Color{200, 30, 30}
	styles := map[string]model.PartStyle{
		"minecraft:stone":       {All: surfPtr("rbxassetid://1", &red, nil)},
		"minecraft:grass_block": {Top: surfPtr("rbxassetid://top", &red, nil), All: surfPtr("rbxassetid://all", &red, nil)},
		"minecraft:furnace":     {Front: surfPtr("rbxassetid://front", &red, nil)},
	}

	parts := make([]model.Part, 10000)
	ids := []string{"minecraft:stone", "minecraft:grass_block", "minecraft:furnace"}
	for i := range parts {
		parts[i] = testPart(ids[i%len(ids)], visibleMask())
	}

	styleIdx := indexWithRotation("minecraft:furnace", 0, 90)
	svc := NewPartStylizer(&mockPartStyleMatcher{styles: styles}, &mockBrightnessMatcher{})

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchSink = svc.Run(parts, styleIdx)
	}
}
