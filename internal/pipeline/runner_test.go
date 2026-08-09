package pipeline

import (
	"bytes"
	"errors"
	"io"
	"math"
	"strings"
	"testing"

	"mc2lua/internal/stateful"
	"mc2lua/internal/model"

	"github.com/stretchr/testify/require"
)

type mockRegionReader struct {
	runFn func(input string, bounds model.Bounds) ([]model.RawBlock, error)
}

func (m *mockRegionReader) Run(input string, bounds model.Bounds) ([]model.RawBlock, error) {
	return m.runFn(input, bounds)
}

type mockCoordNormalizer struct {
	runFn func(blocks []model.RawBlock, noOffset bool) ([]model.RawBlock, error)
}

func (m *mockCoordNormalizer) Run(blocks []model.RawBlock, noOffset bool) ([]model.RawBlock, error) {
	return m.runFn(blocks, noOffset)
}

type mockBoundsResolver struct {
	runFn func(blocks []model.RawBlock) model.Bounds
}

func (m *mockBoundsResolver) Run(blocks []model.RawBlock) model.Bounds {
	if m.runFn != nil {
		return m.runFn(blocks)
	}
	return model.Bounds{}
}

type mockAssetScanner struct {
	runFn func(root string) (map[string][]string, error)
}

func (m *mockAssetScanner) Run(root string) (map[string][]string, error) {
	if m.runFn != nil {
		return m.runFn(root)
	}
	return map[string][]string{}, nil
}

type mockCollector struct {
	runFn func(blocks []model.RawBlock, namespaces map[string][]string) ([]model.ResolvedBlock, map[string]string)
}

func (m *mockCollector) Run(blocks []model.RawBlock, namespaces map[string][]string) ([]model.ResolvedBlock, map[string]string) {
	if m.runFn != nil {
		return m.runFn(blocks, namespaces)
	}
	return nil, nil
}

type mockIndexer struct {
	runFn func(blocks []model.ResolvedBlock, namespaces map[string][]string) *stateful.StyleIndex
}

func (m *mockIndexer) Run(blocks []model.ResolvedBlock, namespaces map[string][]string) *stateful.StyleIndex {
	if m.runFn != nil {
		return m.runFn(blocks, namespaces)
	}
	return stateful.NewStyleIndex()
}

type mockBlockVoxelIndexer struct {
	runFn func(blocks []model.RawBlock, styles stateful.StyleIndex) *stateful.VoxelIndex
}

func (m *mockBlockVoxelIndexer) Run(blocks []model.RawBlock, styles stateful.StyleIndex) *stateful.VoxelIndex {
	if m.runFn != nil {
		return m.runFn(blocks, styles)
	}
	return stateful.NewVoxelIndex()
}

type mockMicroVoxelIndexer struct {
	runFn func(blocks []model.RawBlock, styles stateful.StyleIndex) *stateful.VoxelIndex
}

func (m *mockMicroVoxelIndexer) Run(blocks []model.RawBlock, styles stateful.StyleIndex) *stateful.VoxelIndex {
	if m.runFn != nil {
		return m.runFn(blocks, styles)
	}
	return stateful.NewVoxelIndex()
}

type mockMerger struct {
	runFn func(grid *stateful.VoxelIndex) []model.Cuboid
}

func (m *mockMerger) Run(grid *stateful.VoxelIndex) []model.Cuboid {
	if m.runFn != nil {
		return m.runFn(grid)
	}
	return nil
}

type mockLuaGenerator struct {
	runFn func(blocks []model.RawBlock, blockCuboids, microCuboids []model.Cuboid, styleIndex stateful.StyleIndex, scale float64, outputPath string) (int, error)
}

func (m *mockLuaGenerator) Run(blocks []model.RawBlock, blockCuboids, microCuboids []model.Cuboid, styleIndex stateful.StyleIndex, scale float64, outputPath string) (int, error) {
	if m.runFn != nil {
		return m.runFn(blocks, blockCuboids, microCuboids, styleIndex, scale, outputPath)
	}
	return 0, nil
}

func TestRunner_New(t *testing.T) {
	t.Parallel()

	mockRR := &mockRegionReader{}
	mockBR := &mockBoundsResolver{}
	mockCN := &mockCoordNormalizer{}
	mockAS := &mockAssetScanner{}
	mockCol := &mockCollector{}
	mockIdx := &mockIndexer{}
	mockBVI := &mockBlockVoxelIndexer{}
	mockMVI := &mockMicroVoxelIndexer{}
	mockRM := &mockMerger{}
	mockLG := &mockLuaGenerator{}
	r := NewRunner(mockRR, mockBR, mockCN, mockAS, mockCol, mockIdx, mockBVI, mockMVI, mockRM, mockLG, io.Discard)
	require.NotNil(t, r)
}

func TestRunner_Run(t *testing.T) {
	t.Parallel()

	errWorld := errors.New("world error")
	errNormalize := errors.New("normalize error")

	tests := []struct {
		name          string
		mockRun       func(input string, bounds model.Bounds) ([]model.RawBlock, error)
		mockNormalize func(blocks []model.RawBlock, noOffset bool) ([]model.RawBlock, error)
		mockScan      func(root string) (map[string][]string, error)
		mockCollect   func(blocks []model.RawBlock, namespaces map[string][]string) ([]model.ResolvedBlock, map[string]string)
		mockBVI       func(blocks []model.RawBlock, styles stateful.StyleIndex) *stateful.VoxelIndex
		mockRM        func(grid *stateful.VoxelIndex) []model.Cuboid
		mockLG        func(blocks []model.RawBlock, blockCuboids, microCuboids []model.Cuboid, styleIndex stateful.StyleIndex, scale float64, outputPath string) (int, error)
		wantErr       bool
		wantErrMsg    string
	}{
		{
			name: "success",
			mockRun: func(input string, bounds model.Bounds) ([]model.RawBlock, error) {
				return []model.RawBlock{}, nil
			},
			mockNormalize: func(blocks []model.RawBlock, noOffset bool) ([]model.RawBlock, error) {
				return blocks, nil
			},
		},
		{
			name: "region reader error is logged not returned",
			mockRun: func(input string, bounds model.Bounds) ([]model.RawBlock, error) {
				return nil, errWorld
			},
			mockNormalize: func(blocks []model.RawBlock, noOffset bool) ([]model.RawBlock, error) {
				return blocks, nil
			},
		},
		{
			name: "coord normalizer error",
			mockRun: func(input string, bounds model.Bounds) ([]model.RawBlock, error) {
				return []model.RawBlock{}, nil
			},
			mockNormalize: func(blocks []model.RawBlock, noOffset bool) ([]model.RawBlock, error) {
				return nil, errNormalize
			},
			wantErr:    true,
			wantErrMsg: "normalize coords: normalize error",
		},
		{
			name: "lua generator error",
			mockRun: func(input string, bounds model.Bounds) ([]model.RawBlock, error) {
				return []model.RawBlock{{}}, nil
			},
			mockNormalize: func(blocks []model.RawBlock, noOffset bool) ([]model.RawBlock, error) {
				return blocks, nil
			},
			mockLG: func(blocks []model.RawBlock, blockCuboids, microCuboids []model.Cuboid, styleIndex stateful.StyleIndex, scale float64, outputPath string) (int, error) {
				return 0, errors.New("write failed")
			},
			wantErr:    true,
			wantErrMsg: "generate lua: write failed",
		},
		{
			name: "log output written on success",
			mockRun: func(input string, bounds model.Bounds) ([]model.RawBlock, error) {
				return []model.RawBlock{{ID: "minecraft:stone", X: 0, Y: 0, Z: 0}}, nil
			},
			mockNormalize: func(blocks []model.RawBlock, noOffset bool) ([]model.RawBlock, error) {
				return blocks, nil
			},
		},
		{
			name: "asset scanner error",
			mockRun: func(input string, bounds model.Bounds) ([]model.RawBlock, error) {
				return []model.RawBlock{}, nil
			},
			mockNormalize: func(blocks []model.RawBlock, noOffset bool) ([]model.RawBlock, error) {
				return blocks, nil
			},
			mockScan: func(root string) (map[string][]string, error) {
				return nil, errors.New("scan failed")
			},
			wantErr:    true,
			wantErrMsg: "scan assets: scan failed",
		},
		{
			name: "with non-empty namespaces, unresolved, and merge stats",
			mockRun: func(input string, bounds model.Bounds) ([]model.RawBlock, error) {
				return []model.RawBlock{{ID: "minecraft:stone"}}, nil
			},
			mockNormalize: func(blocks []model.RawBlock, noOffset bool) ([]model.RawBlock, error) {
				return blocks, nil
			},
			mockScan: func(root string) (map[string][]string, error) {
				return map[string][]string{"minecraft": {"assets/minecraft"}, "cobblemon": {"assets/cobblemon"}}, nil
			},
			mockCollect: func(blocks []model.RawBlock, namespaces map[string][]string) ([]model.ResolvedBlock, map[string]string) {
				return []model.ResolvedBlock{{ID: "minecraft:stone"}}, map[string]string{"minecraft:dirt": "not found"}
			},
			mockBVI: func(blocks []model.RawBlock, styles stateful.StyleIndex) *stateful.VoxelIndex {
				grid := stateful.NewVoxelIndex()
				grid.AddBlock(&model.MergedBlock{ID: "minecraft:stone", X: 0, Y: 0, Z: 0})
				return grid
			},
			mockRM: func(grid *stateful.VoxelIndex) []model.Cuboid {
				return []model.Cuboid{{ID: "minecraft:stone", Width: 1, Depth: 1, Height: 1}}
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockRR := &mockRegionReader{runFn: tt.mockRun}
			mockBR := &mockBoundsResolver{}
			mockCN := &mockCoordNormalizer{runFn: tt.mockNormalize}
			mockAS := &mockAssetScanner{}
			if tt.mockScan != nil {
				mockAS.runFn = tt.mockScan
			}
			mockCol := &mockCollector{}
			if tt.mockCollect != nil {
				mockCol.runFn = tt.mockCollect
			}
			mockIdx := &mockIndexer{}
			mockBVI := &mockBlockVoxelIndexer{}
			if tt.mockBVI != nil {
				mockBVI.runFn = tt.mockBVI
			}
			mockMVI := &mockMicroVoxelIndexer{}
			mockRM := &mockMerger{}
			if tt.mockRM != nil {
				mockRM.runFn = tt.mockRM
			}
			mockLG := &mockLuaGenerator{}
			if tt.mockLG != nil {
				mockLG.runFn = tt.mockLG
			}
			r := NewRunner(mockRR, mockBR, mockCN, mockAS, mockCol, mockIdx, mockBVI, mockMVI, mockRM, mockLG, io.Discard)

			err := r.Run(RunConfig{
				Input:     "/test",
				AssetsDir: "assets",
				Bounds:    model.Bounds{XMin: 0, XMax: 10, YMin: 0, YMax: 10, ZMin: 0, ZMax: 10},
			})
			if tt.wantErr {
				require.Error(t, err)
				if tt.wantErrMsg != "" {
					require.Contains(t, err.Error(), tt.wantErrMsg)
				}
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestRunner_RunLogsArea(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		bounds         model.Bounds
		blocks         []model.RawBlock
		runFn          func(blocks []model.RawBlock) model.Bounds
		wantRequested  bool
		wantActualLine string
	}{
		{
			name:   "requested and actual bounds logged",
			bounds: model.Bounds{XMin: -10, XMax: 10, YMin: 0, YMax: 255, ZMin: -100, ZMax: 100},
			blocks: []model.RawBlock{
				{X: -5, Y: 0, Z: -20},
				{X: 7, Y: 200, Z: 40},
			},
			wantRequested:  true,
			wantActualLine: "Actual bounds: x [-5..7], y [0..200], z [-20..40]",
		},
		{
			name:   "requested bounds not logged when unbounded",
			bounds: model.Bounds{
				XMin: math.MinInt32, XMax: math.MaxInt32,
				YMin: math.MinInt32, YMax: math.MaxInt32,
				ZMin: math.MinInt32, ZMax: math.MaxInt32,
			},
			blocks: []model.RawBlock{
				{X: 1, Y: 2, Z: 3},
			},
			wantRequested:  false,
			wantActualLine: "Actual bounds: x [1..1], y [2..2], z [3..3]",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			r := NewRunner(
				&mockRegionReader{runFn: func(input string, bounds model.Bounds) ([]model.RawBlock, error) {
					return tt.blocks, nil
				}},
				NewBoundsResolver(),
				&mockCoordNormalizer{runFn: func(blocks []model.RawBlock, noOffset bool) ([]model.RawBlock, error) {
					return blocks, nil
				}},
				&mockAssetScanner{},
				&mockCollector{},
				&mockIndexer{},
				&mockBlockVoxelIndexer{},
				&mockMicroVoxelIndexer{},
				&mockMerger{},
				&mockLuaGenerator{},
				&buf,
			)

			err := r.Run(RunConfig{Input: "/test", AssetsDir: "assets", Bounds: tt.bounds})
			require.NoError(t, err)

			output := buf.String()
			require.Equal(t, tt.wantRequested, strings.Contains(output, "Requested bounds:"))
			require.Contains(t, output, tt.wantActualLine)
		})
	}
}
