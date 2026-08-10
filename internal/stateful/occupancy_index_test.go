package stateful

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOccupancyIndex_New(t *testing.T) {
	t.Parallel()

	idx := NewOccupancyIndex()
	require.NotNil(t, idx)
	require.Zero(t, idx.Len())
}

func TestOccupancyIndex_FillCell(t *testing.T) {
	t.Parallel()

	idx := NewOccupancyIndex()
	idx.FillCell(1, 2, 3, true)

	require.True(t, idx.Occupied(1, 2, 3))
	require.True(t, idx.Occluding(1, 2, 3))
	require.False(t, idx.Occupied(0, 0, 0))
	require.False(t, idx.Occluding(0, 0, 0))
	require.False(t, idx.Occluding(4, 0, 0))
}

func TestOccupancyIndex_FillCellTransparent(t *testing.T) {
	t.Parallel()

	idx := NewOccupancyIndex()
	idx.FillCell(1, 2, 3, false)

	require.True(t, idx.Occupied(1, 2, 3))
	require.False(t, idx.Occluding(1, 2, 3))
}

func TestOccupancyIndex_FillRegion(t *testing.T) {
	t.Parallel()

	idx := NewOccupancyIndex()
	idx.FillRegion(0, 0, 0, 2, 3, 4, true)

	require.Equal(t, 2*3*4, idx.Len())
	require.True(t, idx.Occupied(1, 2, 3))
	require.True(t, idx.Occluding(1, 2, 3))
	require.False(t, idx.Occupied(2, 0, 0))
	require.False(t, idx.Occupied(0, 3, 0))
	require.False(t, idx.Occluding(2, 2, 2))
}

func TestOccupancyIndex_LenCountsUniqueCells(t *testing.T) {
	t.Parallel()

	idx := NewOccupancyIndex()
	idx.FillCell(0, 0, 0, true)
	idx.FillCell(0, 0, 0, false)
	idx.FillRegion(5, 5, 5, 2, 2, 2, true)

	require.Equal(t, 1+2*2*2, idx.Len())
}
