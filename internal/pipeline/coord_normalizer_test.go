package pipeline

import (
	"testing"

	"mc2lua/internal/model"

	"github.com/stretchr/testify/require"
)

func TestCoordNormalizer_Run(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		blocks     []model.Block
		noOffset   bool
		wantBlocks []model.Block
		wantErr    bool
	}{
		{
			name:    "empty blocks",
			blocks:  nil,
			wantErr: true,
		},
		{
			name:    "empty blocks slice",
			blocks:  []model.Block{},
			wantErr: true,
		},
		{
			name: "single block already at origin",
			blocks: []model.Block{
				{ID: "minecraft:stone", X: 0, Y: 0, Z: 0},
			},
			noOffset: false,
			wantBlocks: []model.Block{
				{ID: "minecraft:stone", X: 0, Y: 0, Z: 0},
			},
		},
		{
			name: "single block with positive coords",
			blocks: []model.Block{
				{ID: "minecraft:dirt", X: 10, Y: 5, Z: 20},
			},
			noOffset: false,
			wantBlocks: []model.Block{
				{ID: "minecraft:dirt", X: 0, Y: 0, Z: 0},
			},
		},
		{
			name: "single block with negative coords",
			blocks: []model.Block{
				{ID: "minecraft:grass_block", X: -15, Y: -3, Z: -42},
			},
			noOffset: false,
			wantBlocks: []model.Block{
				{ID: "minecraft:grass_block", X: 0, Y: 0, Z: 0},
			},
		},
		{
			name: "noOffset=true preserves Y",
			blocks: []model.Block{
				{ID: "minecraft:bedrock", X: 100, Y: -64, Z: 200},
			},
			noOffset: true,
			wantBlocks: []model.Block{
				{ID: "minecraft:bedrock", X: 0, Y: -64, Z: 0},
			},
		},
		{
			name: "noOffset=true with Y already zero",
			blocks: []model.Block{
				{ID: "minecraft:stone", X: 50, Y: 0, Z: 30},
			},
			noOffset: true,
			wantBlocks: []model.Block{
				{ID: "minecraft:stone", X: 0, Y: 0, Z: 0},
			},
		},
		{
			name: "multiple blocks - all positive",
			blocks: []model.Block{
				{ID: "minecraft:stone", X: 10, Y: 5, Z: 20},
				{ID: "minecraft:dirt", X: 12, Y: 5, Z: 25},
				{ID: "minecraft:grass_block", X: 15, Y: 7, Z: 22},
			},
			noOffset: false,
			wantBlocks: []model.Block{
				{ID: "minecraft:stone", X: 0, Y: 0, Z: 0},
				{ID: "minecraft:dirt", X: 2, Y: 0, Z: 5},
				{ID: "minecraft:grass_block", X: 5, Y: 2, Z: 2},
			},
		},
		{
			name: "multiple blocks - all negative",
			blocks: []model.Block{
				{ID: "minecraft:stone", X: -30, Y: -12, Z: -50},
				{ID: "minecraft:dirt", X: -25, Y: -10, Z: -45},
				{ID: "minecraft:grass_block", X: -28, Y: -12, Z: -48},
			},
			noOffset: false,
			wantBlocks: []model.Block{
				{ID: "minecraft:stone", X: 0, Y: 0, Z: 0},
				{ID: "minecraft:dirt", X: 5, Y: 2, Z: 5},
				{ID: "minecraft:grass_block", X: 2, Y: 0, Z: 2},
			},
		},
		{
			name: "multiple blocks - mixed positive and negative",
			blocks: []model.Block{
				{ID: "minecraft:stone", X: -10, Y: -5, Z: 3},
				{ID: "minecraft:dirt", X: 0, Y: 10, Z: 7},
				{ID: "minecraft:grass_block", X: 5, Y: -3, Z: 15},
			},
			noOffset: false,
			wantBlocks: []model.Block{
				{ID: "minecraft:stone", X: 0, Y: 0, Z: 0},
				{ID: "minecraft:dirt", X: 10, Y: 15, Z: 4},
				{ID: "minecraft:grass_block", X: 15, Y: 2, Z: 12},
			},
		},
		{
			name: "multiple blocks with noOffset=true",
			blocks: []model.Block{
				{ID: "minecraft:stone", X: -10, Y: -5, Z: 3},
				{ID: "minecraft:dirt", X: 0, Y: 10, Z: 7},
			},
			noOffset: true,
			wantBlocks: []model.Block{
				{ID: "minecraft:stone", X: 0, Y: -5, Z: 0},
				{ID: "minecraft:dirt", X: 10, Y: 10, Z: 4},
			},
		},
		{
			name: "minY=0 with noOffset=false does not shift Y",
			blocks: []model.Block{
				{ID: "minecraft:stone", X: 10, Y: 0, Z: 20},
				{ID: "minecraft:dirt", X: 15, Y: 1, Z: 25},
			},
			noOffset: false,
			wantBlocks: []model.Block{
				{ID: "minecraft:stone", X: 0, Y: 0, Z: 0},
				{ID: "minecraft:dirt", X: 5, Y: 1, Z: 5},
			},
		},
		{
			name: "minX=0 and minZ=0 already - only Y shifts",
			blocks: []model.Block{
				{ID: "minecraft:stone", X: 0, Y: 10, Z: 0},
				{ID: "minecraft:dirt", X: 5, Y: 15, Z: 3},
			},
			noOffset: false,
			wantBlocks: []model.Block{
				{ID: "minecraft:stone", X: 0, Y: 0, Z: 0},
				{ID: "minecraft:dirt", X: 5, Y: 5, Z: 3},
			},
		},
		{
			name: "all coords already zero - no change",
			blocks: []model.Block{
				{ID: "minecraft:bedrock", X: 0, Y: 0, Z: 0},
			},
			noOffset: false,
			wantBlocks: []model.Block{
				{ID: "minecraft:bedrock", X: 0, Y: 0, Z: 0},
			},
		},
		{
			name: "all coords already zero with noOffset=true - no change",
			blocks: []model.Block{
				{ID: "minecraft:bedrock", X: 0, Y: 0, Z: 0},
			},
			noOffset: true,
			wantBlocks: []model.Block{
				{ID: "minecraft:bedrock", X: 0, Y: 0, Z: 0},
			},
		},
		{
			name: "preserves properties map",
			blocks: []model.Block{
				{ID: "minecraft:oak_log", Props: map[string]string{"axis": "y"}, X: 10, Y: 5, Z: 20},
			},
			noOffset: false,
			wantBlocks: []model.Block{
				{ID: "minecraft:oak_log", Props: map[string]string{"axis": "y"}, X: 0, Y: 0, Z: 0},
			},
		},
		{
			name: "large coordinate values",
			blocks: []model.Block{
				{ID: "minecraft:stone", X: 1000000, Y: 500000, Z: 2000000},
				{ID: "minecraft:dirt", X: 1000010, Y: 500005, Z: 2000020},
			},
			noOffset: false,
			wantBlocks: []model.Block{
				{ID: "minecraft:stone", X: 0, Y: 0, Z: 0},
				{ID: "minecraft:dirt", X: 10, Y: 5, Z: 20},
			},
		},
		{
			name: "large negative coordinate values",
			blocks: []model.Block{
				{ID: "minecraft:stone", X: -1000000, Y: -500000, Z: -2000000},
				{ID: "minecraft:dirt", X: -999990, Y: -499995, Z: -1999980},
			},
			noOffset: false,
			wantBlocks: []model.Block{
				{ID: "minecraft:stone", X: 0, Y: 0, Z: 0},
				{ID: "minecraft:dirt", X: 10, Y: 5, Z: 20},
			},
		},
		{
			name: "blocks with different Y values - Y shifts to lowest",
			blocks: []model.Block{
				{ID: "minecraft:stone", X: 10, Y: -10, Z: 20},
				{ID: "minecraft:dirt", X: 12, Y: -5, Z: 25},
				{ID: "minecraft:grass_block", X: 15, Y: 0, Z: 22},
			},
			noOffset: false,
			wantBlocks: []model.Block{
				{ID: "minecraft:stone", X: 0, Y: 0, Z: 0},
				{ID: "minecraft:dirt", X: 2, Y: 5, Z: 5},
				{ID: "minecraft:grass_block", X: 5, Y: 10, Z: 2},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := NewCoordNormalizer()

			result, err := svc.Run(copyBlocks(tt.blocks), tt.noOffset)

			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantBlocks, result)
		})
	}
}

func TestCoordNormalizer_Run_Immutability(t *testing.T) {
	t.Parallel()

	original := []model.Block{
		{ID: "minecraft:stone", X: 10, Y: 5, Z: 20},
		{ID: "minecraft:dirt", X: 15, Y: 8, Z: 25},
	}
	input := copyBlocks(original)

	svc := NewCoordNormalizer()
	_, err := svc.Run(input, false)
	require.NoError(t, err)
	require.Equal(t, original, input, "input blocks must not be modified")
}

func TestCoordNormalizer_Run_ResultNotAliasingInput(t *testing.T) {
	t.Parallel()

	input := []model.Block{
		{ID: "minecraft:stone", X: 10, Y: 5, Z: 20},
	}

	svc := NewCoordNormalizer()
	result, err := svc.Run(input, false)
	require.NoError(t, err)

	result[0].X = 999
	require.Equal(t, 10, input[0].X, "modifying result must not affect input")
}

func TestComputeMinCoords(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		blocks       []model.Block
		wantMinX, wantMinY, wantMinZ int
	}{
		{
			name:   "single block",
			blocks: []model.Block{{X: 10, Y: 5, Z: 20}},
			wantMinX: 10, wantMinY: 5, wantMinZ: 20,
		},
		{
			name: "multiple blocks",
			blocks: []model.Block{
				{X: 10, Y: 5, Z: 20},
				{X: -3, Y: 8, Z: 15},
				{X: 7, Y: -2, Z: 30},
			},
			wantMinX: -3, wantMinY: -2, wantMinZ: 15,
		},
		{
			name: "all negative coords",
			blocks: []model.Block{
				{X: -10, Y: -5, Z: -20},
				{X: -3, Y: -8, Z: -15},
			},
			wantMinX: -10, wantMinY: -8, wantMinZ: -20,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			minX, minY, minZ := computeMinCoords(tt.blocks)
			require.Equal(t, tt.wantMinX, minX)
			require.Equal(t, tt.wantMinY, minY)
			require.Equal(t, tt.wantMinZ, minZ)
		})
	}
}

func TestComputeOffsets(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                 string
		minX, minY, minZ     int
		noOffset             bool
		wantXOff, wantYOff, wantZOff int
	}{
		{
			name:     "positive coords without noOffset",
			minX: 10, minY: 5, minZ: 20,
			noOffset:     false,
			wantXOff: -10, wantYOff: -5, wantZOff: -20,
		},
		{
			name:     "negative coords without noOffset",
			minX: -10, minY: -5, minZ: -20,
			noOffset:     false,
			wantXOff: 10, wantYOff: 5, wantZOff: 20,
		},
		{
			name:     "positive coords with noOffset",
			minX: 10, minY: 5, minZ: 20,
			noOffset:     true,
			wantXOff: -10, wantYOff: 0, wantZOff: -20,
		},
		{
			name:     "negative coords with noOffset",
			minX: -10, minY: -5, minZ: -20,
			noOffset:     true,
			wantXOff: 10, wantYOff: 0, wantZOff: 20,
		},
		{
			name:     "minY=0 without noOffset keeps Y offset zero",
			minX: 10, minY: 0, minZ: 20,
			noOffset:     false,
			wantXOff: -10, wantYOff: 0, wantZOff: -20,
		},
		{
			name:     "minY=0 with noOffset keeps Y offset zero",
			minX: 10, minY: 0, minZ: 20,
			noOffset:     true,
			wantXOff: -10, wantYOff: 0, wantZOff: -20,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			xOff, yOff, zOff := computeOffsets(tt.minX, tt.minY, tt.minZ, tt.noOffset)
			require.Equal(t, tt.wantXOff, xOff)
			require.Equal(t, tt.wantYOff, yOff)
			require.Equal(t, tt.wantZOff, zOff)
		})
	}
}

func TestApplyOffset(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		blocks     []model.Block
		xOff, yOff, zOff int
		want       []model.Block
	}{
		{
			name:   "zero offset",
			blocks: []model.Block{{ID: "stone", X: 10, Y: 5, Z: 20}},
			want:   []model.Block{{ID: "stone", X: 10, Y: 5, Z: 20}},
		},
		{
			name:   "positive offset",
			blocks: []model.Block{{ID: "stone", X: 10, Y: 5, Z: 20}},
			xOff: 5, yOff: 3, zOff: 7,
			want: []model.Block{{ID: "stone", X: 15, Y: 8, Z: 27}},
		},
		{
			name:   "negative offset",
			blocks: []model.Block{{ID: "stone", X: 10, Y: 5, Z: 20}},
			xOff: -10, yOff: -5, zOff: -20,
			want: []model.Block{{ID: "stone", X: 0, Y: 0, Z: 0}},
		},
		{
			name: "multiple blocks",
			blocks: []model.Block{
				{ID: "stone", X: 10, Y: 5, Z: 20},
				{ID: "dirt", X: 15, Y: 8, Z: 25},
			},
			xOff: -10, yOff: -5, zOff: -20,
			want: []model.Block{
				{ID: "stone", X: 0, Y: 0, Z: 0},
				{ID: "dirt", X: 5, Y: 3, Z: 5},
			},
		},
		{
			name: "preserves props",
			blocks: []model.Block{
				{ID: "oak_log", Props: map[string]string{"axis": "y"}, X: 10, Y: 5, Z: 20},
			},
			xOff: -10, yOff: -5, zOff: -20,
			want: []model.Block{
				{ID: "oak_log", Props: map[string]string{"axis": "y"}, X: 0, Y: 0, Z: 0},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := applyOffset(tt.blocks, tt.xOff, tt.yOff, tt.zOff)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestApplyOffset_Immutability(t *testing.T) {
	t.Parallel()

	original := []model.Block{
		{ID: "stone", X: 10, Y: 5, Z: 20},
	}
	input := make([]model.Block, len(original))
	copy(input, original)

	_ = applyOffset(input, -10, -5, -20)
	require.Equal(t, original, input, "input must not be modified")
}

func copyBlocks(blocks []model.Block) []model.Block {
	if blocks == nil {
		return nil
	}
	out := make([]model.Block, len(blocks))
	copy(out, blocks)
	return out
}
