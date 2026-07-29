package minecraft

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBlockResolver_New(t *testing.T) {
	t.Parallel()

	r := NewBlockResolver()
	require.NotNil(t, r)
}

func TestBlockResolver_Run(t *testing.T) {
	t.Parallel()

	svc := NewBlockResolver()

	tests := []struct {
		name string
		id   string
		props string
	}{
		{name: "empty id and props", id: "", props: ""},
		{name: "non-empty id", id: "minecraft:stone", props: ""},
		{name: "non-empty props", id: "minecraft:oak_log", props: "axis=y"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			resolved, err := svc.Run(tt.id, tt.props)
			require.NoError(t, err)
			require.NotNil(t, resolved)
			require.Empty(t, resolved.Elements)
			require.False(t, resolved.IsFullBlock)
		})
	}
}
