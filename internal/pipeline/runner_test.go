package pipeline

import (
	"errors"
	"testing"

	"mc2lua/internal/index"
	"mc2lua/internal/model"

	"github.com/stretchr/testify/require"
)

type mockRegionReader struct {
	runFn func(input string, bounds model.Bounds) ([]model.Block, error)
}

func (m *mockRegionReader) Run(input string, bounds model.Bounds) ([]model.Block, error) {
	return m.runFn(input, bounds)
}

type mockCoordNormalizer struct {
	runFn func(blocks []model.Block, noOffset bool) ([]model.Block, error)
}

func (m *mockCoordNormalizer) Run(blocks []model.Block, noOffset bool) ([]model.Block, error) {
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

type mockIndexBuilder struct {
	runFn func(blocks []model.Block, namespaces map[string][]string) (*index.BlockIndex, error)
}

func (m *mockIndexBuilder) Run(blocks []model.Block, namespaces map[string][]string) (*index.BlockIndex, map[string]string, error) {
	if m.runFn != nil {
		idx, err := m.runFn(blocks, namespaces)
		return idx, nil, err
	}
	return index.NewBlockIndex(), nil, nil
}

func TestRunner_New(t *testing.T) {
	t.Parallel()

	mockRR := &mockRegionReader{}
	mockCN := &mockCoordNormalizer{}
	mockAS := &mockAssetScanner{}
	mockIB := &mockIndexBuilder{}
	r := NewRunner(mockRR, mockCN, mockAS, mockIB)
	require.NotNil(t, r)
}

func TestRunner_Run(t *testing.T) {
	t.Parallel()

	errWorld := errors.New("world error")
	errNormalize := errors.New("normalize error")

	tests := []struct {
		name           string
		mockRun        func(input string, bounds model.Bounds) ([]model.Block, error)
		mockNormalize  func(blocks []model.Block, noOffset bool) ([]model.Block, error)
		mockAssetScan  func(root string) (map[string][]string, error)
		mockIndexBuild func(blocks []model.Block, namespaces map[string][]string) (*index.BlockIndex, error)
		wantErr        bool
		wantErrMsg     string
	}{
		{
			name: "success",
			mockRun: func(input string, bounds model.Bounds) ([]model.Block, error) {
				return []model.Block{}, nil
			},
			mockNormalize: func(blocks []model.Block, noOffset bool) ([]model.Block, error) {
				return blocks, nil
			},
		},
		{
			name: "region reader error",
			mockRun: func(input string, bounds model.Bounds) ([]model.Block, error) {
				return nil, errWorld
			},
			mockNormalize: func(blocks []model.Block, noOffset bool) ([]model.Block, error) {
				return blocks, nil
			},
			wantErr:    true,
			wantErrMsg: "read world: world error",
		},
		{
			name: "coord normalizer error",
			mockRun: func(input string, bounds model.Bounds) ([]model.Block, error) {
				return []model.Block{}, nil
			},
			mockNormalize: func(blocks []model.Block, noOffset bool) ([]model.Block, error) {
				return nil, errNormalize
			},
			wantErr:    true,
			wantErrMsg: "normalize coords: normalize error",
		},
		{
			name: "asset scanner error",
			mockRun: func(input string, bounds model.Bounds) ([]model.Block, error) {
				return []model.Block{}, nil
			},
			mockNormalize: func(blocks []model.Block, noOffset bool) ([]model.Block, error) {
				return blocks, nil
			},
			mockAssetScan: func(root string) (map[string][]string, error) {
				return nil, errors.New("scan failed")
			},
			wantErr:    true,
			wantErrMsg: "scan assets: scan failed",
		},
		{
			name: "index builder error",
			mockRun: func(input string, bounds model.Bounds) ([]model.Block, error) {
				return []model.Block{{ID: "minecraft:stone"}}, nil
			},
			mockNormalize: func(blocks []model.Block, noOffset bool) ([]model.Block, error) {
				return blocks, nil
			},
			mockIndexBuild: func(blocks []model.Block, namespaces map[string][]string) (*index.BlockIndex, error) {
				return nil, errors.New("build failed")
			},
			wantErr:    true,
			wantErrMsg: "build index: build failed",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockRR := &mockRegionReader{runFn: tt.mockRun}
			mockCN := &mockCoordNormalizer{runFn: tt.mockNormalize}
			mockAS := &mockAssetScanner{}
			if tt.mockAssetScan != nil {
				mockAS.runFn = tt.mockAssetScan
			}
			mockIB := &mockIndexBuilder{}
			if tt.mockIndexBuild != nil {
				mockIB.runFn = tt.mockIndexBuild
			}
			r := NewRunner(mockRR, mockCN, mockAS, mockIB)

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
