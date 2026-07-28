package model

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBounds_ZeroValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                string
		bounds              Bounds
		wantXMin, wantXMax  int
		wantYMin, wantYMax  int
		wantZMin, wantZMax  int
	}{
		{
			name: "zero value",
			bounds: Bounds{},
		},
		{
			name: "set values",
			bounds: Bounds{
				XMin: -10, XMax: 10,
				YMin: 0, YMax: 255,
				ZMin: -100, ZMax: 100,
			},
			wantXMin: -10, wantXMax: 10,
			wantYMin: 0, wantYMax: 255,
			wantZMin: -100, wantZMax: 100,
		},
		{
			name: "negative bounds",
			bounds: Bounds{
				XMin: -100, XMax: -50,
				YMin: -64, YMax: -32,
				ZMin: -200, ZMax: -150,
			},
			wantXMin: -100, wantXMax: -50,
			wantYMin: -64, wantYMax: -32,
			wantZMin: -200, wantZMax: -150,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.wantXMin, tt.bounds.XMin)
			require.Equal(t, tt.wantXMax, tt.bounds.XMax)
			require.Equal(t, tt.wantYMin, tt.bounds.YMin)
			require.Equal(t, tt.wantYMax, tt.bounds.YMax)
			require.Equal(t, tt.wantZMin, tt.bounds.ZMin)
			require.Equal(t, tt.wantZMax, tt.bounds.ZMax)
		})
	}
}

func TestBlock_Properties(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		block Block
	}{
		{
			name: "nil properties",
			block: Block{
				ID: "minecraft:stone",
				X:  10, Y: 20, Z: 30,
			},
		},
		{
			name: "empty properties map",
			block: Block{
				ID:         "minecraft:stone",
				Properties: map[string]string{},
				X:          10, Y: 20, Z: 30,
			},
		},
		{
			name: "with properties",
			block: Block{
				ID:         "minecraft:stairs",
				Properties: map[string]string{"facing": "north"},
				X:          10, Y: 20, Z: 30,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.block.X, 10)
			require.Equal(t, tt.block.Y, 20)
			require.Equal(t, tt.block.Z, 30)
		})
	}
}

func TestWorld_Lookup(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		blocks []Block
		want   map[[3]int]*Block
	}{
		{
			name:   "empty world",
			blocks: nil,
			want:   nil,
		},
		{
			name: "single block",
			blocks: []Block{
				{ID: "minecraft:stone", X: 10, Y: 20, Z: 30},
			},
			want: map[[3]int]*Block{
				{10, 20, 30}: {ID: "minecraft:stone", X: 10, Y: 20, Z: 30},
			},
		},
		{
			name: "multiple blocks at different positions",
			blocks: []Block{
				{ID: "minecraft:stone", X: 0, Y: 0, Z: 0},
				{ID: "minecraft:dirt", X: 1, Y: 0, Z: 0},
			},
			want: map[[3]int]*Block{
				{0, 0, 0}: {ID: "minecraft:stone", X: 0, Y: 0, Z: 0},
				{1, 0, 0}: {ID: "minecraft:dirt", X: 1, Y: 0, Z: 0},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if len(tt.blocks) == 0 {
				w := &World{}
				require.Nil(t, w.Lookup)
				return
			}

			lookup := make(map[[3]int]*Block, len(tt.blocks))
			for i := range tt.blocks {
				b := &tt.blocks[i]
				lookup[[3]int{b.X, b.Y, b.Z}] = b
			}
			w := &World{Blocks: tt.blocks, Lookup: lookup}

			for pos, expectedBlock := range tt.want {
				got, ok := w.Lookup[pos]
				require.True(t, ok, "expected block at position %v", pos)
				require.Equal(t, expectedBlock.ID, got.ID)
				require.Equal(t, expectedBlock.X, got.X)
				require.Equal(t, expectedBlock.Y, got.Y)
				require.Equal(t, expectedBlock.Z, got.Z)
			}
		})
	}
}
