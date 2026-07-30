package pipeline

import (
	"testing"

	"mc2lua/internal/model"

	"github.com/stretchr/testify/require"
)

func TestVoxelIndexer_New(t *testing.T) {
	t.Parallel()

	pkb := &mockMergerPropsKeyBuilder{}
	svc := NewVoxelIndexer(pkb)
	require.NotNil(t, svc)
}

func TestVoxelIndexer_Run(t *testing.T) {
	t.Parallel()

	pkb := &mockMergerPropsKeyBuilder{
		runFn: func(props map[string]string) string {
			return ""
		},
	}

	styles := styleIndex(struct {
		id   string
		prop string
		full bool
	}{"minecraft:stone", "", true})

	tests := []struct {
		name     string
		blocks   []model.RawBlock
		wantLen  int
		wantKeys []string
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
			name: "adds all matching blocks",
			blocks: []model.RawBlock{
				fullBlockAt("minecraft:stone", "", 0, 0, 0),
				fullBlockAt("minecraft:stone", "", 1, 0, 0),
			},
			wantLen: 2,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc := NewVoxelIndexer(pkb)
			grid := svc.Run(tt.blocks, styles)
			require.Equal(t, tt.wantLen, len(grid.Blocks()))
			for _, b := range grid.Blocks() {
				require.False(t, b.Merged)
			}
		})
	}
}
