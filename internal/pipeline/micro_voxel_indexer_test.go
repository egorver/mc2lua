package pipeline

import (
	"testing"

	"mc2lua/internal/index"
	"mc2lua/internal/model"

	"github.com/stretchr/testify/require"
)

func TestMicroVoxelIndexer_New(t *testing.T) {
	t.Parallel()

	pkb := &mockMergerPropsKeyBuilder{}
	svc := NewMicroVoxelIndexer(pkb)
	require.NotNil(t, svc)
}

func TestMicroVoxelIndexer_Run(t *testing.T) {
	t.Parallel()

	pkb := &mockMergerPropsKeyBuilder{
		runFn: func(props map[string]string) string {
			return ""
		},
	}

	styles := index.NewStyleIndex()
	styles.Add("minecraft:stone_slab", "", model.StyledBlock{
		ID: "minecraft:stone_slab", PropsKey: "",
		GridAlignment: model.GridSubBlock,
		Elements: []model.StyledElement{
			{From: model.Vector3{0, 0, 0}, To: model.Vector3{16, 8, 16}, Shade: true},
		},
	})
	styles.Add("minecraft:stone", "", model.StyledBlock{
		ID: "minecraft:stone", PropsKey: "",
		GridAlignment: model.GridFullBlock,
		Elements: []model.StyledElement{
			{From: model.Vector3{0, 0, 0}, To: model.Vector3{16, 16, 16}, Shade: true},
		},
	})

	tests := []struct {
		name      string
		blocks    []model.RawBlock
		wantLen   int
		wantCoord func(t *testing.T, blocks []*model.MergedBlock)
	}{
		{
			name:    "empty input",
			blocks:  nil,
			wantLen: 0,
		},
		{
			name: "slab produces 32 micro-blocks",
			blocks: []model.RawBlock{
				{ID: "minecraft:stone_slab", X: 0, Y: 0, Z: 0},
			},
			wantLen: 32,
			wantCoord: func(t *testing.T, blocks []*model.MergedBlock) {
				seen := make(map[[3]int]bool)
				for _, b := range blocks {
					key := [3]int{b.X, b.Y, b.Z}
					require.False(t, seen[key], "duplicate block at %v", key)
					seen[key] = true
					require.Equal(t, "minecraft:stone_slab", b.ID)
					require.Equal(t, 0, b.X/4)
					require.Equal(t, 0, b.Y/4)
					require.Equal(t, 0, b.Z/4)
					require.True(t, b.Y >= 0 && b.Y <= 1)
				}
			},
		},
		{
			name: "full block filtered out",
			blocks: []model.RawBlock{
				{ID: "minecraft:stone", X: 0, Y: 0, Z: 0},
			},
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewMicroVoxelIndexer(pkb)
			grid := svc.Run(tt.blocks, *styles)
			require.Equal(t, tt.wantLen, len(grid.Blocks()))
			if tt.wantCoord != nil {
				tt.wantCoord(t, grid.Blocks())
			}
		})
	}
}
