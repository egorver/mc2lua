package matcher

import (
	"io"
	"io/fs"
	"testing"

	"github.com/stretchr/testify/require"
)

type mockFS struct {
	data []byte
}

func (m *mockFS) ReadFile(name string) ([]byte, error) {
	if m.data == nil {
		return nil, fs.ErrNotExist
	}
	return m.data, nil
}

func (m *mockFS) ReadDir(name string) ([]fs.DirEntry, error) {
	return nil, nil
}

func (m *mockFS) Create(name string) (io.WriteCloser, error) {
	return nil, nil
}

func TestMaterialMatcher_New(t *testing.T) {
	t.Parallel()

	t.Run("valid config", func(t *testing.T) {
		mock := &mockFS{data: []byte("mappings:\n  stone: Slate\n")}
		m, err := NewMaterialMatcher(mock, "test.yaml")
		require.NoError(t, err)
		require.NotNil(t, m)
	})

	t.Run("empty config", func(t *testing.T) {
		mock := &mockFS{data: []byte("")}
		m, err := NewMaterialMatcher(mock, "test.yaml")
		require.NoError(t, err)
		require.NotNil(t, m)
	})

	t.Run("invalid yaml", func(t *testing.T) {
		mock := &mockFS{data: []byte("{{{")}
		_, err := NewMaterialMatcher(mock, "test.yaml")
		require.Error(t, err)
	})

	t.Run("file not found", func(t *testing.T) {
		_, err := NewMaterialMatcher(&mockFS{data: nil}, "test.yaml")
		require.Error(t, err)
	})
}

func TestMaterialMatcher_Run(t *testing.T) {
	t.Parallel()

	mock := &mockFS{data: []byte(`
mappings:
  planks: Wood
  log: Wood
  glass: Glass
  stone: Slate
  netherrack: Cobblestone
  wool: Fabric
  grass: Grass
  fence: Wood

suffixes:
  - _stairs
  - _slab
  - _fence
  - _door

overrides:
  minecraft:chain: Metal
  minecraft:iron_door: Metal
  minecraft:moss_block: Grass
`)}

	m, err := NewMaterialMatcher(mock, "test.yaml")
	require.NoError(t, err)

	tests := []struct {
		name    string
		blockID string
		want    string
	}{
		{"oak planks", "minecraft:oak_planks", "Wood"},
		{"spruce log", "minecraft:spruce_log", "Wood"},
		{"glass block", "minecraft:glass", "Glass"},
		{"stone bricks", "minecraft:stone_bricks", "Slate"},
		{"netherrack", "minecraft:netherrack", "Cobblestone"},
		{"white wool", "minecraft:white_wool", "Fabric"},
		{"grass block", "minecraft:grass_block", "Grass"},
		{"oak fence (keyword in full name)", "minecraft:oak_fence", "Wood"},
		{"stone stairs (suffix match)", "minecraft:stone_stairs", "Slate"},
		{"stone slab (suffix match)", "minecraft:stone_slab", "Slate"},
		{"chain (override)", "minecraft:chain", "Metal"},
		{"iron door (override)", "minecraft:iron_door", "Metal"},
		{"moss block (override)", "minecraft:moss_block", "Grass"},
		{"unknown block (fallback)", "minecraft:unknown_block", "SmoothPlastic"},
		{"custom mod block (fallback)", "cobblemon:gilded_chest", "SmoothPlastic"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := m.Run(tt.blockID)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestMaterialMatcher_SuffixPath(t *testing.T) {
	t.Parallel()

	mock := &mockFS{data: []byte(`
mappings:
  nether_brick: NetherBrick

suffixes:
  - _frame

overrides: {}
`)}

	m, err := NewMaterialMatcher(mock, "test.yaml")
	require.NoError(t, err)

	tests := []struct {
		name    string
		blockID string
		want    string
	}{
		{
			name:    "suffix stripped then planks fallback",
			blockID: "minecraft:dirt_planks_frame",
			want:    "Wood",
		},
		{
			name:    "suffix stripped no planks fallback",
			blockID: "minecraft:stone_frame",
			want:    "SmoothPlastic",
		},
		{
			name:    "keyword match still works via suffix base",
			blockID: "minecraft:nether_brick_frame",
			want:    "NetherBrick",
		},
		{
			name:    "no suffix no keyword fallback",
			blockID: "minecraft:unknown_stuff",
			want:    "SmoothPlastic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := m.Run(tt.blockID)
			require.Equal(t, tt.want, got)
		})
	}
}
