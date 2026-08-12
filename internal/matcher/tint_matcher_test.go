package matcher

import (
	"testing"

	"mc2lua/internal/model"

	"github.com/stretchr/testify/require"
)

func TestTintMatcher_New(t *testing.T) {
	t.Parallel()

	t.Run("valid config", func(t *testing.T) {
		mock := &mockFS{data: []byte(`
tints:
  minecraft:grass_block: grass
  minecraft:oak_leaves: foliage
  minecraft:water_cauldron: water
  minecraft:redstone_wire: redstone
`)}
		m, err := NewTintMatcher(mock, "tints.yaml")
		require.NoError(t, err)
		require.NotNil(t, m)
	})

	t.Run("empty config", func(t *testing.T) {
		mock := &mockFS{data: []byte("")}
		m, err := NewTintMatcher(mock, "tints.yaml")
		require.NoError(t, err)
		require.NotNil(t, m)
	})

	t.Run("invalid yaml", func(t *testing.T) {
		mock := &mockFS{data: []byte("{{{")}
		_, err := NewTintMatcher(mock, "tints.yaml")
		require.Error(t, err)
	})

	t.Run("unknown tint type", func(t *testing.T) {
		mock := &mockFS{data: []byte("tints:\n  minecraft:foo: rainbow\n")}
		_, err := NewTintMatcher(mock, "tints.yaml")
		require.ErrorContains(t, err, "unknown tint type")
	})

	t.Run("file not found", func(t *testing.T) {
		_, err := NewTintMatcher(&mockFS{data: nil}, "tints.yaml")
		require.Error(t, err)
	})
}

func TestTintMatcher_Run(t *testing.T) {
	t.Parallel()

	mock := &mockFS{data: []byte(`
tints:
  minecraft:grass_block: grass
  minecraft:oak_leaves: foliage
  minecraft:water_cauldron: water
  minecraft:redstone_wire: redstone
`)}
	m, err := NewTintMatcher(mock, "tints.yaml")
	require.NoError(t, err)

	tests := []struct {
		name     string
		blockID  string
		wantType model.TintType
		wantOK   bool
	}{
		{name: "grass block", blockID: "minecraft:grass_block", wantType: model.TintGrass, wantOK: true},
		{name: "leaves", blockID: "minecraft:oak_leaves", wantType: model.TintFoliage, wantOK: true},
		{name: "water cauldron", blockID: "minecraft:water_cauldron", wantType: model.TintWater, wantOK: true},
		{name: "redstone wire", blockID: "minecraft:redstone_wire", wantType: model.TintRedstone, wantOK: true},
		{name: "unlisted block", blockID: "minecraft:stone", wantOK: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, ok := m.Run(tt.blockID)
			require.Equal(t, tt.wantOK, ok)
			require.Equal(t, tt.wantType, got)
		})
	}
}
