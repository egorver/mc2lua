package pipeline

import (
	"testing"

	"mc2lua/internal/model"
	"mc2lua/internal/stateful"

	"github.com/stretchr/testify/require"
)

func TestFaceCuller_New(t *testing.T) {
	t.Parallel()

	pkb := &mockMergerPropsKeyBuilder{}
	svc := NewFaceCuller(pkb)
	require.NotNil(t, svc)
}

func TestFaceCuller_TwoAdjacentFullCuboids(t *testing.T) {
	t.Parallel()

	occ := stateful.NewOccupancyIndex()
	occ.FillRegion(0, 0, 0, 4, 4, 4, true)
	occ.FillRegion(4, 0, 0, 4, 4, 4, true)

	blockRegions := []model.Cuboid{
		{ID: "minecraft:stone", X: 0, Y: 0, Z: 0, Width: 1, Height: 1, Depth: 1},
		{ID: "minecraft:stone", X: 1, Y: 0, Z: 0, Width: 1, Height: 1, Depth: 1},
	}

	vis := NewFaceCuller(&mockMergerPropsKeyBuilder{}).Run(occ, blockRegions, nil, nil, styleIndex())

	require.Equal(t, model.FaceMask{
		model.FaceIndexTop:    true,
		model.FaceIndexBottom: true,
		model.FaceIndexFront:  true,
		model.FaceIndexBack:   true,
		model.FaceIndexLeft:   true,
		model.FaceIndexRight:  false,
	}, vis.BlockFaces[0])

	require.Equal(t, model.FaceMask{
		model.FaceIndexTop:    true,
		model.FaceIndexBottom: true,
		model.FaceIndexFront:  true,
		model.FaceIndexBack:   true,
		model.FaceIndexLeft:   false,
		model.FaceIndexRight:  true,
	}, vis.BlockFaces[1])
}

func TestFaceCuller_FullCuboidPartiallyCoveredByMicro(t *testing.T) {
	t.Parallel()

	occ := stateful.NewOccupancyIndex()
	occ.FillRegion(0, 0, 0, 4, 4, 4, true)
	occ.FillRegion(4, 0, 0, 2, 2, 4, true)

	blockRegions := []model.Cuboid{
		{ID: "minecraft:stone", X: 0, Y: 0, Z: 0, Width: 1, Height: 1, Depth: 1},
	}
	microRegions := []model.Cuboid{
		{ID: "minecraft:flower", X: 4.5, Y: 0.5, Z: 1.5, Width: 2, Height: 2, Depth: 4},
	}

	vis := NewFaceCuller(&mockMergerPropsKeyBuilder{}).Run(occ, blockRegions, microRegions, nil, styleIndex())

	require.Equal(t, model.FaceMask{
		model.FaceIndexTop:    true,
		model.FaceIndexBottom: true,
		model.FaceIndexFront:  true,
		model.FaceIndexBack:   true,
		model.FaceIndexLeft:   true,
		model.FaceIndexRight:  true,
	}, vis.BlockFaces[0])
}

func TestFaceCuller_FullCuboidFullyCoveredByMicro(t *testing.T) {
	t.Parallel()

	occ := stateful.NewOccupancyIndex()
	occ.FillRegion(0, 0, 0, 4, 4, 4, true)
	occ.FillRegion(4, 0, 0, 4, 4, 4, true)

	blockRegions := []model.Cuboid{
		{ID: "minecraft:stone", X: 0, Y: 0, Z: 0, Width: 1, Height: 1, Depth: 1},
	}

	vis := NewFaceCuller(&mockMergerPropsKeyBuilder{}).Run(occ, blockRegions, nil, nil, styleIndex())

	require.Equal(t, model.FaceMask{
		model.FaceIndexTop:    true,
		model.FaceIndexBottom: true,
		model.FaceIndexFront:  true,
		model.FaceIndexBack:   true,
		model.FaceIndexLeft:   true,
		model.FaceIndexRight:  false,
	}, vis.BlockFaces[0])
}

func TestFaceCuller_ComplexBlockAtWorldEdge(t *testing.T) {
	t.Parallel()

	occ := stateful.NewOccupancyIndex()
	blocks := []model.RawBlock{
		{ID: "minecraft:stairs", X: 0, Y: 0, Z: 0},
	}
	styles := styleIndex(struct {
		id        string
		prop      string
		alignment model.GridAlignment
	}{"minecraft:stairs", "", model.GridNotAligned})

	vis := NewFaceCuller(&mockMergerPropsKeyBuilder{}).Run(occ, nil, nil, blocks, styles)

	require.Equal(t, model.FaceMask{
		model.FaceIndexTop:    true,
		model.FaceIndexBottom: true,
		model.FaceIndexFront:  true,
		model.FaceIndexBack:   true,
		model.FaceIndexLeft:   true,
		model.FaceIndexRight:  true,
	}, vis.ComplexFaces[0])
}

func TestFaceCuller_ComplexBlockSurrounded(t *testing.T) {
	t.Parallel()

	occ := stateful.NewOccupancyIndex()
	occ.FillRegion(-4, -4, -4, 12, 12, 12, true)
	blocks := []model.RawBlock{
		{ID: "minecraft:stairs", X: 0, Y: 0, Z: 0},
	}
	styles := styleIndex(struct {
		id        string
		prop      string
		alignment model.GridAlignment
	}{"minecraft:stairs", "", model.GridNotAligned})

	vis := NewFaceCuller(&mockMergerPropsKeyBuilder{}).Run(occ, nil, nil, blocks, styles)

	require.Equal(t, model.FaceMask{
		model.FaceIndexTop:    false,
		model.FaceIndexBottom: false,
		model.FaceIndexFront:  false,
		model.FaceIndexBack:   false,
		model.FaceIndexLeft:   false,
		model.FaceIndexRight:  false,
	}, vis.ComplexFaces[0])
}

func TestFaceCuller_TransparentNeighborDoesNotHideFace(t *testing.T) {
	t.Parallel()

	occ := stateful.NewOccupancyIndex()
	occ.FillRegion(0, 0, 0, 4, 4, 4, true)
	occ.FillRegion(4, 0, 0, 4, 4, 4, false)

	blockRegions := []model.Cuboid{
		{ID: "minecraft:stone", X: 0, Y: 0, Z: 0, Width: 1, Height: 1, Depth: 1},
	}

	vis := NewFaceCuller(&mockMergerPropsKeyBuilder{}).Run(occ, blockRegions, nil, nil, styleIndex())

	require.Equal(t, model.FaceMask{
		model.FaceIndexTop:    true,
		model.FaceIndexBottom: true,
		model.FaceIndexFront:  true,
		model.FaceIndexBack:   true,
		model.FaceIndexLeft:   true,
		model.FaceIndexRight:  true,
	}, vis.BlockFaces[0])
}
