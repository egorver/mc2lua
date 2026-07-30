package model

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRect2D_ZeroValue(t *testing.T) {
	t.Parallel()

	var r Rect2D
	require.Equal(t, 0, r.X)
	require.Equal(t, 0, r.Z)
	require.Equal(t, 0, r.Width)
	require.Equal(t, 0, r.Depth)
}

func TestRect2D_FullInit(t *testing.T) {
	t.Parallel()

	r := Rect2D{X: 5, Z: 10, Width: 15, Depth: 20}
	require.Equal(t, 5, r.X)
	require.Equal(t, 10, r.Z)
	require.Equal(t, 15, r.Width)
	require.Equal(t, 20, r.Depth)
}

func TestRect2D_NegativeCoords(t *testing.T) {
	t.Parallel()

	r := Rect2D{X: -5, Z: -10, Width: 3, Depth: 4}
	require.Equal(t, -5, r.X)
	require.Equal(t, -10, r.Z)
	require.Equal(t, 3, r.Width)
	require.Equal(t, 4, r.Depth)
}

func TestRect2D_UnitRect(t *testing.T) {
	t.Parallel()

	r := Rect2D{X: 0, Z: 0, Width: 1, Depth: 1}
	require.Equal(t, 0, r.X)
	require.Equal(t, 0, r.Z)
	require.Equal(t, 1, r.Width)
	require.Equal(t, 1, r.Depth)
}
