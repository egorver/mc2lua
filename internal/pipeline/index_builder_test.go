package pipeline

import (
	"errors"
	"testing"

	"mc2lua/internal/index"
	"mc2lua/internal/model"

	"github.com/stretchr/testify/require"
)

type mockBlockResolver struct {
	runFn func(id, props string, namespaces map[string][]string) (*model.ResolvedBlock, error)
}

func (m *mockBlockResolver) Run(id, props string, namespaces map[string][]string) (*model.ResolvedBlock, error) {
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

	resolvedStone := &model.ResolvedBlock{IsFullBlock: true}
	resolvedFence := &model.ResolvedBlock{IsFullBlock: false}
	errResolve := errors.New("resolve failed")

	tests := []struct {
		name       string
		blocks     []model.Block
		resolveFn  func(id, props string, namespaces map[string][]string) (*model.ResolvedBlock, error)
		wantErr    bool
		wantErrMsg string
		wantCheck  func(t *testing.T, idx *index.BlockIndex)
	}{
		{
			name:   "empty blocks",
			blocks: []model.Block{},
			resolveFn: func(id, props string, _ map[string][]string) (*model.ResolvedBlock, error) {
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
			resolveFn: func(id, props string, _ map[string][]string) (*model.ResolvedBlock, error) {
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
			resolveFn: func(id, props string, _ map[string][]string) (*model.ResolvedBlock, error) {
				if id == "minecraft:stone" {
					return resolvedStone, nil
				}
				return resolvedFence, nil
			},
			wantCheck: func(t *testing.T, idx *index.BlockIndex) {
				v1, ok := idx.Get("minecraft:stone", "")
				require.True(t, ok)
				require.True(t, v1.IsFullBlock)
				v2, ok := idx.Get("minecraft:oak_fence", "")
				require.True(t, ok)
				require.False(t, v2.IsFullBlock)
			},
		},
		{
			name: "duplicate blocks call resolver once",
			blocks: []model.Block{
				{ID: "minecraft:stone"},
				{ID: "minecraft:stone"},
				{ID: "minecraft:stone"},
			},
			resolveFn: func(id, props string, _ map[string][]string) (*model.ResolvedBlock, error) {
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
			resolveFn: func(id, props string, _ map[string][]string) (*model.ResolvedBlock, error) {
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
			name: "resolver error",
			blocks: []model.Block{
				{ID: "minecraft:stone"},
			},
			resolveFn: func(id, props string, _ map[string][]string) (*model.ResolvedBlock, error) {
				return nil, errResolve
			},
			wantErr:    true,
			wantErrMsg: "resolve block minecraft:stone|",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mr := &mockBlockResolver{runFn: tt.resolveFn}
			svc := NewIndexBuilder(mr)

			idx, err := svc.Run(tt.blocks, nil)
			if tt.wantErr {
				require.Error(t, err)
				if tt.wantErrMsg != "" {
					require.Contains(t, err.Error(), tt.wantErrMsg)
				}
				return
			}
			require.NoError(t, err)
			require.NotNil(t, idx)
			if tt.wantCheck != nil {
				tt.wantCheck(t, idx)
			}
		})
	}
}

func TestIndexBuilder_Run_ResolverCallCount(t *testing.T) {
	t.Parallel()

	callCount := 0
	mr := &mockBlockResolver{
		runFn: func(id, props string, _ map[string][]string) (*model.ResolvedBlock, error) {
			callCount++
			return &model.ResolvedBlock{IsFullBlock: true}, nil
		},
	}
	svc := NewIndexBuilder(mr)

	blocks := []model.Block{
		{ID: "minecraft:stone"},
		{ID: "minecraft:stone"},
		{ID: "minecraft:stone"},
		{ID: "minecraft:dirt"},
	}
	_, err := svc.Run(blocks, nil)
	require.NoError(t, err)
	require.Equal(t, 2, callCount, "resolver should be called once per unique id|props")
}

func TestIndexBuilder_Run_DuplicatePropsAreDistinct(t *testing.T) {
	t.Parallel()

	mr := &mockBlockResolver{
		runFn: func(id, props string, _ map[string][]string) (*model.ResolvedBlock, error) {
			return &model.ResolvedBlock{IsFullBlock: id == "minecraft:stone"}, nil
		},
	}
	svc := NewIndexBuilder(mr)

	blocks := []model.Block{
		{ID: "minecraft:stone", Props: map[string]string{"variant": "andesite"}},
		{ID: "minecraft:stone", Props: map[string]string{"variant": "granite"}},
	}
	idx, err := svc.Run(blocks, nil)
	require.NoError(t, err)

	v, ok := idx.Get("minecraft:stone", "variant=andesite")
	require.True(t, ok)
	require.True(t, v.IsFullBlock)

	v, ok = idx.Get("minecraft:stone", "variant=granite")
	require.True(t, ok)
	require.True(t, v.IsFullBlock)
}

func TestIndexBuilder_SerializeProps(t *testing.T) {
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
			got := svc.serializeProps(tt.props)
			require.Equal(t, tt.want, got)
		})
	}
}
