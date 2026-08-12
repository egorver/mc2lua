package matcher

import (
	"fmt"
	"testing"
)

var benchSink any

func BenchmarkMaterialMatcher_Run(b *testing.B) {
	cfg := "mappings:\n"
	for i := 0; i < 50; i++ {
		cfg += fmt.Sprintf("  block_%d: Material%d\n", i, i)
	}
	cfg += "suffixes:\n  - _slab\n  - _stairs\n  - _fence\noverrides: {}\n"

	m, err := NewMaterialMatcher(&mockFS{data: []byte(cfg)}, "test.yaml")
	if err != nil {
		b.Fatal(err)
	}

	queries := make([]string, 1000)
	for i := range queries {
		queries[i] = fmt.Sprintf("minecraft:block_%d_slab", i%50)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, q := range queries {
			benchSink = m.Run(q)
		}
	}
}
