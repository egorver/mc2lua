package pipeline

import (
	"mc2lua/internal/index"
	"mc2lua/internal/model"
)

type indexerPropsKeyBuilder interface {
	Run(props map[string]string) string
}

type VoxelIndexer struct {
	propsKeyBuilder indexerPropsKeyBuilder
}

func NewVoxelIndexer(pkb indexerPropsKeyBuilder) *VoxelIndexer {
	return &VoxelIndexer{propsKeyBuilder: pkb}
}

func (svc *VoxelIndexer) Run(blocks []model.RawBlock, styles index.StyleIndex) *index.VoxelIndex {
	grid := index.NewVoxelIndex()
	for _, b := range blocks {
		propsKey := svc.propsKeyBuilder.Run(b.Props)
		resolved, ok := styles.Get(b.ID, propsKey)
		if !ok || !resolved.IsGridAligned {
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
