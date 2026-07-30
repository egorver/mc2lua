package pipeline

import (
	"errors"
	"io"
	"testing"

	"mc2lua/internal/index"
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
	runFn func(blocks []model.ResolvedBlock) *index.StyleIndex
}

func (m *mockIndexer) Run(blocks []model.ResolvedBlock) *index.StyleIndex {
	if m.runFn != nil {
		return m.runFn(blocks)
	}
	return index.NewStyleIndex()
}

type mockVoxelIndexer struct {
	runFn func(blocks []model.RawBlock, styles index.StyleIndex) *index.VoxelIndex
}

func (m *mockVoxelIndexer) Run(blocks []model.RawBlock, styles index.StyleIndex) *index.VoxelIndex {
	if m.runFn != nil {
		return m.runFn(blocks, styles)
	}
	return index.NewVoxelIndex()
}

type mockMerger struct {
	runFn func(grid *index.VoxelIndex) []model.Cuboid
}

func (m *mockMerger) Run(grid *index.VoxelIndex) []model.Cuboid {
	if m.runFn != nil {
		return m.runFn(grid)
	}
	return nil
}

type mockLuaGenerator struct {
	runFn func(blocks []model.RawBlock, cuboids []model.Cuboid, styleIndex index.StyleIndex, scale float64, outputPath string) (int, error)
}

func (m *mockLuaGenerator) Run(blocks []model.RawBlock, cuboids []model.Cuboid, styleIndex index.StyleIndex, scale float64, outputPath string) (int, error) {
	if m.runFn != nil {
		return m.runFn(blocks, cuboids, styleIndex, scale, outputPath)
	}
	return 0, nil
}

func TestRunner_New(t *testing.T) {
	t.Parallel()

	mockRR := &mockRegionReader{}
	mockCN := &mockCoordNormalizer{}
	mockAS := &mockAssetScanner{}
	mockCol := &mockCollector{}
	mockIdx := &mockIndexer{}
	mockVI := &mockVoxelIndexer{}
	mockRM := &mockMerger{}
	mockLG := &mockLuaGenerator{}
	r := NewRunner(mockRR, mockCN, mockAS, mockCol, mockIdx, mockVI, mockRM, mockLG, io.Discard)
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
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockRR := &mockRegionReader{runFn: tt.mockRun}
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
			mockVI := &mockVoxelIndexer{}
			mockRM := &mockMerger{}
			mockLG := &mockLuaGenerator{}
			r := NewRunner(mockRR, mockCN, mockAS, mockCol, mockIdx, mockVI, mockRM, mockLG, io.Discard)

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
