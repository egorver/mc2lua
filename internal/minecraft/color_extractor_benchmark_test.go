package minecraft

import (
	"image/color"
	"testing"

	"mc2lua/internal/runtime"
)

func BenchmarkColorExtractor_Run(b *testing.B) {
	fs := runtime.NewFSMock()
	extractor := NewColorExtractor(fs)
	roots := map[string][]string{"minecraft": {"assets/minecraft"}}
	fs.AddFile("assets/minecraft/textures/block/t.png",
		makePNG(b, 16, 16, func(x, y int) color.NRGBA { return color.NRGBA{150, 150, 150, 255} }), 0644)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c, err := extractor.Run(sample("block/t"), roots, "minecraft:test_block")
		if err != nil {
			b.Fatal(err)
		}
		benchSink = c
	}
}
