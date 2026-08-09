package pipeline

import (
	"testing"

	"mc2lua/internal/model"

	"github.com/stretchr/testify/require"
)

func TestBlockVoxelIndexer_New(t *testing.T) {
	t.Parallel()

	pkb := &mockMergerPropsKeyBuilder{}
	svc := NewBlockVoxelIndexer(pkb)
	require.NotNil(t, svc)
}

func TestBlockVoxelIndexer_Run(t *testing.T) {
	t.Parallel()

	pkb := &mockMergerPropsKeyBuilder{
		runFn: func(props map[string]string) string {
			return ""
		},
	}

	stylesFull := styleIndex(struct {
		id        string
		prop      string
		alignment model.GridAlignment
	}{"minecraft:stone", "", model.GridFullBlock})

	tests := []struct {
		name    string
		blocks  []model.RawBlock
		wantLen int
	}{
		{
			name:    "empty input",
			blocks:  nil,
			wantLen: 0,
		},
		{
			name: "filters non-full blocks",
			blocks: []model.RawBlock{
				fullBlockAt("minecraft:stone", "", 0, 0, 0),
				fullBlockAt("minecraft:oak_fence", "", 1, 0, 0),
			},
			wantLen: 1,
		},
		{
			name: "adds all matching full blocks",
			blocks: []model.RawBlock{
				fullBlockAt("minecraft:stone", "", 0, 0, 0),
				fullBlockAt("minecraft:stone", "", 1, 0, 0),
			},
			wantLen: 2,
		},
		{
			name: "non-aligned block filtered out",
			blocks: []model.RawBlock{
				fullBlockAt("minecraft:stone", "", 0, 0, 0),
			},
			wantLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewBlockVoxelIndexer(pkb)
			grid := svc.Run(tt.blocks, stylesFull)
			require.Equal(t, tt.wantLen, len(grid.Blocks()))
			for _, b := range grid.Blocks() {
				require.False(t, b.Merged)
			}
		})
	}
}
