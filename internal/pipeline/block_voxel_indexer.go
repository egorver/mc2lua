package pipeline

import (
	"mc2lua/internal/stateful"
	"mc2lua/internal/model"
)

type indexerPropsKeyBuilder interface {
	Run(props map[string]string) string
}

type BlockVoxelIndexer struct {
	propsKeyBuilder indexerPropsKeyBuilder
}

func NewBlockVoxelIndexer(pkb indexerPropsKeyBuilder) *BlockVoxelIndexer {
	return &BlockVoxelIndexer{propsKeyBuilder: pkb}
}

func (svc *BlockVoxelIndexer) Run(blocks []model.RawBlock, styles stateful.StyleIndex) *stateful.VoxelIndex {
	grid := stateful.NewVoxelIndex()
	for _, b := range blocks {
		propsKey := svc.propsKeyBuilder.Run(b.Props)
		resolved, ok := styles.Get(b.ID, propsKey)
		if !ok || resolved.GridAlignment != model.GridFullBlock {
			continue
		}
		grid.AddBlock(&model.MergedBlock{
			ID:       b.ID,
			PropsKey: resolved.PropsKey,
			X:        b.X, Y: b.Y, Z: b.Z,
			Merged: false,
		})
	}
	return grid
}
