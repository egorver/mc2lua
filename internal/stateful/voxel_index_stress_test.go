package stateful

import (
	"testing"

	"mc2lua/internal/model"

	"github.com/stretchr/testify/require"
)

func TestStress_VoxelIndex(t *testing.T) {
	const total = 1_000_000
	const spanX = 128
	const spanY = 64

	grid := NewVoxelIndex()
	for i := 0; i < total; i++ {
		grid.AddBlock(&model.MergedBlock{
			ID: "minecraft:stone",
			X:  i % spanX,
			Y:  (i / spanX) % spanY,
			Z:  i / (spanX * spanY),
		})
	}

	require.Len(t, grid.Blocks(), total)

	for i := 0; i < total; i += 100_000 {
		x := i % spanX
		y := (i / spanX) % spanY
		z := i / (spanX * spanY)
		b := grid.GetBlock(x, y, z)
		require.NotNil(t, b)
		require.Equal(t, "minecraft:stone", b.ID)
	}
}
