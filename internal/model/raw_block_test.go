package model

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRawBlock_ZeroValue(t *testing.T) {
	t.Parallel()

	var b RawBlock
	require.Equal(t, "", b.ID)
	require.Nil(t, b.Props)
	require.Equal(t, 0, b.X)
	require.Equal(t, 0, b.Y)
	require.Equal(t, 0, b.Z)
}

func TestRawBlock_FullInit(t *testing.T) {
	t.Parallel()

	b := RawBlock{
		ID:    "minecraft:stone",
		Props: map[string]string{"variant": "andesite"},
		X:     10,
		Y:     -5,
		Z:     20,
	}
	require.Equal(t, "minecraft:stone", b.ID)
	require.Equal(t, map[string]string{"variant": "andesite"}, b.Props)
	require.Equal(t, 10, b.X)
	require.Equal(t, -5, b.Y)
	require.Equal(t, 20, b.Z)
}

func TestRawBlock_PartialInit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		block    RawBlock
		wantID   string
		wantNil  bool
		wantX    int
		wantY    int
		wantZ    int
	}{
		{
			name:   "only ID",
			block:  RawBlock{ID: "minecraft:dirt"},
			wantID: "minecraft:dirt",
			wantNil: true,
		},
		{
			name:   "only coords",
			block:  RawBlock{X: 1, Y: 2, Z: 3},
			wantX:  1,
			wantY:  2,
			wantZ:  3,
			wantNil: true,
		},
		{
			name: "with props and coords",
			block: RawBlock{
				ID:    "minecraft:oak_log",
				Props: map[string]string{"axis": "y"},
				X:     5,
				Y:     10,
				Z:     15,
			},
			wantID:  "minecraft:oak_log",
			wantX:   5,
			wantY:   10,
			wantZ:   15,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.wantID, tt.block.ID)
			require.Equal(t, tt.wantX, tt.block.X)
			require.Equal(t, tt.wantY, tt.block.Y)
			require.Equal(t, tt.wantZ, tt.block.Z)
			if tt.wantNil {
				require.Nil(t, tt.block.Props)
			}
		})
	}
}
