package pipeline

import (
	"errors"
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
	var errs []error
	for _, name := range files {
		rx, rz := svc.parseRegionCoord(name)

		if !svc.rangeOverlap(rx*512, rx*512+511, bounds.XMin, bounds.XMax) ||
			!svc.rangeOverlap(rz*512, rz*512+511, bounds.ZMin, bounds.ZMax) {
			continue
		}

		regionBlocks, err := svc.processRegion(input, name, rx, rz, bounds)
		if err != nil {
			errs = append(errs, err)
		}
		blocks = append(blocks, regionBlocks...)
	}

	if len(errs) > 0 {
		return blocks, fmt.Errorf("region reader: %w", errors.Join(errs...))
	}
	return blocks, nil
}

func (svc *RegionReader) listRegionFiles(input string) ([]string, error) {
	entries, err := svc.fs.ReadDir(input)
	if err != nil {
		return nil, fmt.Errorf("list region files in %s: %w", input, err)
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

func (svc *RegionReader) processRegion(input, name string, rx, rz int, bounds model.Bounds) ([]model.RawBlock, error) {
	r, err := region.Open(filepath.Join(input, name))
	if err != nil {
		return nil, fmt.Errorf("open region %s: %w", name, err)
	}
	defer r.Close()

	var blocks []model.RawBlock
	var errs []error
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

			chunkBlocks, err := svc.processChunk(r, lx, lz, chunkX, chunkZ, bounds)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			blocks = append(blocks, chunkBlocks...)
		}
	}
	if len(errs) > 0 {
		return blocks, fmt.Errorf("process region %s: %w", name, errors.Join(errs...))
	}
	return blocks, nil
}

func (svc *RegionReader) processChunk(r *region.Region, lx, lz, chunkX, chunkZ int, bounds model.Bounds) ([]model.RawBlock, error) {
	data, err := r.ReadSector(lx, lz)
	if err != nil {
		return nil, fmt.Errorf("read sector for chunk (%d, %d): %w", chunkX, chunkZ, err)
	}

	blocks, err := svc.chunkDecoder.Run(data, chunkX, chunkZ)
	if err != nil {
		return nil, err
	}

	filtered := blocks[:0]
	for _, b := range blocks {
		if b.X >= bounds.XMin && b.X <= bounds.XMax &&
			b.Y >= bounds.YMin && b.Y <= bounds.YMax &&
			b.Z >= bounds.ZMin && b.Z <= bounds.ZMax {
			filtered = append(filtered, b)
		}
	}
	return filtered, nil
}

func (svc *RegionReader) parseRegionCoord(name string) (int, int) {
	var rx, rz int
	_, _ = fmt.Sscanf(name, "r.%d.%d.mca", &rx, &rz)
	return rx, rz
}

func (svc *RegionReader) rangeOverlap(min1, max1, min2, max2 int) bool {
	return max(min1, min2) <= min(max1, max2)
}
