package pipeline

import (
	"fmt"
	"path/filepath"

	"mc2lua/internal/model"

	"github.com/Tnze/go-mc/save/region"
)

type chunkDecoder interface {
	Run(data []byte, chunkX, chunkZ int) ([]model.RawBlock, error)
}

type RegionReader struct {
	fs           fsApi
	chunkDecoder chunkDecoder
}

func NewRegionReader(fs fsApi, decoder chunkDecoder) *RegionReader {
	return &RegionReader{fs: fs, chunkDecoder: decoder}
}

func (svc *RegionReader) Run(input string, bounds model.Bounds) ([]model.RawBlock, error) {
	files, err := svc.listRegionFiles(input)
	if err != nil {
		return nil, fmt.Errorf("list region files: %w", err)
	}

	var blocks []model.RawBlock
	for _, name := range files {
		rx, rz := svc.parseRegionCoord(name)

		if !svc.rangeOverlap(rx*512, rx*512+511, bounds.XMin, bounds.XMax) ||
			!svc.rangeOverlap(rz*512, rz*512+511, bounds.ZMin, bounds.ZMax) {
			continue
		}

		regionBlocks := svc.processRegion(input, name, rx, rz, bounds)
		blocks = append(blocks, regionBlocks...)
	}

	return blocks, nil
}

func (svc *RegionReader) listRegionFiles(input string) ([]string, error) {
	entries, err := svc.fs.ReadDir(input)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		ok, _ := filepath.Match("r.*.*.mca", name)
		if ok {
			files = append(files, name)
		}
	}
	return files, nil
}

func (svc *RegionReader) processRegion(input, name string, rx, rz int, bounds model.Bounds) []model.RawBlock {
	r, err := region.Open(filepath.Join(input, name))
	if err != nil {
		return nil
	}
	defer r.Close()

	var blocks []model.RawBlock
	for lx := 0; lx < 32; lx++ {
		for lz := 0; lz < 32; lz++ {
			if !r.ExistSector(lx, lz) {
				continue
			}

			chunkX := rx*32 + lx
			chunkZ := rz*32 + lz

			if !svc.rangeOverlap(chunkX*16, chunkX*16+15, bounds.XMin, bounds.XMax) ||
				!svc.rangeOverlap(chunkZ*16, chunkZ*16+15, bounds.ZMin, bounds.ZMax) {
				continue
			}

			chunkBlocks := svc.processChunk(r, lx, lz, chunkX, chunkZ, bounds)
			blocks = append(blocks, chunkBlocks...)
		}
	}
	return blocks
}

func (svc *RegionReader) processChunk(r *region.Region, lx, lz, chunkX, chunkZ int, bounds model.Bounds) []model.RawBlock {
	data, err := r.ReadSector(lx, lz)
	if err != nil {
		return nil
	}

	blocks, err := svc.chunkDecoder.Run(data, chunkX, chunkZ)
	if err != nil {
		return nil
	}

	filtered := blocks[:0]
	for _, b := range blocks {
		if b.X >= bounds.XMin && b.X <= bounds.XMax &&
			b.Y >= bounds.YMin && b.Y <= bounds.YMax &&
			b.Z >= bounds.ZMin && b.Z <= bounds.ZMax {
			filtered = append(filtered, b)
		}
	}
	return filtered
}

func (svc *RegionReader) parseRegionCoord(name string) (int, int) {
	var rx, rz int
	_, _ = fmt.Sscanf(name, "r.%d.%d.mca", &rx, &rz)
	return rx, rz
}

func (svc *RegionReader) rangeOverlap(min1, max1, min2, max2 int) bool {
	return max(min1, min2) <= min(max1, max2)
}
