package matcher

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBrightnessMatcher_New(t *testing.T) {
	t.Parallel()

	t.Run("valid config", func(t *testing.T) {
		mock := &mockFS{data: []byte("brightness:\n  Slate: 1.3\n  Wood: 1.15\n")}
		m, err := NewBrightnessMatcher(mock, "test.yaml")
		require.NoError(t, err)
		require.NotNil(t, m)
	})

	t.Run("missing brightness section", func(t *testing.T) {
		mock := &mockFS{data: []byte("mappings:\n  stone: Slate\n")}
		m, err := NewBrightnessMatcher(mock, "test.yaml")
		require.NoError(t, err)
		require.NotNil(t, m)
	})

	t.Run("invalid yaml", func(t *testing.T) {
		mock := &mockFS{data: []byte("{{{")}
		_, err := NewBrightnessMatcher(mock, "test.yaml")
		require.Error(t, err)
	})

	t.Run("file not found", func(t *testing.T) {
		_, err := NewBrightnessMatcher(&mockFS{data: nil}, "test.yaml")
		require.Error(t, err)
	})
}

func TestBrightnessMatcher_Run(t *testing.T) {
	t.Parallel()

	mock := &mockFS{data: []byte(`
mappings:
  stone: Slate

brightness:
  SmoothPlastic: 1.0
  Wood: 1.15
  Slate: 1.3
  Cobblestone: 1.3
`)}
	m, err := NewBrightnessMatcher(mock, "test.yaml")
	require.NoError(t, err)

	tests := []struct {
		name     string
		material string
		want     float64
	}{
		{"smooth plastic", "SmoothPlastic", 1.0},
		{"wood", "Wood", 1.15},
		{"slate", "Slate", 1.3},
		{"cobblestone", "Cobblestone", 1.3},
		{"unlisted material default", "Glass", 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, m.Run(tt.material))
		})
	}
}
