package pipeline

import (
	"mc2lua/internal/index"
	"mc2lua/internal/model"
)

type MicroVoxelIndexer struct {
	propsKeyBuilder indexerPropsKeyBuilder
}

func NewMicroVoxelIndexer(pkb indexerPropsKeyBuilder) *MicroVoxelIndexer {
	return &MicroVoxelIndexer{propsKeyBuilder: pkb}
}

func (svc *MicroVoxelIndexer) Run(blocks []model.RawBlock, styles index.StyleIndex) *index.VoxelIndex {
	grid := index.NewVoxelIndex()
	for _, b := range blocks {
		propsKey := svc.propsKeyBuilder.Run(b.Props)
		resolved, ok := styles.Get(b.ID, propsKey)
		if !ok || resolved.GridAlignment != model.GridSubBlock {
			continue
		}
		svc.addMicroBlocks(grid, b, resolved)
	}
	return grid
}

func (svc *MicroVoxelIndexer) addMicroBlocks(grid *index.VoxelIndex, b model.RawBlock, style model.StyledBlock) {
	seen := make(map[[3]int]bool)
	for _, elem := range style.Elements {
		gxFrom := int(elem.From[0]) / 4
		gxTo := int(elem.To[0]) / 4
		gyFrom := int(elem.From[1]) / 4
		gyTo := int(elem.To[1]) / 4
		gzFrom := int(elem.From[2]) / 4
		gzTo := int(elem.To[2]) / 4

		for gx := gxFrom; gx < gxTo; gx++ {
			for gy := gyFrom; gy < gyTo; gy++ {
				for gz := gzFrom; gz < gzTo; gz++ {
					key := [3]int{gx, gy, gz}
					if seen[key] {
						continue
					}
					seen[key] = true
					grid.AddBlock(&model.MergedBlock{
						ID:       b.ID,
						PropsKey: style.PropsKey,
						X:        b.X*4 + gx,
						Y:        b.Y*4 + gy,
						Z:        b.Z*4 + gz,
						Merged:   false,
					})
				}
			}
		}
	}
}
