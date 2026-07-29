package index

import "mc2lua/internal/model"

type VoxelIndex struct {
	blocks []*model.MergedBlock
	lookup map[[3]int]*model.MergedBlock
}

func NewVoxelIndex() *VoxelIndex {
	return &VoxelIndex{
		lookup: make(map[[3]int]*model.MergedBlock),
	}
}

func (vg *VoxelIndex) AddBlock(b *model.MergedBlock) {
	vg.blocks = append(vg.blocks, b)
	vg.lookup[[3]int{b.X, b.Y, b.Z}] = b
}

func (vg *VoxelIndex) GetBlock(x, y, z int) *model.MergedBlock {
	return vg.lookup[[3]int{x, y, z}]
}

func (vg *VoxelIndex) Blocks() []*model.MergedBlock {
	return vg.blocks
}
