package pipeline

import (
	"mc2lua/internal/model"
	"mc2lua/internal/stateful"
)

type microIndexerPropsKeyBuilder interface {
	Run(props map[string]string) string
}

type MicroVoxelIndexer struct {
	propsKeyBuilder microIndexerPropsKeyBuilder
}

func NewMicroVoxelIndexer(pkb microIndexerPropsKeyBuilder) *MicroVoxelIndexer {
	return &MicroVoxelIndexer{propsKeyBuilder: pkb}
}

func (svc *MicroVoxelIndexer) Run(blocks []model.RawBlock, styles stateful.StyleIndex) *stateful.VoxelIndex {
	grid := stateful.NewVoxelIndex()
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

func (svc *MicroVoxelIndexer) addMicroBlocks(grid *stateful.VoxelIndex, b model.RawBlock, style model.StyledBlock) {
	seen := make(map[[3]int]bool)
	for _, elem := range style.Elements {
		gxFrom := int(elem.From[0]) / model.SubGridSize
		gxTo := int(elem.To[0]) / model.SubGridSize
		gyFrom := int(elem.From[1]) / model.SubGridSize
		gyTo := int(elem.To[1]) / model.SubGridSize
		gzFrom := int(elem.From[2]) / model.SubGridSize
		gzTo := int(elem.To[2]) / model.SubGridSize

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
						X:        b.X*model.SubGridSize + gx,
						Y:        b.Y*model.SubGridSize + gy,
						Z:        b.Z*model.SubGridSize + gz,
						Merged:   false,
					})
				}
			}
		}
	}
}
