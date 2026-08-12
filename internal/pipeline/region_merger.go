package pipeline

import (
	"sort"

	"mc2lua/internal/model"
	"mc2lua/internal/stateful"
)

type mergerCuboidHelper interface {
	Center(x, size int) float64
}

type RegionMerger struct {
	cuboidHelper mergerCuboidHelper
}

func NewRegionMerger(ch mergerCuboidHelper) *RegionMerger {
	return &RegionMerger{cuboidHelper: ch}
}

func (svc *RegionMerger) Run(grid *stateful.VoxelIndex) []model.Cuboid {
	var regions []model.Cuboid
	for _, comp := range svc.connectedComponents(grid) {
		regions = append(regions, svc.processComponent(grid, comp)...)
	}
	return regions
}

func (svc *RegionMerger) connectedComponents(grid *stateful.VoxelIndex) [][]*model.MergedBlock {
	visited := make(map[*model.MergedBlock]bool)
	var components [][]*model.MergedBlock

	for _, seed := range grid.Blocks() {
		if visited[seed] {
			continue
		}
		comp := svc.bfsComponent(grid, seed, visited)
		components = append(components, comp)
	}

	return components
}

func (svc *RegionMerger) bfsComponent(grid *stateful.VoxelIndex, seed *model.MergedBlock, visited map[*model.MergedBlock]bool) []*model.MergedBlock {
	var neighborDirs = [][3]int{
		{-1, 0, 0}, {1, 0, 0},
		{0, -1, 0}, {0, 1, 0},
		{0, 0, -1}, {0, 0, 1},
	}

	var comp []*model.MergedBlock
	queue := []*model.MergedBlock{seed}
	visited[seed] = true

	for len(queue) > 0 {
		b := queue[0]
		queue = queue[1:]
		comp = append(comp, b)

		for _, d := range neighborDirs {
			nx, ny, nz := b.X+d[0], b.Y+d[1], b.Z+d[2]
			nb := grid.GetBlock(nx, ny, nz)
			if nb != nil && !visited[nb] && nb.ID == b.ID && nb.PropsKey == b.PropsKey {
				visited[nb] = true
				queue = append(queue, nb)
			}
		}
	}

	return comp
}

func (svc *RegionMerger) processComponent(grid *stateful.VoxelIndex, blocks []*model.MergedBlock) []model.Cuboid {
	compID := blocks[0].ID
	compProps := blocks[0].PropsKey

	byY := make(map[int][]*model.MergedBlock)
	for _, b := range blocks {
		byY[b.Y] = append(byY[b.Y], b)
	}

	var ys []int
	for y := range byY {
		ys = append(ys, y)
	}
	sort.Ints(ys)

	var regions []model.Cuboid
	for _, y := range ys {
		regions = append(regions, svc.processLayer(grid, byY[y], y, compID, compProps)...)
	}
	return regions
}

func (svc *RegionMerger) processLayer(grid *stateful.VoxelIndex, blocks []*model.MergedBlock, y int, id, props string) []model.Cuboid {
	var unmerged []*model.MergedBlock
	for _, b := range blocks {
		if !b.Merged {
			unmerged = append(unmerged, b)
		}
	}
	if len(unmerged) == 0 {
		return nil
	}

	rects := svc.decomposeLayer(unmerged)

	var regions []model.Cuboid
	for _, rect := range rects {
		regions = append(regions, svc.expandRegion(grid, rect, y, id, props))
	}
	return regions
}

func (svc *RegionMerger) buildBoolGrid(blocks []*model.MergedBlock) (grid [][]bool, xMin, zMin int) {
	xMin, xMax := blocks[0].X, blocks[0].X
	zMin, zMax := blocks[0].Z, blocks[0].Z
	for _, b := range blocks {
		if b.X < xMin {
			xMin = b.X
		}
		if b.X > xMax {
			xMax = b.X
		}
		if b.Z < zMin {
			zMin = b.Z
		}
		if b.Z > zMax {
			zMax = b.Z
		}
	}

	rows := xMax - xMin + 1
	cols := zMax - zMin + 1

	grid = make([][]bool, rows)
	for i := range grid {
		grid[i] = make([]bool, cols)
	}
	for _, b := range blocks {
		grid[b.X-xMin][b.Z-zMin] = true
	}

	return grid, xMin, zMin
}

func (svc *RegionMerger) decomposeLayer(blocks []*model.MergedBlock) []model.Rect2D {
	if len(blocks) == 0 {
		return nil
	}

	grid, xMin, zMin := svc.buildBoolGrid(blocks)

	var rects []model.Rect2D
	for {
		rectRow, rectCol, rectRows, rectCols := svc.findLargestRect(grid)
		if rectRows == 0 || rectCols == 0 {
			break
		}
		rects = append(rects, model.Rect2D{
			X: xMin + rectRow, Z: zMin + rectCol,
			Width: rectRows, Depth: rectCols,
		})
		for x := rectRow; x < rectRow+rectRows; x++ {
			for z := rectCol; z < rectCol+rectCols; z++ {
				grid[x][z] = false
			}
		}
	}

	return rects
}

func (svc *RegionMerger) findLargestRect(grid [][]bool) (row, col, rows, cols int) {
	totalRows := len(grid)
	if totalRows == 0 {
		return 0, 0, 0, 0
	}
	totalCols := len(grid[0])
	if totalCols == 0 {
		return 0, 0, 0, 0
	}

	heights := make([]int, totalCols)
	bestArea := 0

	for r := 0; r < totalRows; r++ {
		svc.updateHistogram(heights, grid[r])
		area, rowStart, colStart, nCols, nRows := svc.maxRectInHistogram(heights, r)
		if area > bestArea {
			bestArea = area
			row, col, rows, cols = rowStart, colStart, nRows, nCols
		}
	}

	return row, col, rows, cols
}

func (svc *RegionMerger) updateHistogram(heights []int, row []bool) {
	for c := range heights {
		if row[c] {
			heights[c]++
		} else {
			heights[c] = 0
		}
	}
}

func (svc *RegionMerger) maxRectInHistogram(heights []int, row int) (area, rowStart, colStart, cols, rows int) {
	type bar struct {
		height, idx int
	}
	var stack []bar
	totalCols := len(heights)

	for c := 0; c <= totalCols; c++ {
		ch := 0
		if c < totalCols {
			ch = heights[c]
		}

		start := c
		for len(stack) > 0 && stack[len(stack)-1].height > ch {
			b := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			candidate := b.height * (c - b.idx)
			if candidate > area {
				area = candidate
				rowStart = row - b.height + 1
				colStart = b.idx
				rows = b.height
				cols = c - b.idx
			}
			start = b.idx
		}
		stack = append(stack, bar{ch, start})
	}

	return
}

func (svc *RegionMerger) expandRegion(grid *stateful.VoxelIndex, rect model.Rect2D, y int, id, props string) model.Cuboid {
	yMax := svc.expandUp(grid, y, rect, id, props)
	yMin := svc.expandDown(grid, y, rect, id, props)
	height := yMax - yMin + 1

	cx := svc.cuboidHelper.Center(rect.X, rect.Width)
	cz := svc.cuboidHelper.Center(rect.Z, rect.Depth)
	cy := svc.cuboidHelper.Center(yMin, height)

	svc.markMerged(grid, yMin, yMax, rect)

	return model.Cuboid{
		ID: id, PropsKey: props,
		X: cx, Y: cy, Z: cz,
		Width: rect.Width, Depth: rect.Depth, Height: height,
	}
}

func (svc *RegionMerger) expandUp(grid *stateful.VoxelIndex, yStart int, rect model.Rect2D, id, props string) int {
	yMax := yStart
	for {
		nextY := yMax + 1
		if !svc.canMergeColumn(grid, nextY, rect, id, props) {
			break
		}
		yMax++
	}
	return yMax
}

func (svc *RegionMerger) expandDown(grid *stateful.VoxelIndex, yStart int, rect model.Rect2D, id, props string) int {
	yMin := yStart
	for {
		nextY := yMin - 1
		if !svc.canMergeColumn(grid, nextY, rect, id, props) {
			break
		}
		yMin--
	}
	return yMin
}

func (svc *RegionMerger) canMergeColumn(grid *stateful.VoxelIndex, y int, rect model.Rect2D, id, props string) bool {
	for x := rect.X; x < rect.X+rect.Width; x++ {
		for z := rect.Z; z < rect.Z+rect.Depth; z++ {
			if b := grid.GetBlock(x, y, z); b == nil || b.Merged || b.ID != id || b.PropsKey != props {
				return false
			}
		}
	}
	return true
}

func (svc *RegionMerger) markMerged(grid *stateful.VoxelIndex, yMin, yMax int, rect model.Rect2D) {
	for yb := yMin; yb <= yMax; yb++ {
		for x := rect.X; x < rect.X+rect.Width; x++ {
			for z := rect.Z; z < rect.Z+rect.Depth; z++ {
				if blk := grid.GetBlock(x, yb, z); blk != nil {
					blk.Merged = true
				}
			}
		}
	}
}
