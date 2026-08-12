package stateful

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStress_OccupancyIndex(t *testing.T) {
	const size = 32

	occ := NewOccupancyIndex()
	occ.FillRegion(0, 0, 0, size, size, size, true)

	require.Equal(t, size*size*size, occ.Len())
	require.True(t, occ.Occupied(0, 0, 0))
	require.True(t, occ.Occluding(size-1, size-1, size-1))
}
