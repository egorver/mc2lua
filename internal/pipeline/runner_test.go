package pipeline

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"mc2lua/internal/model"
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

func TestRunner_New(t *testing.T) {
	t.Parallel()

	mockRR := &mockRegionReader{}
	mockCN := &mockCoordNormalizer{}
	r := NewRunner(mockRR, mockCN)
	require.NotNil(t, r)
}

func TestRunner_Run(t *testing.T) {
	t.Parallel()

	errWorld := errors.New("world error")
	errNormalize := errors.New("normalize error")

	tests := []struct {
		name            string
		mockRun         func(input string, bounds model.Bounds) ([]model.Block, error)
		mockNormalize   func(blocks []model.Block, noOffset bool) ([]model.Block, error)
		wantErr         bool
		wantErrMsg      string
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
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockRR := &mockRegionReader{runFn: tt.mockRun}
			mockCN := &mockCoordNormalizer{runFn: tt.mockNormalize}
			r := NewRunner(mockRR, mockCN)

			err := r.Run(RunConfig{
				Input:  "/test",
				Bounds: model.Bounds{XMin: 0, XMax: 10, YMin: 0, YMax: 10, ZMin: 0, ZMax: 10},
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
