package pipeline

import (
	"testing"

	"mc2lua/internal/index"
	"mc2lua/internal/model"

	"github.com/stretchr/testify/require"
)

type mockMergerPropsKeyBuilder struct {
	runFn func(props map[string]string) string
}

func (m *mockMergerPropsKeyBuilder) Run(props map[string]string) string {
	if m.runFn != nil {
		return m.runFn(props)
	}
	return ""
}

func styleIndex(entries ...struct {
	id        string
	prop      string
	alignment model.GridAlignment
}) index.StyleIndex {
	idx := index.NewStyleIndex()
	for _, e := range entries {
		idx.Add(e.id, e.prop, model.StyledBlock{
			ID:            e.id,
			PropsKey:      e.prop,
			GridAlignment: e.alignment,
		})
	}
	return *idx
}

func fullBlockAt(id, props string, x, y, z int) model.RawBlock {
	return model.RawBlock{ID: id, Props: map[string]string{}, X: x, Y: y, Z: z}
}

func blockAt(id string, props map[string]string, x, y, z int) model.RawBlock {
	return model.RawBlock{ID: id, Props: props, X: x, Y: y, Z: z}
}

func cuboidVol(c model.Cuboid) int {
	return c.Width * c.Depth * c.Height
}

func totalVolume(cuboids []model.Cuboid) int {
	v := 0
	for _, c := range cuboids {
		v += cuboidVol(c)
	}
	return v
}

func collectVoxels(cuboids []model.Cuboid) map[[3]int]int {
	voxels := make(map[[3]int]int)
	for _, c := range cuboids {
		xStart := int(c.X - float64(c.Width-1)/2.0)
		yStart := int(c.Y - float64(c.Height-1)/2.0)
		zStart := int(c.Z - float64(c.Depth-1)/2.0)
		for x := 0; x < c.Width; x++ {
			for y := 0; y < c.Height; y++ {
				for z := 0; z < c.Depth; z++ {
					key := [3]int{xStart + x, yStart + y, zStart + z}
					voxels[key]++
				}
			}
		}
	}
	return voxels
}

func countOverlaps(voxels map[[3]int]int) int {
	n := 0
	for _, count := range voxels {
		if count > 1 {
			n++
		}
	}
	return n
}

func TestRegionMerger_New(t *testing.T) {
	t.Parallel()

	svc := NewRegionMerger()
	require.NotNil(t, svc)
}

func TestRegionMerger_Run(t *testing.T) {
	t.Parallel()

	defaultPKB := &mockMergerPropsKeyBuilder{
		runFn: func(props map[string]string) string {
			return ""
		},
	}

	tests := []struct {
		name       string
		blocks     []model.RawBlock
		styles     index.StyleIndex
		pkb        *mockMergerPropsKeyBuilder
		wantCount  int
		wantVolume int
		wantNoOver bool
		wantCheck  func(t *testing.T, cuboids []model.Cuboid)
	}{
		{
			name:       "empty blocks",
			blocks:     nil,
			styles:     styleIndex(),
			wantCount:  0,
			wantVolume: 0,
		},
		{
			name:       "no matching styles",
			blocks:     []model.RawBlock{fullBlockAt("minecraft:stone", "", 0, 0, 0)},
			styles:     styleIndex(),
			wantCount:  0,
			wantVolume: 0,
		},
		{
			name:   "non-full blocks filtered out",
			blocks: []model.RawBlock{fullBlockAt("minecraft:stone", "", 0, 0, 0)},
			styles: styleIndex(struct {
				id   string
				prop string
				alignment model.GridAlignment
			}{"minecraft:stone", "", model.GridNotAligned}),
			wantCount:  0,
			wantVolume: 0,
		},
		{
			name:   "single block",
			blocks: []model.RawBlock{fullBlockAt("minecraft:stone", "", 0, 0, 0)},
			styles: styleIndex(struct {
				id   string
				prop string
				alignment model.GridAlignment
			}{"minecraft:stone", "", model.GridFullBlock}),
			wantCount:  1,
			wantVolume: 1,
			wantCheck: func(t *testing.T, cuboids []model.Cuboid) {
				c := cuboids[0]
				require.Equal(t, "minecraft:stone", c.ID)
				require.Equal(t, "", c.PropsKey)
				require.Equal(t, 0.0, c.X)
				require.Equal(t, 0.0, c.Y)
				require.Equal(t, 0.0, c.Z)
				require.Equal(t, 1, c.Width)
				require.Equal(t, 1, c.Depth)
				require.Equal(t, 1, c.Height)
			},
		},
		{
			name: "single row along X",
			blocks: []model.RawBlock{
				fullBlockAt("minecraft:stone", "", 0, 0, 0),
				fullBlockAt("minecraft:stone", "", 1, 0, 0),
				fullBlockAt("minecraft:stone", "", 2, 0, 0),
			},
			styles: styleIndex(struct {
				id   string
				prop string
				alignment model.GridAlignment
			}{"minecraft:stone", "", model.GridFullBlock}),
			wantCount:  1,
			wantVolume: 3,
			wantCheck: func(t *testing.T, cuboids []model.Cuboid) {
				c := cuboids[0]
				require.Equal(t, 3, c.Width)
				require.Equal(t, 1, c.Depth)
				require.Equal(t, 1, c.Height)
				require.Equal(t, 1.0, c.X)
				require.Equal(t, 0.0, c.Z)
			},
		},
		{
			name: "single column along Z",
			blocks: []model.RawBlock{
				fullBlockAt("minecraft:stone", "", 0, 0, 0),
				fullBlockAt("minecraft:stone", "", 0, 0, 1),
				fullBlockAt("minecraft:stone", "", 0, 0, 2),
			},
			styles: styleIndex(struct {
				id   string
				prop string
				alignment model.GridAlignment
			}{"minecraft:stone", "", model.GridFullBlock}),
			wantCount:  1,
			wantVolume: 3,
			wantCheck: func(t *testing.T, cuboids []model.Cuboid) {
				c := cuboids[0]
				require.Equal(t, 1, c.Width)
				require.Equal(t, 3, c.Depth)
				require.Equal(t, 1, c.Height)
				require.Equal(t, 1.0, c.Z)
			},
		},
		{
			name: "vertical stack merges into one cuboid",
			blocks: []model.RawBlock{
				fullBlockAt("minecraft:stone", "", 0, 0, 0),
				fullBlockAt("minecraft:stone", "", 0, 1, 0),
				fullBlockAt("minecraft:stone", "", 0, 2, 0),
			},
			styles: styleIndex(struct {
				id   string
				prop string
				alignment model.GridAlignment
			}{"minecraft:stone", "", model.GridFullBlock}),
			wantCount:  1,
			wantVolume: 3,
			wantCheck: func(t *testing.T, cuboids []model.Cuboid) {
				c := cuboids[0]
				require.Equal(t, 1, c.Width)
				require.Equal(t, 1, c.Depth)
				require.Equal(t, 3, c.Height)
				require.Equal(t, 1.0, c.Y)
			},
		},
		{
			name: "2x2 layer on same Y",
			blocks: []model.RawBlock{
				fullBlockAt("minecraft:stone", "", 0, 0, 0),
				fullBlockAt("minecraft:stone", "", 1, 0, 0),
				fullBlockAt("minecraft:stone", "", 0, 0, 1),
				fullBlockAt("minecraft:stone", "", 1, 0, 1),
			},
			styles: styleIndex(struct {
				id   string
				prop string
				alignment model.GridAlignment
			}{"minecraft:stone", "", model.GridFullBlock}),
			wantCount:  1,
			wantVolume: 4,
			wantCheck: func(t *testing.T, cuboids []model.Cuboid) {
				c := cuboids[0]
				require.Equal(t, 2, c.Width)
				require.Equal(t, 2, c.Depth)
				require.Equal(t, 1, c.Height)
			},
		},
		{
			name: "3x3x3 cube",
			blocks: func() []model.RawBlock {
				var bb []model.RawBlock
				for x := 0; x < 3; x++ {
					for y := 0; y < 3; y++ {
						for z := 0; z < 3; z++ {
							bb = append(bb, fullBlockAt("minecraft:stone", "", x, y, z))
						}
					}
				}
				return bb
			}(),
			styles: styleIndex(struct {
				id   string
				prop string
				alignment model.GridAlignment
			}{"minecraft:stone", "", model.GridFullBlock}),
			wantCount:  1,
			wantVolume: 27,
			wantCheck: func(t *testing.T, cuboids []model.Cuboid) {
				c := cuboids[0]
				require.Equal(t, 3, c.Width)
				require.Equal(t, 3, c.Depth)
				require.Equal(t, 3, c.Height)
			},
		},
		{
			name: "two disconnected components on same layer",
			blocks: []model.RawBlock{
				fullBlockAt("minecraft:stone", "", 0, 0, 0),
				fullBlockAt("minecraft:stone", "", 0, 0, 1),
				fullBlockAt("minecraft:stone", "", 10, 0, 10),
				fullBlockAt("minecraft:stone", "", 10, 0, 11),
			},
			styles: styleIndex(struct {
				id   string
				prop string
				alignment model.GridAlignment
			}{"minecraft:stone", "", model.GridFullBlock}),
			wantCount:  2,
			wantVolume: 4,
			wantNoOver: true,
		},
		{
			name: "different IDs produce separate cuboids",
			blocks: []model.RawBlock{
				fullBlockAt("minecraft:stone", "", 0, 0, 0),
				fullBlockAt("minecraft:dirt", "", 0, 0, 1),
			},
			styles: styleIndex(
				struct {
					id   string
					prop string
					alignment model.GridAlignment
				}{"minecraft:stone", "", model.GridFullBlock},
				struct {
					id   string
					prop string
					alignment model.GridAlignment
				}{"minecraft:dirt", "", model.GridFullBlock},
			),
			wantCount:  2,
			wantVolume: 2,
		},
		{
			name: "different props produce separate cuboids",
			blocks: []model.RawBlock{
				blockAt("minecraft:stone", map[string]string{"axis": "x"}, 0, 0, 0),
				blockAt("minecraft:stone", map[string]string{"axis": "y"}, 0, 0, 1),
			},
			styles: styleIndex(
				struct {
					id   string
					prop string
					alignment model.GridAlignment
				}{"minecraft:stone", "axis=x", model.GridFullBlock},
				struct {
					id   string
					prop string
					alignment model.GridAlignment
				}{"minecraft:stone", "axis=y", model.GridFullBlock},
			),
			pkb: &mockMergerPropsKeyBuilder{
				runFn: func(props map[string]string) string {
					if v, ok := props["axis"]; ok {
						return "axis=" + v
					}
					return ""
				},
			},
			wantCount:  2,
			wantVolume: 2,
			wantCheck: func(t *testing.T, cuboids []model.Cuboid) {
				propsSet := make(map[string]int)
				for _, c := range cuboids {
					propsSet[c.PropsKey]++
				}
				require.Equal(t, 1, propsSet["axis=x"])
				require.Equal(t, 1, propsSet["axis=y"])
			},
		},
		{
			name: "L-shape on same layer produces two rects",
			blocks: []model.RawBlock{
				fullBlockAt("minecraft:stone", "", 0, 0, 0),
				fullBlockAt("minecraft:stone", "", 1, 0, 0),
				fullBlockAt("minecraft:stone", "", 2, 0, 0),
				fullBlockAt("minecraft:stone", "", 2, 0, 1),
				fullBlockAt("minecraft:stone", "", 2, 0, 2),
			},
			styles: styleIndex(struct {
				id   string
				prop string
				alignment model.GridAlignment
			}{"minecraft:stone", "", model.GridFullBlock}),
			wantCount:  2,
			wantVolume: 5,
			wantNoOver: true,
		},
		{
			name: "vertical gap produces two cuboids",
			blocks: []model.RawBlock{
				fullBlockAt("minecraft:stone", "", 0, 0, 0),
				fullBlockAt("minecraft:stone", "", 0, 1, 0),
				fullBlockAt("minecraft:stone", "", 0, 3, 0),
				fullBlockAt("minecraft:stone", "", 0, 4, 0),
			},
			styles: styleIndex(struct {
				id   string
				prop string
				alignment model.GridAlignment
			}{"minecraft:stone", "", model.GridFullBlock}),
			wantCount:  2,
			wantVolume: 4,
			wantNoOver: true,
		},
		{
			name: "irregular shape - no overlap between cuboids",
			blocks: []model.RawBlock{
				fullBlockAt("minecraft:stone", "", 0, 0, 0),
				fullBlockAt("minecraft:stone", "", 1, 0, 0),
				fullBlockAt("minecraft:stone", "", 0, 0, 1),
				fullBlockAt("minecraft:stone", "", 0, 1, 0),
				fullBlockAt("minecraft:stone", "", 0, 1, 1),
				fullBlockAt("minecraft:stone", "", 2, 0, 0),
				fullBlockAt("minecraft:stone", "", 2, 0, 1),
				fullBlockAt("minecraft:stone", "", 2, 0, 2),
			},
			styles: styleIndex(struct {
				id   string
				prop string
				alignment model.GridAlignment
			}{"minecraft:stone", "", model.GridFullBlock}),
			wantVolume: 8,
			wantNoOver: true,
		},
		{
			name: "overlap regression - upper partial layer",
			blocks: []model.RawBlock{
				fullBlockAt("minecraft:stone", "", 0, 0, 0),
				fullBlockAt("minecraft:stone", "", 1, 0, 0),
				fullBlockAt("minecraft:stone", "", 0, 1, 0),
				fullBlockAt("minecraft:stone", "", 1, 1, 0),
				fullBlockAt("minecraft:stone", "", 2, 1, 0),
			},
			styles: styleIndex(struct {
				id   string
				prop string
				alignment model.GridAlignment
			}{"minecraft:stone", "", model.GridFullBlock}),
			wantVolume: 5,
			wantNoOver: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			pkb := tt.pkb
			if pkb == nil {
				pkb = defaultPKB
			}

			indexer := NewBlockVoxelIndexer(pkb)
			grid := indexer.Run(tt.blocks, tt.styles)
			svc := NewRegionMerger()
			cuboids := svc.Run(grid)

			if tt.wantCount > 0 {
				require.Len(t, cuboids, tt.wantCount)
			}

			gotVol := totalVolume(cuboids)
			require.Equal(t, tt.wantVolume, gotVol, "total volume mismatch")

			if tt.wantNoOver {
				voxels := collectVoxels(cuboids)
				overlaps := countOverlaps(voxels)
				require.Zero(t, overlaps, "cuboids must not overlap")
			}

			if tt.wantCheck != nil {
				tt.wantCheck(t, cuboids)
			}
		})
	}
}

func TestRegionMerger_findLargestRect(t *testing.T) {
	t.Parallel()

	svc := NewRegionMerger()

	tests := []struct {
		name               string
		grid               [][]bool
		wantRow, wantCol   int
		wantRows, wantCols int
	}{
		{
			name:    "nil grid",
			grid:    nil,
			wantRow: 0, wantCol: 0,
			wantRows: 0, wantCols: 0,
		},
		{
			name:    "empty grid",
			grid:    [][]bool{},
			wantRow: 0, wantCol: 0,
			wantRows: 0, wantCols: 0,
		},
		{
			name:    "empty row",
			grid:    [][]bool{{}},
			wantRow: 0, wantCol: 0,
			wantRows: 0, wantCols: 0,
		},
		{
			name:    "single cell",
			grid:    [][]bool{{true}},
			wantRow: 0, wantCol: 0,
			wantRows: 1, wantCols: 1,
		},
		{
			name:    "empty grid",
			grid:    [][]bool{{false}},
			wantRow: 0, wantCol: 0,
			wantRows: 0, wantCols: 0,
		},
		{
			name: "2x2 all true",
			grid: [][]bool{
				{true, true},
				{true, true},
			},
			wantRow: 0, wantCol: 0,
			wantRows: 2, wantCols: 2,
		},
		{
			name: "single row",
			grid: [][]bool{
				{true, true, true},
			},
			wantRow: 0, wantCol: 0,
			wantRows: 1, wantCols: 3,
		},
		{
			name: "single column",
			grid: [][]bool{
				{true},
				{true},
				{true},
			},
			wantRow: 0, wantCol: 0,
			wantRows: 3, wantCols: 1,
		},
		{
			name: "L-shape prefers larger rectangle",
			grid: [][]bool{
				{true, false},
				{true, true},
			},
			wantRow: 0, wantCol: 0,
			wantRows: 2, wantCols: 1,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			row, col, rows, cols := svc.findLargestRect(tt.grid)
			require.Equal(t, tt.wantRow, row, "row")
			require.Equal(t, tt.wantCol, col, "col")
			require.Equal(t, tt.wantRows, rows, "rows")
			require.Equal(t, tt.wantCols, cols, "cols")
		})
	}
}

func TestRegionMerger_maxRectInHistogram(t *testing.T) {
	t.Parallel()

	svc := NewRegionMerger()

	tests := []struct {
		name                       string
		heights                    []int
		row                        int
		wantArea                   int
		wantRowStart, wantColStart int
		wantCols, wantRows         int
	}{
		{
			name:         "single bar height 1",
			heights:      []int{1},
			row:          0,
			wantArea:     1,
			wantRowStart: 0, wantColStart: 0,
			wantCols: 1, wantRows: 1,
		},
		{
			name:         "single bar height 3 at row 2",
			heights:      []int{3},
			row:          2,
			wantArea:     3,
			wantRowStart: 0, wantColStart: 0,
			wantCols: 1, wantRows: 3,
		},
		{
			name:         "two bars same height",
			heights:      []int{2, 2},
			row:          1,
			wantArea:     4,
			wantRowStart: 0, wantColStart: 0,
			wantCols: 2, wantRows: 2,
		},
		{
			name:         "descending heights picks tallest",
			heights:      []int{3, 2, 1},
			row:          2,
			wantArea:     4,
			wantRowStart: 1, wantColStart: 0,
			wantCols: 2, wantRows: 2,
		},
		{
			name:         "all zeros",
			heights:      []int{0, 0, 0},
			row:          0,
			wantArea:     0,
			wantRowStart: 0, wantColStart: 0,
			wantCols: 0, wantRows: 0,
		},
		{
			name:         "valley picks largest area",
			heights:      []int{2, 1, 2},
			row:          2,
			wantArea:     3,
			wantRowStart: 2, wantColStart: 0,
			wantCols: 3, wantRows: 1,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			area, rowStart, colStart, cols, rows := svc.maxRectInHistogram(tt.heights, tt.row)
			require.Equal(t, tt.wantArea, area, "area")
			require.Equal(t, tt.wantRowStart, rowStart, "rowStart")
			require.Equal(t, tt.wantColStart, colStart, "colStart")
			require.Equal(t, tt.wantCols, cols, "cols")
			require.Equal(t, tt.wantRows, rows, "rows")
		})
	}
}

func TestRegionMerger_decomposeLayer(t *testing.T) {
	t.Parallel()

	svc := NewRegionMerger()

	tests := []struct {
		name   string
		blocks []*model.MergedBlock
		want   []model.Rect2D
	}{
		{
			name:   "empty",
			blocks: nil,
			want:   nil,
		},
		{
			name: "single block",
			blocks: []*model.MergedBlock{
				{X: 5, Z: 10},
			},
			want: []model.Rect2D{
				{X: 5, Z: 10, Width: 1, Depth: 1},
			},
		},
		{
			name: "two disconnected blocks",
			blocks: []*model.MergedBlock{
				{X: 0, Z: 0},
				{X: 5, Z: 5},
			},
			want: []model.Rect2D{
				{X: 0, Z: 0, Width: 1, Depth: 1},
				{X: 5, Z: 5, Width: 1, Depth: 1},
			},
		},
		{
			name: "row of three blocks decomposes to one rect",
			blocks: []*model.MergedBlock{
				{X: 0, Z: 0},
				{X: 1, Z: 0},
				{X: 2, Z: 0},
			},
			want: []model.Rect2D{
				{X: 0, Z: 0, Width: 3, Depth: 1},
			},
		},
		{
			name: "2x2 square decomposes to one rect",
			blocks: []*model.MergedBlock{
				{X: 0, Z: 0}, {X: 1, Z: 0},
				{X: 0, Z: 1}, {X: 1, Z: 1},
			},
			want: []model.Rect2D{
				{X: 0, Z: 0, Width: 2, Depth: 2},
			},
		},
		{
			name: "L-shape decomposes to two rects",
			blocks: []*model.MergedBlock{
				{X: 0, Z: 0},
				{X: 0, Z: 1},
				{X: 0, Z: 2},
				{X: 1, Z: 2},
				{X: 2, Z: 2},
			},
			want: []model.Rect2D{
				{X: 0, Z: 0, Width: 1, Depth: 3},
				{X: 1, Z: 2, Width: 2, Depth: 1},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := svc.decomposeLayer(tt.blocks)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestRegionMerger_expandRegion(t *testing.T) {
	t.Parallel()

	svc := NewRegionMerger()

	tests := []struct {
		name       string
		setup      func() (*index.VoxelIndex, model.Rect2D, int, string, string)
		want       model.Cuboid
		wantMerged []string
	}{
		{
			name: "single block at origin expands nowhere",
			setup: func() (*index.VoxelIndex, model.Rect2D, int, string, string) {
				grid := index.NewVoxelIndex()
				grid.AddBlock(&model.MergedBlock{ID: "stone", X: 0, Y: 5, Z: 0})
				rect := model.Rect2D{X: 0, Z: 0, Width: 1, Depth: 1}
				return grid, rect, 5, "stone", ""
			},
			want: model.Cuboid{
				ID: "stone", X: 0, Y: 5, Z: 0,
				Width: 1, Depth: 1, Height: 1,
			},
			wantMerged: []string{"0,5,0"},
		},
		{
			name: "expands up through contiguous blocks",
			setup: func() (*index.VoxelIndex, model.Rect2D, int, string, string) {
				grid := index.NewVoxelIndex()
				grid.AddBlock(&model.MergedBlock{ID: "stone", X: 0, Y: 5, Z: 0})
				grid.AddBlock(&model.MergedBlock{ID: "stone", X: 0, Y: 6, Z: 0})
				grid.AddBlock(&model.MergedBlock{ID: "stone", X: 0, Y: 7, Z: 0})
				rect := model.Rect2D{X: 0, Z: 0, Width: 1, Depth: 1}
				return grid, rect, 5, "stone", ""
			},
			want: model.Cuboid{
				ID: "stone", X: 0, Y: 6, Z: 0,
				Width: 1, Depth: 1, Height: 3,
			},
			wantMerged: []string{"0,5,0", "0,6,0", "0,7,0"},
		},
		{
			name: "does not expand past gap",
			setup: func() (*index.VoxelIndex, model.Rect2D, int, string, string) {
				grid := index.NewVoxelIndex()
				grid.AddBlock(&model.MergedBlock{ID: "stone", X: 0, Y: 5, Z: 0})
				grid.AddBlock(&model.MergedBlock{ID: "stone", X: 0, Y: 7, Z: 0})
				rect := model.Rect2D{X: 0, Z: 0, Width: 1, Depth: 1}
				return grid, rect, 5, "stone", ""
			},
			want: model.Cuboid{
				ID: "stone", X: 0, Y: 5, Z: 0,
				Width: 1, Depth: 1, Height: 1,
			},
			wantMerged: []string{"0,5,0"},
		},
		{
			name: "expands down through contiguous blocks",
			setup: func() (*index.VoxelIndex, model.Rect2D, int, string, string) {
				grid := index.NewVoxelIndex()
				grid.AddBlock(&model.MergedBlock{ID: "stone", X: 0, Y: 5, Z: 0})
				grid.AddBlock(&model.MergedBlock{ID: "stone", X: 0, Y: 4, Z: 0})
				grid.AddBlock(&model.MergedBlock{ID: "stone", X: 0, Y: 3, Z: 0})
				rect := model.Rect2D{X: 0, Z: 0, Width: 1, Depth: 1}
				return grid, rect, 5, "stone", ""
			},
			want: model.Cuboid{
				ID: "stone", X: 0, Y: 4, Z: 0,
				Width: 1, Depth: 1, Height: 3,
			},
			wantMerged: []string{"0,3,0", "0,4,0", "0,5,0"},
		},
		{
			name: "stops at already merged blocks",
			setup: func() (*index.VoxelIndex, model.Rect2D, int, string, string) {
				grid := index.NewVoxelIndex()
				grid.AddBlock(&model.MergedBlock{ID: "stone", X: 0, Y: 5, Z: 0})
				merged := &model.MergedBlock{ID: "stone", X: 0, Y: 6, Z: 0, Merged: true}
				grid.AddBlock(merged)
				rect := model.Rect2D{X: 0, Z: 0, Width: 1, Depth: 1}
				return grid, rect, 5, "stone", ""
			},
			want: model.Cuboid{
				ID: "stone", X: 0, Y: 5, Z: 0,
				Width: 1, Depth: 1, Height: 1,
			},
			wantMerged: []string{"0,5,0"},
		},
		{
			name: "expands in 2x2 column",
			setup: func() (*index.VoxelIndex, model.Rect2D, int, string, string) {
				grid := index.NewVoxelIndex()
				for x := 0; x < 2; x++ {
					for z := 0; z < 2; z++ {
						for y := 5; y <= 7; y++ {
							grid.AddBlock(&model.MergedBlock{
								ID: "stone", X: x, Y: y, Z: z,
							})
						}
					}
				}
				rect := model.Rect2D{X: 0, Z: 0, Width: 2, Depth: 2}
				return grid, rect, 5, "stone", ""
			},
			want: model.Cuboid{
				ID: "stone", X: 0.5, Y: 6, Z: 0.5,
				Width: 2, Depth: 2, Height: 3,
			},
			wantMerged: []string{
				"0,5,0", "1,5,0", "0,5,1", "1,5,1",
				"0,6,0", "1,6,0", "0,6,1", "1,6,1",
				"0,7,0", "1,7,0", "0,7,1", "1,7,1",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			grid, rect, y, id, props := tt.setup()
			got := svc.expandRegion(grid, rect, y, id, props)
			require.Equal(t, tt.want, got)

			mergedSet := make(map[string]bool)
			for _, k := range tt.wantMerged {
				mergedSet[k] = true
			}
			for _, b := range grid.Blocks() {
				key := itoa(b.X) + "," + itoa(b.Y) + "," + itoa(b.Z)
				if mergedSet[key] {
					require.True(t, b.Merged, "expected merged: %s", key)
				}
			}
		})
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	s := ""
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	if neg {
		s = "-" + s
	}
	return s
}

func TestRegionMerger_connectedComponents(t *testing.T) {
	t.Parallel()

	svc := NewRegionMerger()

	tests := []struct {
		name  string
		setup func() *index.VoxelIndex
		want  int
	}{
		{
			name: "empty grid",
			setup: func() *index.VoxelIndex {
				return index.NewVoxelIndex()
			},
			want: 0,
		},
		{
			name: "single block",
			setup: func() *index.VoxelIndex {
				grid := index.NewVoxelIndex()
				grid.AddBlock(&model.MergedBlock{ID: "stone", X: 0, Y: 0, Z: 0})
				return grid
			},
			want: 1,
		},
		{
			name: "two disconnected blocks",
			setup: func() *index.VoxelIndex {
				grid := index.NewVoxelIndex()
				grid.AddBlock(&model.MergedBlock{ID: "stone", X: 0, Y: 0, Z: 0})
				grid.AddBlock(&model.MergedBlock{ID: "stone", X: 10, Y: 0, Z: 10})
				return grid
			},
			want: 2,
		},
		{
			name: "two adjacent blocks are one component",
			setup: func() *index.VoxelIndex {
				grid := index.NewVoxelIndex()
				grid.AddBlock(&model.MergedBlock{ID: "stone", X: 0, Y: 0, Z: 0})
				grid.AddBlock(&model.MergedBlock{ID: "stone", X: 1, Y: 0, Z: 0})
				return grid
			},
			want: 1,
		},
		{
			name: "different IDs are separate components",
			setup: func() *index.VoxelIndex {
				grid := index.NewVoxelIndex()
				grid.AddBlock(&model.MergedBlock{ID: "stone", X: 0, Y: 0, Z: 0})
				grid.AddBlock(&model.MergedBlock{ID: "dirt", X: 1, Y: 0, Z: 0})
				return grid
			},
			want: 2,
		},
		{
			name: "different props are separate components",
			setup: func() *index.VoxelIndex {
				grid := index.NewVoxelIndex()
				grid.AddBlock(&model.MergedBlock{ID: "stone", PropsKey: "axis=x", X: 0, Y: 0, Z: 0})
				grid.AddBlock(&model.MergedBlock{ID: "stone", PropsKey: "axis=y", X: 1, Y: 0, Z: 0})
				return grid
			},
			want: 2,
		},
		{
			name: "vertical neighbors are one component",
			setup: func() *index.VoxelIndex {
				grid := index.NewVoxelIndex()
				grid.AddBlock(&model.MergedBlock{ID: "stone", X: 0, Y: 0, Z: 0})
				grid.AddBlock(&model.MergedBlock{ID: "stone", X: 0, Y: 1, Z: 0})
				grid.AddBlock(&model.MergedBlock{ID: "stone", X: 0, Y: 2, Z: 0})
				return grid
			},
			want: 1,
		},
		{
			name: "diagonal is not adjacent",
			setup: func() *index.VoxelIndex {
				grid := index.NewVoxelIndex()
				grid.AddBlock(&model.MergedBlock{ID: "stone", X: 0, Y: 0, Z: 0})
				grid.AddBlock(&model.MergedBlock{ID: "stone", X: 1, Y: 0, Z: 1})
				return grid
			},
			want: 2,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			grid := tt.setup()
			comps := svc.connectedComponents(grid)
			require.Equal(t, tt.want, len(comps))

			totalBlocks := 0
			for _, comp := range comps {
				totalBlocks += len(comp)
			}
			require.Equal(t, len(grid.Blocks()), totalBlocks)
		})
	}
}


