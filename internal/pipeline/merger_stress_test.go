package pipeline

import (
	"math/rand"
	"testing"

	"mc2lua/internal/model"

	"github.com/stretchr/testify/require"
)

func TestStress_RegionMerger_DenseCube(t *testing.T) {
	const size = 64

	blocks := make([]model.RawBlock, 0, size*size*size)
	for x := 0; x < size; x++ {
		for y := 0; y < size; y++ {
			for z := 0; z < size; z++ {
				blocks = append(blocks, fullBlockAt("minecraft:stone", "", x, y, z))
			}
		}
	}

	styles := styleIndex(struct {
		id        string
		prop      string
		alignment model.GridAlignment
	}{"minecraft:stone", "", model.GridFullBlock})

	indexer := NewBlockVoxelIndexer(&mockMergerPropsKeyBuilder{})
	grid := indexer.Run(blocks, styles)
	svc := NewRegionMerger(NewCuboidHelper())
	cuboids := svc.Run(grid)

	require.Len(t, cuboids, 1)
	require.Equal(t, size*size*size, totalVolume(cuboids))
	require.Equal(t, size, cuboids[0].Width)
	require.Equal(t, size, cuboids[0].Depth)
	require.Equal(t, size, cuboids[0].Height)
}

func TestStress_RegionMerger_SparseWorld(t *testing.T) {
	const (
		blockCount = 20000
		worldX     = 128
		worldY     = 64
		worldZ     = 128
	)

	ids := []string{"minecraft:stone", "minecraft:dirt", "minecraft:grass_block"}
	styles := styleIndex(
		struct {
			id        string
			prop      string
			alignment model.GridAlignment
		}{"minecraft:stone", "", model.GridFullBlock},
		struct {
			id        string
			prop      string
			alignment model.GridAlignment
		}{"minecraft:dirt", "", model.GridFullBlock},
		struct {
			id        string
			prop      string
			alignment model.GridAlignment
		}{"minecraft:grass_block", "", model.GridFullBlock},
	)

	rng := rand.New(rand.NewSource(42))
	used := make(map[[3]int]bool, blockCount)
	blocks := make([]model.RawBlock, 0, blockCount)
	for len(blocks) < blockCount {
		pos := [3]int{rng.Intn(worldX), rng.Intn(worldY), rng.Intn(worldZ)}
		if used[pos] {
			continue
		}
		used[pos] = true
		blocks = append(blocks, fullBlockAt(ids[len(blocks)%len(ids)], "", pos[0], pos[1], pos[2]))
	}

	indexer := NewBlockVoxelIndexer(&mockMergerPropsKeyBuilder{})
	grid := indexer.Run(blocks, styles)
	svc := NewRegionMerger(NewCuboidHelper())
	cuboids := svc.Run(grid)

	require.Equal(t, blockCount, totalVolume(cuboids))
	require.Zero(t, countOverlaps(collectVoxels(cuboids)))
	require.Equal(t, len(grid.Blocks()), blockCount)
}
