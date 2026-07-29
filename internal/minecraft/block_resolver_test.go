package minecraft

import (
	"errors"
	"testing"

	"mc2lua/internal/model"

	"github.com/stretchr/testify/require"
)

type mockBlockstateParser struct {
	runFn func(ns, blockID string, props map[string]string, namespaces map[string][]string) ([]blockstateMatch, error)
}

func (m *mockBlockstateParser) Run(ns, blockID string, props map[string]string, namespaces map[string][]string) ([]blockstateMatch, error) {
	return m.runFn(ns, blockID, props, namespaces)
}

type mockModelParser struct {
	runFn func(modelName string, namespaces map[string][]string) (*flattenedModel, error)
}

func (m *mockModelParser) Run(modelName string, namespaces map[string][]string) (*flattenedModel, error) {
	return m.runFn(modelName, namespaces)
}

type mockModelAnalyzer struct {
	runFn func(elements []model.ModelElement) bool
}

func (m *mockModelAnalyzer) Run(elements []model.ModelElement) bool {
	return m.runFn(elements)
}

func TestBlockResolver_SplitBlockID(t *testing.T) {
	t.Parallel()

	svc := &BlockResolver{}
	tests := []struct {
		name      string
		id        string
		wantNS    string
		wantBlock string
	}{
		{name: "with namespace", id: "minecraft:stone", wantNS: "minecraft", wantBlock: "stone"},
		{name: "custom namespace", id: "custom:block", wantNS: "custom", wantBlock: "block"},
		{name: "without namespace defaults to minecraft", id: "stone", wantNS: "minecraft", wantBlock: "stone"},
		{name: "multiple colons", id: "a:b:c", wantNS: "a", wantBlock: "b:c"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ns, block := svc.splitBlockID(tt.id)
			require.Equal(t, tt.wantNS, ns)
			require.Equal(t, tt.wantBlock, block)
		})
	}
}

func TestBlockResolver_New(t *testing.T) {
	t.Parallel()

	r := NewBlockResolver(nil, nil, nil)
	require.NotNil(t, r)
}

func TestBlockResolver_Run_BlockstateParserError(t *testing.T) {
	t.Parallel()

	mockBSP := &mockBlockstateParser{runFn: func(_, _ string, _ map[string]string, _ map[string][]string) ([]blockstateMatch, error) {
		return nil, errors.New("parse error")
	}}

	svc := NewBlockResolver(mockBSP, nil, nil)
	_, err := svc.Run("minecraft:stone", nil, nil)
	require.ErrorContains(t, err, "blockstate")
}

func TestBlockResolver_Run_NoElements(t *testing.T) {
	t.Parallel()

	mockBSP := &mockBlockstateParser{runFn: func(_, _ string, _ map[string]string, _ map[string][]string) ([]blockstateMatch, error) {
		return []blockstateMatch{{Model: "model1"}, {Model: "model2"}}, nil
	}}
	mockMP := &mockModelParser{runFn: func(_ string, _ map[string][]string) (*flattenedModel, error) {
		return nil, errors.New("model error")
	}}

	svc := NewBlockResolver(mockBSP, nil, mockMP)
	_, err := svc.Run("minecraft:stone", nil, nil)
	require.ErrorContains(t, err, "no elements found")
}

func TestBlockResolver_Run_DuplicateModels(t *testing.T) {
	t.Parallel()

	mockBSP := &mockBlockstateParser{runFn: func(_, _ string, _ map[string]string, _ map[string][]string) ([]blockstateMatch, error) {
		return []blockstateMatch{
			{Model: "minecraft:block/stone"},
			{Model: "minecraft:block/stone"},
		}, nil
	}}
	callCount := 0
	mockMP := &mockModelParser{runFn: func(_ string, _ map[string][]string) (*flattenedModel, error) {
		callCount++
		return &flattenedModel{
			Elements: []model.ModelElement{{From: model.Vector3{0, 0, 0}, To: model.Vector3{16, 16, 16}, Shade: true}},
		}, nil
	}}
	mockMA := &mockModelAnalyzer{runFn: func(_ []model.ModelElement) bool {
		return true
	}}

	svc := NewBlockResolver(mockBSP, mockMA, mockMP)
	resolved, err := svc.Run("minecraft:stone", nil, nil)
	require.NoError(t, err)
	require.NotNil(t, resolved)
	require.Equal(t, 1, callCount)
}

func TestBlockResolver_ResolveElements(t *testing.T) {
	t.Parallel()

	elem1 := model.ModelElement{From: model.Vector3{0, 0, 0}, To: model.Vector3{16, 16, 16}, Shade: true}
	elem2 := model.ModelElement{From: model.Vector3{0, 0, 0}, To: model.Vector3{8, 16, 16}, Shade: true}

	tests := []struct {
		name    string
		matches []blockstateMatch
		modelFn func(name string, ns map[string][]string) (*flattenedModel, error)
		want    []model.ModelElement
	}{
		{
			name:    "empty matches",
			matches: nil,
			want:    nil,
		},
		{
			name:    "empty matches slice",
			matches: []blockstateMatch{},
			want:    nil,
		},
		{
			name:    "single match resolved",
			matches: []blockstateMatch{{Model: "minecraft:block/stone"}},
			modelFn: func(name string, ns map[string][]string) (*flattenedModel, error) {
				return &flattenedModel{Elements: []model.ModelElement{elem1}}, nil
			},
			want: []model.ModelElement{elem1},
		},
		{
			name: "duplicate models parsed once",
			matches: []blockstateMatch{
				{Model: "minecraft:block/stone"},
				{Model: "minecraft:block/stone"},
			},
			modelFn: func(name string, ns map[string][]string) (*flattenedModel, error) {
				return &flattenedModel{Elements: []model.ModelElement{elem1}}, nil
			},
			want: []model.ModelElement{elem1},
		},
		{
			name:    "model parse error skipped",
			matches: []blockstateMatch{{Model: "minecraft:block/broken"}},
			modelFn: func(name string, ns map[string][]string) (*flattenedModel, error) {
				return nil, errors.New("parse error")
			},
			want: nil,
		},
		{
			name: "multiple different models",
			matches: []blockstateMatch{
				{Model: "minecraft:block/stone"},
				{Model: "minecraft:block/slab"},
			},
			modelFn: func(name string, ns map[string][]string) (*flattenedModel, error) {
				switch name {
				case "minecraft:block/stone":
					return &flattenedModel{Elements: []model.ModelElement{elem1}}, nil
				default:
					return &flattenedModel{Elements: []model.ModelElement{elem2}}, nil
				}
			},
			want: []model.ModelElement{elem1, elem2},
		},
		{
			name: "mixed success and error models",
			matches: []blockstateMatch{
				{Model: "minecraft:block/stone"},
				{Model: "minecraft:block/broken"},
			},
			modelFn: func(name string, ns map[string][]string) (*flattenedModel, error) {
				if name == "minecraft:block/broken" {
					return nil, errors.New("parse error")
				}
				return &flattenedModel{Elements: []model.ModelElement{elem1}}, nil
			},
			want: []model.ModelElement{elem1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockMP := &mockModelParser{runFn: tt.modelFn}
			svc := &BlockResolver{modelParser: mockMP}

			got := svc.resolveElements(tt.matches, nil)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestBlockResolver_Run(t *testing.T) {
	t.Parallel()

	mockBSP := &mockBlockstateParser{runFn: func(ns, blockID string, props map[string]string, namespaces map[string][]string) ([]blockstateMatch, error) {
		return []blockstateMatch{{Model: "minecraft:block/stone"}}, nil
	}}
	mockMP := &mockModelParser{runFn: func(modelName string, namespaces map[string][]string) (*flattenedModel, error) {
		return &flattenedModel{
			Elements: []model.ModelElement{
				{From: model.Vector3{0, 0, 0}, To: model.Vector3{16, 16, 16}},
			},
		}, nil
	}}
	mockMA := &mockModelAnalyzer{runFn: func(elements []model.ModelElement) bool {
		return false
	}}

	svc := NewBlockResolver(mockBSP, mockMA, mockMP)

	tests := []struct {
		name  string
		id    string
		props map[string]string
	}{
		{name: "nil props", id: "", props: nil},
		{name: "non-empty id", id: "minecraft:stone", props: nil},
		{name: "non-empty props", id: "minecraft:oak_log", props: map[string]string{"axis": "y"}},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			resolved, err := svc.Run(tt.id, tt.props, nil)
			require.NoError(t, err)
			require.NotNil(t, resolved)
			require.NotEmpty(t, resolved.Elements)
			require.False(t, resolved.IsFullBlock)
		})
	}
}
