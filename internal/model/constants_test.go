package model

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConstants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		got  any
		want any
	}{
		{name: "full block size", got: FullBlockSize, want: 16.0},
		{name: "block center", got: BlockCenter, want: 8.0},
		{name: "sub grid size", got: SubGridSize, want: 4},
		{name: "block dimensions", got: BlockDimensions, want: 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.want, tt.got)
		})
	}
}

func TestConstantsTyped(t *testing.T) {
	t.Parallel()

	require.IsType(t, float64(0), FullBlockSize)
	require.IsType(t, float64(0), BlockCenter)
	require.IsType(t, 0, SubGridSize)
	require.IsType(t, 0, BlockDimensions)
}

func TestConstantsDerived(t *testing.T) {
	t.Parallel()

	require.Equal(t, FullBlockSize/2, BlockCenter)
	require.Equal(t, FullBlockSize/SubGridSize, 4.0)
	require.Greater(t, FullBlockSize, BlockCenter)
}
