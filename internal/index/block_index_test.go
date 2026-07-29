package index

import (
	"testing"

	"mc2lua/internal/model"

	"github.com/stretchr/testify/require"
)

func TestNewBlockIndex(t *testing.T) {
	idx := NewBlockIndex()
	require.NotNil(t, idx)
}

func TestBlockIndexAddAndGet(t *testing.T) {
	idx := NewBlockIndex()

	idx.Add("minecraft:stone", "", &model.ResolvedBlock{IsFullBlock: true})
	idx.Add("minecraft:oak_fence", "water=true", &model.ResolvedBlock{IsFullBlock: false})

	tests := []struct {
		name      string
		id        string
		props     string
		wantBlock bool
		wantFull  bool
		wantOk    bool
	}{
		{name: "existing key no props", id: "minecraft:stone", props: "", wantBlock: true, wantFull: true, wantOk: true},
		{name: "existing key with props", id: "minecraft:oak_fence", props: "water=true", wantBlock: true, wantFull: false, wantOk: true},
		{name: "non-existing key", id: "minecraft:air", props: "", wantBlock: false, wantOk: false},
		{name: "non-existing props", id: "minecraft:stone", props: "variant=andesite", wantBlock: false, wantOk: false},
		{name: "wrong ID", id: "stone", props: "", wantBlock: false, wantOk: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := idx.Get(tt.id, tt.props)
			require.Equal(t, tt.wantOk, ok)
			if tt.wantOk {
				require.NotNil(t, got)
				require.Equal(t, tt.wantFull, got.IsFullBlock)
			} else {
				require.Nil(t, got)
			}
		})
	}
}

func TestBlockIndex_Add_NilResolved_DoesNotPanic(t *testing.T) {
	t.Parallel()

	idx := NewBlockIndex()
	require.NotPanics(t, func() {
		idx.Add("minecraft:stone", "", nil)
	})

	v, ok := idx.Get("minecraft:stone", "")
	require.True(t, ok)
	require.Nil(t, v)
}

func TestBlockIndexGetWithColonInProperties(t *testing.T) {
	idx := NewBlockIndex()
	idx.Add("minecraft:stone", "mod:variant", &model.ResolvedBlock{IsFullBlock: true})

	got, ok := idx.Get("minecraft:stone", "mod:variant")
	require.True(t, ok)
	require.NotNil(t, got)
	require.True(t, got.IsFullBlock)
}

func TestBlockIndex_Len(t *testing.T) {
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
			idx := NewBlockIndex()
			for _, id := range tt.ids {
				idx.Add(id, "", &model.ResolvedBlock{})
			}
			require.Equal(t, tt.want, idx.Len())
		})
	}
}
