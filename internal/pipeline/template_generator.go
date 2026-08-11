package pipeline

import (
	"fmt"
	"sort"
	"strings"

	"mc2lua/internal/model"
)

type blockColor struct {
	sum   [3]int
	count int
}

type TemplateGenerator struct {
	fs               fsApi
	partStyleMatcher partStyleMatcher
}

func NewTemplateGenerator(fs fsApi, psm partStyleMatcher) *TemplateGenerator {
	return &TemplateGenerator{fs: fs, partStyleMatcher: psm}
}

func (svc *TemplateGenerator) Run(parts []model.Part, blocks []model.RawBlock, outputPath string) error {
	colors := svc.collectColors(parts)
	popularity := svc.countPopularity(blocks)
	ids := svc.orderedIDs(colors, popularity)

	data := svc.buildYAML(ids, colors)
	return svc.writeFile(outputPath, []byte(data))
}

func (svc *TemplateGenerator) orderedIDs(colors map[string]*blockColor, popularity map[string]int) []string {
	ids := make([]string, 0, len(colors))
	for id := range colors {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool {
		pi, pj := popularity[ids[i]], popularity[ids[j]]
		if pi != pj {
			return pi > pj
		}
		return ids[i] < ids[j]
	})
	return ids
}

func (svc *TemplateGenerator) buildYAML(ids []string, colors map[string]*blockColor) string {
	var sb strings.Builder
	for _, id := range ids {
		if _, ok := svc.partStyleMatcher.Run(id); ok {
			continue
		}
		if sb.Len() == 0 {
			sb.WriteString("parts:\n")
		}
		avg := colors[id]
		fmt.Fprintf(&sb, "  %s:\n", id)
		fmt.Fprintf(&sb, "    color: [%d, %d, %d]\n", avg.sum[0]/avg.count, avg.sum[1]/avg.count, avg.sum[2]/avg.count)
	}
	if sb.Len() == 0 {
		sb.WriteString("parts: {}\n")
	}
	return sb.String()
}

func (svc *TemplateGenerator) writeFile(outputPath string, data []byte) error {
	w, err := svc.fs.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to write parts template file: %w", err)
	}
	if _, err := w.Write(data); err != nil {
		_ = w.Close()
		return fmt.Errorf("failed to write parts template file: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("failed to write parts template file: %w", err)
	}
	return nil
}

func (svc *TemplateGenerator) collectColors(parts []model.Part) map[string]*blockColor {
	colors := make(map[string]*blockColor)
	for _, p := range parts {
		bc, ok := colors[p.BlockID]
		if !ok {
			bc = &blockColor{}
			colors[p.BlockID] = bc
		}
		bc.sum[0] += int(p.Color[0])
		bc.sum[1] += int(p.Color[1])
		bc.sum[2] += int(p.Color[2])
		bc.count++
	}
	return colors
}

func (svc *TemplateGenerator) countPopularity(blocks []model.RawBlock) map[string]int {
	popularity := make(map[string]int)
	for _, b := range blocks {
		popularity[b.ID]++
	}
	return popularity
}
