package matcher

import (
	"testing"
)

func FuzzMaterialMatcher_Run(f *testing.F) {
	mock := &mockFS{data: []byte(`
mappings:
  planks: Wood
  log: Wood
  stone: Slate

suffixes:
  - _stairs
  - _slab

overrides:
  minecraft:chain: Metal
`)}
	m, err := NewMaterialMatcher(mock, "test.yaml")
	if err != nil {
		f.Fatal(err)
	}

	f.Add("minecraft:oak_planks")
	f.Add("minecraft:stone_stairs")
	f.Add("minecraft:unknown_block")

	f.Fuzz(func(t *testing.T, blockID string) {
		_ = m.Run(blockID)
	})
}
