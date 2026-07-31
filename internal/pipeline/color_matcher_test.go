package pipeline

import (
	"testing"

	"mc2lua/internal/model"

	"github.com/stretchr/testify/require"
)

func TestColorMatcher_New(t *testing.T) {
	t.Parallel()

	t.Run("valid config", func(t *testing.T) {
		mock := &mockFS{data: []byte("colors:\n  minecraft:water: [38, 94, 173]\n")}
		m, err := NewColorMatcher(mock, "test.yaml")
		require.NoError(t, err)
		require.NotNil(t, m)
	})

	t.Run("empty config", func(t *testing.T) {
		mock := &mockFS{data: []byte("")}
		m, err := NewColorMatcher(mock, "test.yaml")
		require.NoError(t, err)
		require.NotNil(t, m)
	})

	t.Run("invalid yaml", func(t *testing.T) {
		mock := &mockFS{data: []byte("{{{")}
		_, err := NewColorMatcher(mock, "test.yaml")
		require.Error(t, err)
	})

	t.Run("file not found", func(t *testing.T) {
		_, err := NewColorMatcher(&mockFS{data: nil}, "test.yaml")
		require.Error(t, err)
	})
}

func TestColorMatcher_Run(t *testing.T) {
	t.Parallel()

	mock := &mockFS{data: []byte(`
colors:
  minecraft:water: [38, 94, 173]
  minecraft:lava: [200, 82, 18]
`)}
	m, err := NewColorMatcher(mock, "test.yaml")
	require.NoError(t, err)

	tests := []struct {
		name    string
		blockID string
		want    model.Color
		wantOk  bool
	}{
		{"water", "minecraft:water", model.Color{38, 94, 173}, true},
		{"lava", "minecraft:lava", model.Color{200, 82, 18}, true},
		{"unlisted block", "minecraft:stone", model.Color{}, false},
		{"custom mod block", "cobblemon:gilded_chest", model.Color{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := m.Run(tt.blockID)
			require.Equal(t, tt.wantOk, ok)
			require.Equal(t, tt.want, got)
		})
	}
}
