package pipeline

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"mc2lua/internal/model"
)

type mockWorldReader struct {
	runFn func(input string, bounds model.Bounds) (*model.World, error)
}

func (m *mockWorldReader) Run(input string, bounds model.Bounds) (*model.World, error) {
	return m.runFn(input, bounds)
}

func TestRunner_New(t *testing.T) {
	t.Parallel()

	mock := &mockWorldReader{}
	r := NewRunner(mock)
	require.NotNil(t, r)
}

func TestRunner_Run(t *testing.T) {
	t.Parallel()

	errWorld := errors.New("world error")

	tests := []struct {
		name        string
		mockRun     func(input string, bounds model.Bounds) (*model.World, error)
		wantErr     bool
		wantErrMsg  string
	}{
		{
			name: "success",
			mockRun: func(input string, bounds model.Bounds) (*model.World, error) {
				return &model.World{}, nil
			},
		},
		{
			name: "world reader error",
			mockRun: func(input string, bounds model.Bounds) (*model.World, error) {
				return nil, errWorld
			},
			wantErr:    true,
			wantErrMsg: "read world: world error",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mock := &mockWorldReader{runFn: tt.mockRun}
			r := NewRunner(mock)

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
