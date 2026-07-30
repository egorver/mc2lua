package model

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCuboid_ZeroValue(t *testing.T) {
	t.Parallel()

	var c Cuboid
	require.Equal(t, "", c.ID)
	require.Equal(t, "", c.PropsKey)
	require.Equal(t, 0.0, c.X)
	require.Equal(t, 0.0, c.Y)
	require.Equal(t, 0.0, c.Z)
	require.Equal(t, 0, c.Width)
	require.Equal(t, 0, c.Depth)
	require.Equal(t, 0, c.Height)
}

func TestCuboid_FullInit(t *testing.T) {
	t.Parallel()

	c := Cuboid{
		ID:       "stone",
		PropsKey: "variant=andesite",
		X:        10.5,
		Y:        20.5,
		Z:        30.5,
		Width:    5,
		Depth:    10,
		Height:   15,
	}
	require.Equal(t, "stone", c.ID)
	require.Equal(t, "variant=andesite", c.PropsKey)
	require.Equal(t, 10.5, c.X)
	require.Equal(t, 20.5, c.Y)
	require.Equal(t, 30.5, c.Z)
	require.Equal(t, 5, c.Width)
	require.Equal(t, 10, c.Depth)
	require.Equal(t, 15, c.Height)
}

func TestCuboid_NegativeCoordinates(t *testing.T) {
	t.Parallel()

	c := Cuboid{X: -5, Y: -3, Z: -1, Width: 2, Depth: 2, Height: 2}
	require.Equal(t, -5.0, c.X)
	require.Equal(t, -3.0, c.Y)
	require.Equal(t, -1.0, c.Z)
	require.Equal(t, 2, c.Width)
	require.Equal(t, 2, c.Depth)
	require.Equal(t, 2, c.Height)
}
