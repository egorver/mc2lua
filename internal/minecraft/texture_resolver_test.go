package minecraft

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTextureResolver_Run(t *testing.T) {
	t.Parallel()

	svc := NewTextureResolver()
	tests := []struct {
		name string
		in   map[string]string
		want map[string]string
	}{
		{
			name: "no refs",
			in:   map[string]string{"a": "block/stone"},
			want: map[string]string{"a": "block/stone"},
		},
		{
			name: "direct ref",
			in:   map[string]string{"a": "#b", "b": "block/stone"},
			want: map[string]string{"a": "block/stone", "b": "block/stone"},
		},
		{
			name: "chain",
			in:   map[string]string{"a": "#b", "b": "#c", "c": "block/stone"},
			want: map[string]string{"a": "block/stone", "b": "block/stone", "c": "block/stone"},
		},
		{
			name: "self-cycle",
			in:   map[string]string{"a": "#a"},
			want: map[string]string{"a": "#a"},
		},
		{
			name: "cross-cycle",
			in:   map[string]string{"a": "#b", "b": "#a"},
			want: map[string]string{"a": "#b", "b": "#a"},
		},
		{
			name: "broken ref",
			in:   map[string]string{"a": "#nonexistent"},
			want: map[string]string{"a": "#nonexistent"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := svc.Run(tt.in)
			require.Equal(t, tt.want, got)
		})
	}
}
