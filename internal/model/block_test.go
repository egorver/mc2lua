package model

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBlock_Properties(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		block Block
	}{
		{
			name: "nil props",
			block: Block{
				ID: "minecraft:stone",
				X:  10, Y: 20, Z: 30,
			},
		},
		{
			name: "empty props map",
			block: Block{
				ID:    "minecraft:stone",
				Props: map[string]string{},
				X:     10, Y: 20, Z: 30,
			},
		},
		{
			name: "with props",
			block: Block{
				ID:    "minecraft:stairs",
				Props: map[string]string{"facing": "north"},
				X:     10, Y: 20, Z: 30,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.block.X, 10)
			require.Equal(t, tt.block.Y, 20)
			require.Equal(t, tt.block.Z, 30)
		})
	}
}
