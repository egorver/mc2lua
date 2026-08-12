package minecraft

import (
	"fmt"
	"testing"
)

func BenchmarkTextureResolver_Run(b *testing.B) {
	const chainLen = 50
	textures := make(map[string]string, chainLen+3)
	for i := 0; i < chainLen; i++ {
		textures[fmt.Sprintf("t%d", i)] = "#" + fmt.Sprintf("t%d", i+1)
	}
	textures[fmt.Sprintf("t%d", chainLen)] = "block/stone"
	textures["cycle_a"] = "#cycle_b"
	textures["cycle_b"] = "#cycle_a"

	svc := NewTextureResolver()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchSink = svc.Run(textures)
	}
}
