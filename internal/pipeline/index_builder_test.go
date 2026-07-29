package pipeline

import (
	"errors"
	"testing"

	"mc2lua/internal/index"
	"mc2lua/internal/model"

	"github.com/stretchr/testify/require"
)

type mockBlockResolver struct {
	runFn func(id string, props map[string]string, namespaces map[string][]string) (*model.ResolvedBlock, error)
}

func (m *mockBlockResolver) Run(id string, props map[string]string, namespaces map[string][]string) (*model.ResolvedBlock, error) {
	return m.runFn(id, props, namespaces)
}

func TestIndexBuilder_New(t *testing.T) {
	t.Parallel()

	mr := &mockBlockResolver{}
	ib := NewIndexBuilder(mr)
	require.NotNil(t, ib)
}

func TestIndexBuilder_Run(t *testing.T) {
	t.Parallel()

	resolvedStone := &model.ResolvedBlock{}
	resolvedFence := &model.ResolvedBlock{}
	errResolve := errors.New("resolve failed")

	tests := []struct {
		name             string
		blocks           []model.Block
		resolveFn        func(id string, props map[string]string, namespaces map[string][]string) (*model.ResolvedBlock, error)
		wantErr          bool
		wantErrMsg       string
		wantUnresolved   map[string]string
		resolveCallCount int
		wantCheck        func(t *testing.T, idx *index.BlockIndex)
	}{
		{
			name:   "empty blocks",
			blocks: []model.Block{},
			resolveFn: func(id string, props map[string]string, _ map[string][]string) (*model.ResolvedBlock, error) {
				return resolvedStone, nil
			},
			wantCheck: func(t *testing.T, idx *index.BlockIndex) {
				_, ok := idx.Get("minecraft:stone", "")
				require.False(t, ok)
			},
		},
		{
			name: "single block",
			blocks: []model.Block{
				{ID: "minecraft:stone"},
			},
			resolveFn: func(id string, props map[string]string, _ map[string][]string) (*model.ResolvedBlock, error) {
				return resolvedStone, nil
			},
			wantCheck: func(t *testing.T, idx *index.BlockIndex) {
				v, ok := idx.Get("minecraft:stone", "")
				require.True(t, ok)
				require.Equal(t, resolvedStone, v)
			},
		},
		{
			name: "multiple different blocks",
			blocks: []model.Block{
				{ID: "minecraft:stone"},
				{ID: "minecraft:oak_fence"},
			},
			resolveFn: func(id string, props map[string]string, _ map[string][]string) (*model.ResolvedBlock, error) {
				if id == "minecraft:stone" {
					return resolvedStone, nil
				}
				return resolvedFence, nil
			},
			wantCheck: func(t *testing.T, idx *index.BlockIndex) {
				_, ok := idx.Get("minecraft:stone", "")
				require.True(t, ok)
				_, ok = idx.Get("minecraft:oak_fence", "")
				require.True(t, ok)
			},
		},
		{
			name: "duplicate blocks call resolver once",
			blocks: []model.Block{
				{ID: "minecraft:stone"},
				{ID: "minecraft:stone"},
				{ID: "minecraft:stone"},
			},
			resolveFn: func(id string, props map[string]string, _ map[string][]string) (*model.ResolvedBlock, error) {
				return resolvedStone, nil
			},
			wantCheck: func(t *testing.T, idx *index.BlockIndex) {
				_, ok := idx.Get("minecraft:stone", "")
				require.True(t, ok)
				_, ok = idx.Get("minecraft:dirt", "")
				require.False(t, ok)
			},
		},
		{
			name: "blocks with properties",
			blocks: []model.Block{
				{ID: "minecraft:oak_fence", Props: map[string]string{"water": "true"}},
				{ID: "minecraft:stone", Props: map[string]string{"variant": "andesite"}},
			},
			resolveFn: func(id string, props map[string]string, _ map[string][]string) (*model.ResolvedBlock, error) {
				return resolvedFence, nil
			},
			wantCheck: func(t *testing.T, idx *index.BlockIndex) {
				v, ok := idx.Get("minecraft:oak_fence", "water=true")
				require.True(t, ok)
				require.Equal(t, resolvedFence, v)
				v, ok = idx.Get("minecraft:stone", "variant=andesite")
				require.True(t, ok)
				require.Equal(t, resolvedFence, v)
			},
		},
		{
			name: "resolver error skips block and reports unresolved",
			blocks: []model.Block{
				{ID: "minecraft:stone"},
				{ID: "minecraft:dirt"},
			},
			resolveFn: func(id string, props map[string]string, _ map[string][]string) (*model.ResolvedBlock, error) {
				if id == "minecraft:stone" {
					return nil, errResolve
				}
				return resolvedStone, nil
			},
			wantUnresolved: map[string]string{"minecraft:stone": errResolve.Error()},
			wantCheck: func(t *testing.T, idx *index.BlockIndex) {
				_, ok := idx.Get("minecraft:stone", "")
				require.False(t, ok)
				v, ok := idx.Get("minecraft:dirt", "")
				require.True(t, ok)
				require.Equal(t, resolvedStone, v)
			},
		},
		{
			name: "resolver called once per unique id|props",
			blocks: []model.Block{
				{ID: "minecraft:stone"},
				{ID: "minecraft:stone"},
				{ID: "minecraft:stone"},
				{ID: "minecraft:dirt"},
			},
			resolveFn: func(id string, props map[string]string, _ map[string][]string) (*model.ResolvedBlock, error) {
				return &model.ResolvedBlock{}, nil
			},
			resolveCallCount: 2,
			wantCheck: func(t *testing.T, idx *index.BlockIndex) {
				_, ok := idx.Get("minecraft:stone", "")
				require.True(t, ok)
			},
		},
		{
			name: "duplicate props are distinct keys",
			blocks: []model.Block{
				{ID: "minecraft:stone", Props: map[string]string{"variant": "andesite"}},
				{ID: "minecraft:stone", Props: map[string]string{"variant": "granite"}},
			},
			resolveFn: func(id string, props map[string]string, _ map[string][]string) (*model.ResolvedBlock, error) {
				return &model.ResolvedBlock{}, nil
			},
			wantCheck: func(t *testing.T, idx *index.BlockIndex) {
				_, ok := idx.Get("minecraft:stone", "variant=andesite")
				require.True(t, ok)
				_, ok = idx.Get("minecraft:stone", "variant=granite")
				require.True(t, ok)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			callCount := 0
			resolveFn := tt.resolveFn
			if resolveFn != nil {
				wrappedFn := resolveFn
				resolveFn = func(id string, props map[string]string, ns map[string][]string) (*model.ResolvedBlock, error) {
					callCount++
					return wrappedFn(id, props, ns)
				}
			}
			mr := &mockBlockResolver{runFn: resolveFn}
			svc := NewIndexBuilder(mr)

			idx, unresolved, err := svc.Run(tt.blocks, nil)
			if tt.wantErr {
				require.Error(t, err)
				if tt.wantErrMsg != "" {
					require.Contains(t, err.Error(), tt.wantErrMsg)
				}
				return
			}
			require.NoError(t, err)
			require.NotNil(t, idx)
			require.Equal(t, tt.wantUnresolved, unresolved)
			if tt.resolveCallCount > 0 {
				require.Equal(t, tt.resolveCallCount, callCount)
			}
			if tt.wantCheck != nil {
				tt.wantCheck(t, idx)
			}
		})
	}
}

func TestIndexBuilder_PropsToKey(t *testing.T) {
	t.Parallel()

	svc := &IndexBuilder{}

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
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := svc.propsToKey(tt.props)
			require.Equal(t, tt.want, got)
		})
	}
}
