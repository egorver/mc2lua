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
	svc := NewFaceCuller(pkb, NewCuboidHelper())
	require.NotNil(t, svc)
}

func faceMask(top, bottom, front, back, left, right bool) model.FaceMask {
	return model.FaceMask{top, bottom, front, back, left, right}
}

func twoAdjacentFullCuboidsOcc() *stateful.OccupancyIndex {
	occ := stateful.NewOccupancyIndex()
	occ.FillRegion(0, 0, 0, 4, 4, 4, true)
	occ.FillRegion(4, 0, 0, 4, 4, 4, true)
	return occ
}

func TestFaceCuller_Run(t *testing.T) {
	t.Parallel()

	stairsStyle := styleIndex(struct {
		id        string
		prop      string
		alignment model.GridAlignment
	}{"minecraft:stairs", "", model.GridNotAligned})

	emptyMasks := []model.FaceMask{}

	tests := []struct {
		name         string
		setupOcc     func() *stateful.OccupancyIndex
		blockRegions []model.Cuboid
		microRegions []model.Cuboid
		blocks       []model.RawBlock
		styles       stateful.StyleIndex
		wantBlock    []model.FaceMask
		wantMicro    []model.FaceMask
		wantComplex  []model.FaceMask
	}{
		{
			name:     "two adjacent full cuboids",
			setupOcc: twoAdjacentFullCuboidsOcc,
			blockRegions: []model.Cuboid{
				{ID: "minecraft:stone", X: 0, Y: 0, Z: 0, Width: 1, Height: 1, Depth: 1},
				{ID: "minecraft:stone", X: 1, Y: 0, Z: 0, Width: 1, Height: 1, Depth: 1},
			},
			styles:      styleIndex(),
			wantBlock:   []model.FaceMask{faceMask(true, true, true, true, true, false), faceMask(true, true, true, true, false, true)},
			wantMicro:   emptyMasks,
			wantComplex: emptyMasks,
		},
		{
			name: "full cuboid partially covered by micro",
			setupOcc: func() *stateful.OccupancyIndex {
				occ := stateful.NewOccupancyIndex()
				occ.FillRegion(0, 0, 0, 4, 4, 4, true)
				occ.FillRegion(4, 0, 0, 2, 2, 4, true)
				return occ
			},
			blockRegions: []model.Cuboid{
				{ID: "minecraft:stone", X: 0, Y: 0, Z: 0, Width: 1, Height: 1, Depth: 1},
			},
			microRegions: []model.Cuboid{
				{ID: "minecraft:flower", X: 4.5, Y: 0.5, Z: 1.5, Width: 2, Height: 2, Depth: 4},
			},
			styles:      styleIndex(),
			wantBlock:   []model.FaceMask{faceMask(true, true, true, true, true, true)},
			wantMicro:   []model.FaceMask{faceMask(true, true, true, true, false, true)},
			wantComplex: emptyMasks,
		},
		{
			name:     "full cuboid fully covered by micro",
			setupOcc: twoAdjacentFullCuboidsOcc,
			blockRegions: []model.Cuboid{
				{ID: "minecraft:stone", X: 0, Y: 0, Z: 0, Width: 1, Height: 1, Depth: 1},
			},
			styles:      styleIndex(),
			wantBlock:   []model.FaceMask{faceMask(true, true, true, true, true, false)},
			wantMicro:   emptyMasks,
			wantComplex: emptyMasks,
		},
		{
			name:        "complex block at world edge",
			blocks:      []model.RawBlock{{ID: "minecraft:stairs", X: 0, Y: 0, Z: 0}},
			styles:      stairsStyle,
			wantBlock:   emptyMasks,
			wantMicro:   emptyMasks,
			wantComplex: []model.FaceMask{faceMask(true, true, true, true, true, true)},
		},
		{
			name: "complex block surrounded",
			setupOcc: func() *stateful.OccupancyIndex {
				occ := stateful.NewOccupancyIndex()
				occ.FillRegion(-4, -4, -4, 12, 12, 12, true)
				return occ
			},
			blocks:      []model.RawBlock{{ID: "minecraft:stairs", X: 0, Y: 0, Z: 0}},
			styles:      stairsStyle,
			wantBlock:   emptyMasks,
			wantMicro:   emptyMasks,
			wantComplex: []model.FaceMask{faceMask(false, false, false, false, false, false)},
		},
		{
			name:     "micro faces",
			setupOcc: twoAdjacentFullCuboidsOcc,
			microRegions: []model.Cuboid{
				{ID: "minecraft:slab", X: 0, Y: 0, Z: 0, Width: 1, Height: 1, Depth: 1},
				{ID: "minecraft:slab", X: 1, Y: 0, Z: 0, Width: 1, Height: 1, Depth: 1},
			},
			styles:      styleIndex(),
			wantBlock:   emptyMasks,
			wantMicro:   []model.FaceMask{faceMask(false, true, true, false, true, false), faceMask(false, true, true, false, false, false)},
			wantComplex: emptyMasks,
		},
		{
			name:        "complex block without style",
			blocks:      []model.RawBlock{{ID: "minecraft:unknown_block", X: 0, Y: 0, Z: 0}},
			styles:      styleIndex(),
			wantBlock:   emptyMasks,
			wantMicro:   emptyMasks,
			wantComplex: []model.FaceMask{faceMask(false, false, false, false, false, false)},
		},
		{
			name: "complex block with full block style",
			blocks: []model.RawBlock{
				{ID: "minecraft:stone", X: 0, Y: 0, Z: 0},
			},
			styles: styleIndex(struct {
				id        string
				prop      string
				alignment model.GridAlignment
			}{"minecraft:stone", "", model.GridFullBlock}),
			wantBlock:   emptyMasks,
			wantMicro:   emptyMasks,
			wantComplex: []model.FaceMask{faceMask(false, false, false, false, false, false)},
		},
		{
			name: "transparent neighbor does not hide face",
			setupOcc: func() *stateful.OccupancyIndex {
				occ := stateful.NewOccupancyIndex()
				occ.FillRegion(0, 0, 0, 4, 4, 4, true)
				occ.FillRegion(4, 0, 0, 4, 4, 4, false)
				return occ
			},
			blockRegions: []model.Cuboid{
				{ID: "minecraft:stone", X: 0, Y: 0, Z: 0, Width: 1, Height: 1, Depth: 1},
			},
			styles:      styleIndex(),
			wantBlock:   []model.FaceMask{faceMask(true, true, true, true, true, true)},
			wantMicro:   emptyMasks,
			wantComplex: emptyMasks,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			occ := stateful.NewOccupancyIndex()
			if tt.setupOcc != nil {
				occ = tt.setupOcc()
			}

			vis := NewFaceCuller(&mockMergerPropsKeyBuilder{}, NewCuboidHelper()).
				Run(occ, tt.blockRegions, tt.microRegions, tt.blocks, tt.styles)

			require.Equal(t, tt.wantBlock, vis.BlockFaces)
			require.Equal(t, tt.wantMicro, vis.MicroFaces)
			require.Equal(t, tt.wantComplex, vis.ComplexFaces)
		})
	}
}
