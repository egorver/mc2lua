package model

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMergedBlock_ZeroValue(t *testing.T) {
	t.Parallel()

	var b MergedBlock
	require.Equal(t, "", b.ID)
	require.Equal(t, "", b.PropsKey)
	require.Equal(t, 0, b.X)
	require.Equal(t, 0, b.Y)
	require.Equal(t, 0, b.Z)
	require.False(t, b.Merged)
}

func TestMergedBlock_FullInit(t *testing.T) {
	t.Parallel()

	b := MergedBlock{
		ID:       "minecraft:stone",
		PropsKey: "variant=andesite",
		X:        10,
		Y:        20,
		Z:        30,
		Merged:   true,
	}
	require.Equal(t, "minecraft:stone", b.ID)
	require.Equal(t, "variant=andesite", b.PropsKey)
	require.Equal(t, 10, b.X)
	require.Equal(t, 20, b.Y)
	require.Equal(t, 30, b.Z)
	require.True(t, b.Merged)
}

func TestMergedBlock_NegativeCoords(t *testing.T) {
	t.Parallel()

	b := MergedBlock{X: -5, Y: -10, Z: -3, Merged: false}
	require.Equal(t, -5, b.X)
	require.Equal(t, -10, b.Y)
	require.Equal(t, -3, b.Z)
	require.False(t, b.Merged)
}

func TestMergedBlock_OnlyID(t *testing.T) {
	t.Parallel()

	b := MergedBlock{ID: "minecraft:air"}
	require.Equal(t, "minecraft:air", b.ID)
	require.Equal(t, "", b.PropsKey)
	require.False(t, b.Merged)
}
