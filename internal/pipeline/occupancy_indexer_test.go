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

	si := stateful.NewStyleIndex()
	si.Add("minecraft:stone", "", model.StyledBlock{
		ID: "minecraft:stone", PropsKey: "", GridAlignment: model.GridFullBlock,
	})
	si.Add("minecraft:flower", "", model.StyledBlock{
		ID: "minecraft:flower", PropsKey: "", GridAlignment: model.GridSubBlock,
	})
	si.Add("minecraft:stairs", "", model.StyledBlock{
		ID: "minecraft:stairs", PropsKey: "", GridAlignment: model.GridNotAligned,
		Elements: []model.StyledElement{
			{From: model.Vector3{0, 0, 0}, To: model.Vector3{16, 8, 16}, Shade: true},
			{From: model.Vector3{0, 8, 0}, To: model.Vector3{16, 16, 16}, Shade: true},
		},
	})

	occ := NewOccupancyIndexer(pkb).Run(blocks, blockIdx, microIdx, *si)

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

func TestOccupancyIndexer_ComplexBlock_PartialOccupancy(t *testing.T) {
	t.Parallel()

	pkb := &mockMergerPropsKeyBuilder{
		runFn: func(props map[string]string) string {
			return ""
		},
	}

	blockIdx := stateful.NewVoxelIndex()
	blockIdx.AddBlock(&model.MergedBlock{ID: "minecraft:stone", PropsKey: "", X: 0, Y: 0, Z: 0})

	blocks := []model.RawBlock{
		{ID: "minecraft:fence", X: 1, Y: 0, Z: 0},
	}

	si := stateful.NewStyleIndex()
	si.Add("minecraft:stone", "", model.StyledBlock{
		ID: "minecraft:stone", PropsKey: "", GridAlignment: model.GridFullBlock,
	})
	si.Add("minecraft:fence", "", model.StyledBlock{
		ID: "minecraft:fence", PropsKey: "", GridAlignment: model.GridNotAligned,
		Elements: []model.StyledElement{
			{From: model.Vector3{6, 0, 6}, To: model.Vector3{10, 16, 10}, Shade: true},
		},
	})

	occ := NewOccupancyIndexer(pkb).Run(blocks, blockIdx, stateful.NewVoxelIndex(), *si)

	require.True(t, occ.Occluding(0, 0, 0))
	require.True(t, occ.Occluding(3, 3, 3))

	fenceXFrom := 1*model.SubGridSize + 6/model.SubGridSize
	fenceXTo := 1*model.SubGridSize + 10/model.SubGridSize
	fenceYTo := 16 / model.SubGridSize
	fenceZFrom := 0*model.SubGridSize + 6/model.SubGridSize
	fenceZTo := 0*model.SubGridSize + 10/model.SubGridSize
	require.True(t, occ.Occluding(fenceXFrom, 0, fenceZFrom))
	require.True(t, occ.Occluding(fenceXTo-1, fenceYTo-1, fenceZTo-1))
	require.False(t, occ.Occluding(fenceXFrom, fenceYTo, fenceZFrom))
	require.False(t, occ.Occluding(fenceXTo, 0, fenceZFrom))

	require.False(t, occ.Occluding(4, 0, 0))
	require.False(t, occ.Occluding(7, 3, 3))
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

func TestOccupancyIndexer_Run_SkipsBlockWithoutStyle(t *testing.T) {
	t.Parallel()

	pkb := &mockMergerPropsKeyBuilder{
		runFn: func(props map[string]string) string {
			return ""
		},
	}

	blockIdx := stateful.NewVoxelIndex()
	blockIdx.AddBlock(&model.MergedBlock{ID: "minecraft:unknown", X: 0, Y: 0, Z: 0})

	microIdx := stateful.NewVoxelIndex()
	microIdx.AddBlock(&model.MergedBlock{ID: "minecraft:unknown_micro", X: 6, Y: 7, Z: 8})

	styles := styleIndex(struct {
		id        string
		prop      string
		alignment model.GridAlignment
	}{"minecraft:stone", "", model.GridFullBlock})

	occ := NewOccupancyIndexer(pkb).Run(nil, blockIdx, microIdx, styles)

	require.False(t, occ.Occupied(0, 0, 0))
	require.False(t, occ.Occluding(6, 7, 8))
}
