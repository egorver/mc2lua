package pipeline

import (
	"testing"

	"mc2lua/internal/model"
	"mc2lua/internal/stateful"

	"github.com/stretchr/testify/require"
)

func TestOccupancyIndexer_New(t *testing.T) {
	t.Parallel()

	pkb := &mockMergerPropsKeyBuilder{}
	svc := NewOccupancyIndexer(pkb)
	require.NotNil(t, svc)
}

func TestOccupancyIndexer_Run(t *testing.T) {
	t.Parallel()

	pkb := &mockMergerPropsKeyBuilder{
		runFn: func(props map[string]string) string {
			return ""
		},
	}

	blockIdx := stateful.NewVoxelIndex()
	blockIdx.AddBlock(&model.MergedBlock{ID: "minecraft:stone", PropsKey: "", X: 0, Y: 0, Z: 0})

	microIdx := stateful.NewVoxelIndex()
	microIdx.AddBlock(&model.MergedBlock{ID: "minecraft:flower", PropsKey: "", X: 6, Y: 7, Z: 8})

	blocks := []model.RawBlock{
		{ID: "minecraft:stairs", X: 2, Y: 0, Z: 0},
		{ID: "minecraft:stone", X: 10, Y: 0, Z: 0},
	}

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
		}{"minecraft:flower", "", model.GridSubBlock},
		struct {
			id        string
			prop      string
			alignment model.GridAlignment
		}{"minecraft:stairs", "", model.GridNotAligned},
	)

	occ := NewOccupancyIndexer(pkb).Run(blocks, blockIdx, microIdx, styles)

	require.True(t, occ.Occluding(0, 0, 0))
	require.True(t, occ.Occluding(3, 3, 3))
	require.False(t, occ.Occluding(4, 0, 0))

	require.True(t, occ.Occluding(6, 7, 8))
	require.False(t, occ.Occluding(5, 7, 8))

	require.True(t, occ.Occluding(8, 0, 0))
	require.True(t, occ.Occluding(11, 3, 3))
	require.False(t, occ.Occluding(12, 0, 0))

	require.False(t, occ.Occluding(40, 0, 0))
}

func TestOccupancyIndexer_RunTransparent(t *testing.T) {
	t.Parallel()

	pkb := &mockMergerPropsKeyBuilder{
		runFn: func(props map[string]string) string {
			return ""
		},
	}

	blockIdx := stateful.NewVoxelIndex()
	blockIdx.AddBlock(&model.MergedBlock{ID: "minecraft:glass", PropsKey: "", X: 0, Y: 0, Z: 0})

	styles := stateful.NewStyleIndex()
	styles.Add("minecraft:glass", "", model.StyledBlock{
		ID:            "minecraft:glass",
		PropsKey:      "",
		GridAlignment: model.GridFullBlock,
		Transparent:   true,
	})

	occ := NewOccupancyIndexer(pkb).Run(nil, blockIdx, stateful.NewVoxelIndex(), *styles)

	require.True(t, occ.Occupied(0, 0, 0))
	require.False(t, occ.Occluding(0, 0, 0))
}
