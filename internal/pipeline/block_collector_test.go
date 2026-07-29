package pipeline

import (
	"errors"
	"testing"

	"mc2lua/internal/model"

	"github.com/stretchr/testify/require"
)

type mockCollectorBlockResolver struct {
	runFn func(id string, propsKey string, props map[string]string, namespaces map[string][]string) (*model.ResolvedBlock, error)
}

func (m *mockCollectorBlockResolver) Run(id string, propsKey string, props map[string]string, namespaces map[string][]string) (*model.ResolvedBlock, error) {
	return m.runFn(id, propsKey, props, namespaces)
}

func TestBlockCollector_New(t *testing.T) {
	t.Parallel()

	mr := &mockCollectorBlockResolver{}
	bc := NewBlockCollector(mr, &mockCollectorPropsKeyBuilder{})
	require.NotNil(t, bc)
}

func TestBlockCollector_Run(t *testing.T) {
	t.Parallel()

	resolvedStone := &model.ResolvedBlock{}
	resolvedFence := &model.ResolvedBlock{}
	errResolve := errors.New("resolve failed")

	tests := []struct {
		name       string
		blocks     []model.RawBlock
		resolveFn  func(id string, propsKey string, props map[string]string, namespaces map[string][]string) (*model.ResolvedBlock, error)
		wantUnres  map[string]string
		wantCount  int
		wantCheck  func(t *testing.T, result []model.ResolvedBlock)
	}{
		{
			name:   "empty blocks",
			blocks: []model.RawBlock{},
			resolveFn: func(id string, _ string, props map[string]string, _ map[string][]string) (*model.ResolvedBlock, error) {
				return resolvedStone, nil
			},
			wantCount: 0,
		},
		{
			name: "single block",
			blocks: []model.RawBlock{
				{ID: "minecraft:stone"},
			},
			resolveFn: func(id string, _ string, props map[string]string, _ map[string][]string) (*model.ResolvedBlock, error) {
				return &model.ResolvedBlock{ID: id}, nil
			},
			wantCount: 1,
			wantCheck: func(t *testing.T, result []model.ResolvedBlock) {
				require.Len(t, result, 1)
				require.Equal(t, "minecraft:stone", result[0].ID)
				require.Equal(t, "", result[0].PropsKey)
			},
		},
		{
			name: "multiple different blocks",
			blocks: []model.RawBlock{
				{ID: "minecraft:stone"},
				{ID: "minecraft:oak_fence"},
			},
			resolveFn: func(id string, _ string, props map[string]string, _ map[string][]string) (*model.ResolvedBlock, error) {
				if id == "minecraft:stone" {
					return resolvedStone, nil
				}
				return resolvedFence, nil
			},
			wantCount: 2,
			wantCheck: func(t *testing.T, result []model.ResolvedBlock) {
				require.Len(t, result, 2)
			},
		},
		{
			name: "duplicate blocks call resolver once",
			blocks: []model.RawBlock{
				{ID: "minecraft:stone"},
				{ID: "minecraft:stone"},
				{ID: "minecraft:stone"},
			},
			resolveFn: func(id string, _ string, props map[string]string, _ map[string][]string) (*model.ResolvedBlock, error) {
				return resolvedStone, nil
			},
			wantCount: 1,
		},
		{
			name: "blocks with properties",
			blocks: []model.RawBlock{
				{ID: "minecraft:oak_fence", Props: map[string]string{"water": "true"}},
				{ID: "minecraft:stone", Props: map[string]string{"variant": "andesite"}},
			},
			resolveFn: func(id string, _ string, props map[string]string, _ map[string][]string) (*model.ResolvedBlock, error) {
				propsKey := ""
				if len(props) > 0 {
					keys := make([]string, 0, len(props))
					for k := range props {
						keys = append(keys, k)
					}
					propsKey = keys[0] + "=" + props[keys[0]]
				}
				return &model.ResolvedBlock{ID: id, PropsKey: propsKey}, nil
			},
			wantCount: 2,
			wantCheck: func(t *testing.T, result []model.ResolvedBlock) {
				require.Len(t, result, 2)
				for _, r := range result {
					switch r.ID {
					case "minecraft:oak_fence":
						require.Equal(t, "water=true", r.PropsKey)
					case "minecraft:stone":
						require.Equal(t, "variant=andesite", r.PropsKey)
					}
				}
			},
		},
		{
			name: "resolver error skips block and reports unresolved",
			blocks: []model.RawBlock{
				{ID: "minecraft:stone"},
				{ID: "minecraft:dirt"},
			},
			resolveFn: func(id string, _ string, props map[string]string, _ map[string][]string) (*model.ResolvedBlock, error) {
				if id == "minecraft:stone" {
					return nil, errResolve
				}
				return &model.ResolvedBlock{ID: id}, nil
			},
			wantUnres: map[string]string{"minecraft:stone": errResolve.Error()},
			wantCount: 1,
			wantCheck: func(t *testing.T, result []model.ResolvedBlock) {
				require.Len(t, result, 1)
				require.Equal(t, "minecraft:dirt", result[0].ID)
			},
		},
		{
			name: "duplicate props are distinct keys",
			blocks: []model.RawBlock{
				{ID: "minecraft:stone", Props: map[string]string{"variant": "andesite"}},
				{ID: "minecraft:stone", Props: map[string]string{"variant": "granite"}},
			},
			resolveFn: func(id string, _ string, props map[string]string, _ map[string][]string) (*model.ResolvedBlock, error) {
				propsKey := ""
				if len(props) > 0 {
					keys := make([]string, 0, len(props))
					for k := range props {
						keys = append(keys, k)
					}
					propsKey = keys[0] + "=" + props[keys[0]]
				}
				return &model.ResolvedBlock{ID: id, PropsKey: propsKey}, nil
			},
			wantCount: 2,
			wantCheck: func(t *testing.T, result []model.ResolvedBlock) {
				require.Len(t, result, 2)
				propsSet := make(map[string]bool)
				for _, r := range result {
					propsSet[r.PropsKey] = true
				}
				require.True(t, propsSet["variant=andesite"])
				require.True(t, propsSet["variant=granite"])
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mr := &mockCollectorBlockResolver{runFn: tt.resolveFn}
			svc := NewBlockCollector(mr, &mockCollectorPropsKeyBuilder{})

			result, unresolved := svc.Run(tt.blocks, nil)
			require.Equal(t, tt.wantUnres, unresolved)
			require.Len(t, result, tt.wantCount)
			if tt.wantCheck != nil {
				tt.wantCheck(t, result)
			}
		})
	}
}

type mockCollectorPropsKeyBuilder struct{}

func (m *mockCollectorPropsKeyBuilder) Run(props map[string]string) string {
	if len(props) == 0 {
		return ""
	}
	keys := make([]string, 0, len(props))
	for k := range props {
		keys = append(keys, k)
	}
	return keys[0] + "=" + props[keys[0]]
}
