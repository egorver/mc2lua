package pipeline

import (
	"mc2lua/internal/model"
	"mc2lua/internal/stateful"
)

type occupancyPropsKeyBuilder interface {
	Run(props map[string]string) string
}

type OccupancyIndexer struct {
	propsKeyBuilder occupancyPropsKeyBuilder
}

func NewOccupancyIndexer(pkb occupancyPropsKeyBuilder) *OccupancyIndexer {
	return &OccupancyIndexer{propsKeyBuilder: pkb}
}

func (svc *OccupancyIndexer) Run(blocks []model.RawBlock, blockIdx, microIdx *stateful.VoxelIndex, styles stateful.StyleIndex) *stateful.OccupancyIndex {
	occ := stateful.NewOccupancyIndex()

	for _, b := range blockIdx.Blocks() {
		style, ok := styles.Get(b.ID, b.PropsKey)
		if !ok {
			continue
		}
		svc.fillCell(occ, b.X, b.Y, b.Z, style)
	}

	for _, b := range microIdx.Blocks() {
		style, ok := styles.Get(b.ID, b.PropsKey)
		if !ok {
			continue
		}
		occ.FillCell(b.X, b.Y, b.Z, !style.Transparent)
	}

	for _, b := range blocks {
		propsKey := svc.propsKeyBuilder.Run(b.Props)
		style, ok := styles.Get(b.ID, propsKey)
		if !ok || style.GridAlignment != model.GridNotAligned {
			continue
		}
		svc.fillComplexCell(occ, b.X, b.Y, b.Z, style)
	}

	return occ
}

func (svc *OccupancyIndexer) fillCell(occ *stateful.OccupancyIndex, x, y, z int, style model.StyledBlock) {
	occ.FillRegion(x*model.SubGridSize, y*model.SubGridSize, z*model.SubGridSize,
		model.SubGridSize, model.SubGridSize, model.SubGridSize, !style.Transparent)
}

func (svc *OccupancyIndexer) fillComplexCell(occ *stateful.OccupancyIndex, x, y, z int, style model.StyledBlock) {
	occluding := !style.Transparent
	for _, elem := range style.Elements {
		xFrom := int(elem.From[0]) / model.SubGridSize
		xTo := int(elem.To[0]) / model.SubGridSize
		yFrom := int(elem.From[1]) / model.SubGridSize
		yTo := int(elem.To[1]) / model.SubGridSize
		zFrom := int(elem.From[2]) / model.SubGridSize
		zTo := int(elem.To[2]) / model.SubGridSize

		w := xTo - xFrom
		h := yTo - yFrom
		d := zTo - zFrom
		if w <= 0 || h <= 0 || d <= 0 {
			continue
		}

		occ.FillRegion(
			x*model.SubGridSize+xFrom,
			y*model.SubGridSize+yFrom,
			z*model.SubGridSize+zFrom,
			w, h, d,
			occluding,
		)
	}
}
