package stateful

import (
	"mc2lua/internal/model"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewStyleIndex(t *testing.T) {
	idx := NewStyleIndex()
	require.NotNil(t, idx)
}

func TestStyleIndexAddAndGet(t *testing.T) {
	idx := NewStyleIndex()

	idx.Add("minecraft:stone", "", model.StyledBlock{})
	idx.Add("minecraft:oak_fence", "water=true", model.StyledBlock{GridAlignment: model.GridNotAligned})

	tests := []struct {
		name   string
		id     string
		props  string
		want   model.StyledBlock
		wantOk bool
	}{
		{name: "existing key no props", id: "minecraft:stone", props: "", want: model.StyledBlock{}, wantOk: true},
		{name: "existing key with props", id: "minecraft:oak_fence", props: "water=true", want: model.StyledBlock{GridAlignment: model.GridNotAligned}, wantOk: true},
		{name: "non-existing key", id: "minecraft:air", props: "", wantOk: false},
		{name: "non-existing props", id: "minecraft:stone", props: "variant=andesite", wantOk: false},
		{name: "wrong ID", id: "stone", props: "", wantOk: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := idx.Get(tt.id, tt.props)
			require.Equal(t, tt.wantOk, ok)
			if tt.wantOk {
				require.Equal(t, tt.want, got)
			}
		})
	}
}

func TestStyleIndex_Add_ZeroValue(t *testing.T) {
	t.Parallel()

	idx := NewStyleIndex()
	idx.Add("minecraft:stone", "", model.StyledBlock{})

	v, ok := idx.Get("minecraft:stone", "")
	require.True(t, ok)
	require.Equal(t, model.StyledBlock{}, v)
}

func TestStyleIndexGetWithColonInProperties(t *testing.T) {
	idx := NewStyleIndex()
	idx.Add("minecraft:stone", "mod:variant", model.StyledBlock{})

	got, ok := idx.Get("minecraft:stone", "mod:variant")
	require.True(t, ok)
	require.Equal(t, model.StyledBlock{}, got)
}

func TestStyleIndex_Len(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		ids  []string
		want int
	}{
		{name: "empty index", ids: nil, want: 0},
		{name: "one entry", ids: []string{"stone"}, want: 1},
		{name: "multiple entries", ids: []string{"stone", "dirt", "grass"}, want: 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			idx := NewStyleIndex()
			for _, id := range tt.ids {
				idx.Add(id, "", model.StyledBlock{})
			}
			require.Equal(t, tt.want, idx.Len())
		})
	}
}
