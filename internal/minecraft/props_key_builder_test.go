package minecraft

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPropsKeyBuilder_Run(t *testing.T) {
	t.Parallel()

	svc := NewPropsKeyBuilder()

	tests := []struct {
		name  string
		props map[string]string
		want  string
	}{
		{name: "nil map", props: nil, want: ""},
		{name: "empty map", props: map[string]string{}, want: ""},
		{name: "single key", props: map[string]string{"water": "true"}, want: "water=true"},
		{name: "multiple keys sorted", props: map[string]string{"facing": "north", "water": "true"}, want: "facing=north,water=true"},
		{name: "reverse order keys", props: map[string]string{"z": "1", "a": "2", "m": "3"}, want: "a=2,m=3,z=1"},
		{name: "value with special chars", props: map[string]string{"variant": "oak_planks[1]"}, want: "variant=oak_planks[1]"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := svc.Run(tt.props)
			require.Equal(t, tt.want, got)
		})
	}
}
