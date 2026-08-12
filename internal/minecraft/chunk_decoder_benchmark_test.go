package minecraft

import (
	"testing"
)

var benchSink any

func BenchmarkChunkDecoder_Run(b *testing.B) {
	sections := make([]interface{}, 0, 4)
	for s := 0; s < 4; s++ {
		sections = append(sections, map[string]interface{}{
			"Y": int8(s),
			"block_states": map[string]interface{}{
				"palette": []interface{}{
					map[string]interface{}{"Name": "minecraft:stone"},
					map[string]interface{}{"Name": "minecraft:dirt"},
				},
				"data": make([]int64, 256),
			},
		})
	}
	chunk := map[string]interface{}{
		"Status":   "full",
		"sections": sections,
	}
	data, err := encodeChunkNBT(chunk)
	if err != nil {
		b.Fatal(err)
	}

	svc := NewChunkDecoder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		blocks, err := svc.Run(data, 0, 0)
		if err != nil {
			b.Fatal(err)
		}
		benchSink = blocks
	}
}
