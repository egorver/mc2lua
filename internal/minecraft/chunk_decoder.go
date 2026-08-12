package minecraft

import (
	"fmt"
	"math/bits"

	"mc2lua/internal/model"

	"github.com/Tnze/go-mc/level"
	"github.com/Tnze/go-mc/save"
)

const (
	minBitsPerEntry           = 4
	chunkStatusFull           = "full"
	chunkStatusFullNamespaced = "minecraft:full"
)

var airBlockIDs = map[string]bool{
	"minecraft:air":      true,
	"minecraft:cave_air": true,
	"minecraft:void_air": true,
}

type ChunkDecoder struct{}

func NewChunkDecoder() *ChunkDecoder {
	return &ChunkDecoder{}
}

func (svc *ChunkDecoder) Run(data []byte, chunkX, chunkZ int) ([]model.RawBlock, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("decode chunk (%d, %d): empty data", chunkX, chunkZ)
	}
	var sc save.Chunk
	if err := sc.Load(data); err != nil {
		return nil, fmt.Errorf("decode chunk (%d, %d): %w", chunkX, chunkZ, err)
	}

	if sc.Status != chunkStatusFull && sc.Status != chunkStatusFullNamespaced {
		return nil, nil
	}

	var blocks []model.RawBlock
	for _, ssec := range sc.Sections {
		sectionBlocks := svc.decodeSection(ssec, chunkX, chunkZ)
		blocks = append(blocks, sectionBlocks...)
	}
	return blocks, nil
}

func (svc *ChunkDecoder) decodeSection(ssec save.Section, chunkX, chunkZ int) []model.RawBlock {
	if len(ssec.BlockStates.Palette) == 0 {
		return nil
	}

	baseY := int(ssec.Y) * BlocksPerChunkSection
	bs := svc.newBitStorage(len(ssec.BlockStates.Palette), ssec.BlockStates.Data)
	palette := ssec.BlockStates.Palette

	var blocks []model.RawBlock
	for j := 0; j < BlocksPerSection; j++ {
		block := svc.decodeBlock(bs, j, palette, chunkX, chunkZ, baseY)
		if block != nil {
			blocks = append(blocks, *block)
		}
	}
	return blocks
}

func (svc *ChunkDecoder) newBitStorage(palSize int, data []uint64) *level.BitStorage {
	if palSize <= 1 {
		return level.NewBitStorage(0, BlocksPerSection, nil)
	}
	bitsPerEntry := bits.Len(uint(palSize - 1)) //nolint:gosec
	if bitsPerEntry < minBitsPerEntry {
		bitsPerEntry = minBitsPerEntry
	}
	return level.NewBitStorage(bitsPerEntry, BlocksPerSection, data)
}

func (svc *ChunkDecoder) decodeBlock(bs *level.BitStorage, j int, palette []save.BlockState, chunkX, chunkZ, baseY int) *model.RawBlock {
	idx := bs.Get(j)
	if idx >= len(palette) {
		return nil
	}

	blockName := palette[idx].Name
	if airBlockIDs[blockName] {
		return nil
	}

	var props map[string]string
	if p := palette[idx].Properties; p.Type != 0 {
		var m map[string]string
		if err := p.Unmarshal(&m); err == nil && len(m) > 0 {
			props = m
		}
	}

	x := j & ChunkSectionXMask
	z := (j >> ChunkSectionZShift) & ChunkSectionXMask
	y := j >> ChunkSectionYShift

	return &model.RawBlock{
		ID:    blockName,
		Props: props,
		X:     chunkX*BlocksPerChunkSection + x,
		Y:     baseY + y,
		Z:     chunkZ*BlocksPerChunkSection + z,
	}
}
