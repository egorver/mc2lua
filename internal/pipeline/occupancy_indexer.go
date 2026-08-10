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
		svc.fillCell(occ, b.X, b.Y, b.Z, style)
	}

	return occ
}

func (svc *OccupancyIndexer) fillCell(occ *stateful.OccupancyIndex, x, y, z int, style model.StyledBlock) {
	occ.FillRegion(x*model.SubGridSize, y*model.SubGridSize, z*model.SubGridSize,
		model.SubGridSize, model.SubGridSize, model.SubGridSize, !style.Transparent)
}
