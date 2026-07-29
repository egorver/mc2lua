package index

import (
	"mc2lua/internal/model"
	"testing"

	"github.com/stretchr/testify/require"
)

func makeTestVoxelIndex(blocks ...*model.MergedBlock) *VoxelIndex {
	g := NewVoxelIndex()
	for i := range blocks {
		g.AddBlock(blocks[i])
	}
	return g
}

func TestVoxelIndexBlocks(t *testing.T) {
	g := makeTestVoxelIndex(&model.MergedBlock{ID: "a", X: 0, Y: 0, Z: 0})
	blocks := g.Blocks()
	require.Len(t, blocks, 1)
	require.Equal(t, "a", blocks[0].ID)
}

func TestAddBlockAndGetBlock(t *testing.T) {
	g := NewVoxelIndex()
	b := &model.MergedBlock{ID: "test", X: 1, Y: 2, Z: 3}
	g.AddBlock(b)

	got := g.GetBlock(1, 2, 3)
	require.NotNil(t, got)
	require.Equal(t, "test", got.ID)

	require.Nil(t, g.GetBlock(0, 0, 0))
}
