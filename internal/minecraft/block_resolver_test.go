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

type mockTextureResolver struct {
	runFn func(textures map[string]string) map[string]string
}

func (m *mockTextureResolver) Run(textures map[string]string) map[string]string {
	return m.runFn(textures)
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

	r := NewBlockResolver(nil, nil, nil, NewElementRotator())
	require.NotNil(t, r)
}

func TestBlockResolver_Run(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		id         string
		propsKey   string
		props      map[string]string
		bspFn      func(_, _ string, _ map[string]string, _ map[string][]string) ([]blockstateMatch, error)
		mpFn       func(_ string, _ map[string][]string) (*flattenedModel, error)
		wantErr    string
		wantModels int
	}{
		{
			name: "blockstate parser error",
			id:   "minecraft:stone",
			bspFn: func(_, _ string, _ map[string]string, _ map[string][]string) ([]blockstateMatch, error) {
				return nil, errors.New("parse error")
			},
			wantErr: "blockstate",
		},
		{
			name: "no elements found",
			id:   "minecraft:stone",
			bspFn: func(_, _ string, _ map[string]string, _ map[string][]string) ([]blockstateMatch, error) {
				return []blockstateMatch{{Model: "model1"}, {Model: "model2"}}, nil
			},
			mpFn: func(_ string, _ map[string][]string) (*flattenedModel, error) {
				return nil, errors.New("model error")
			},
			wantErr: "no elements found",
		},
		{
			name: "duplicate models resolved once",
			id:   "minecraft:stone",
			bspFn: func(_, _ string, _ map[string]string, _ map[string][]string) ([]blockstateMatch, error) {
				return []blockstateMatch{{Model: "minecraft:block/stone"}, {Model: "minecraft:block/stone"}}, nil
			},
			mpFn: func(_ string, _ map[string][]string) (*flattenedModel, error) {
				return &flattenedModel{Elements: []model.ModelElement{{From: model.Vector3{0, 0, 0}, To: model.Vector3{16, 16, 16}, Shade: true}}}, nil
			},
			wantModels: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			callCount := 0
			mpFn := tt.mpFn
			if mpFn == nil {
				mpFn = func(_ string, _ map[string][]string) (*flattenedModel, error) {
					return &flattenedModel{Elements: []model.ModelElement{{}}}, nil
				}
			}
			mockBSP := &mockBlockstateParser{runFn: tt.bspFn}
			mockMP := &mockModelParser{runFn: func(name string, ns map[string][]string) (*flattenedModel, error) {
				callCount++
				return mpFn(name, ns)
			}}
			mockTR := &mockTextureResolver{
				runFn: func(textures map[string]string) map[string]string {
					return textures
				},
			}

			svc := NewBlockResolver(mockBSP, mockMP, mockTR, NewElementRotator())
			resolved, err := svc.Run(tt.id, tt.propsKey, tt.props, nil)
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resolved)
			if tt.wantModels > 0 {
				require.Equal(t, tt.wantModels, callCount)
			}
		})
	}
}

func TestBlockResolver_Run_WithRotation(t *testing.T) {
	t.Parallel()

	t.Run("rotation from match sets RotY", func(t *testing.T) {
		mockBSP := &mockBlockstateParser{
			runFn: func(_, _ string, _ map[string]string, _ map[string][]string) ([]blockstateMatch, error) {
				return []blockstateMatch{{Model: "minecraft:block/furnace", RotY: 180}}, nil
			},
		}
		mockMP := &mockModelParser{
			runFn: func(_ string, _ map[string][]string) (*flattenedModel, error) {
				return &flattenedModel{Elements: []model.ModelElement{{Shade: true}}}, nil
			},
		}
		mockTR := &mockTextureResolver{
			runFn: func(textures map[string]string) map[string]string { return textures },
		}
		svc := NewBlockResolver(mockBSP, mockMP, mockTR, NewElementRotator())
		resolved, err := svc.Run("minecraft:furnace", "", nil, nil)
		require.NoError(t, err)
		require.Equal(t, 0.0, resolved.RotX)
		require.Equal(t, 180.0, resolved.RotY)
	})

	t.Run("same rotation across matches kept as global rotation", func(t *testing.T) {
		mockBSP := &mockBlockstateParser{
			runFn: func(_, _ string, _ map[string]string, _ map[string][]string) ([]blockstateMatch, error) {
				return []blockstateMatch{
					{Model: "minecraft:block/stone", RotY: 90},
					{Model: "minecraft:block/stone_alt", RotY: 90},
				}, nil
			},
		}
		mockMP := &mockModelParser{
			runFn: func(name string, _ map[string][]string) (*flattenedModel, error) {
				return &flattenedModel{Elements: []model.ModelElement{{Shade: true}}}, nil
			},
		}
		mockTR := &mockTextureResolver{
			runFn: func(textures map[string]string) map[string]string { return textures },
		}
		svc := NewBlockResolver(mockBSP, mockMP, mockTR, NewElementRotator())
		resolved, err := svc.Run("minecraft:stone", "", nil, nil)
		require.NoError(t, err)
		require.Equal(t, 0.0, resolved.RotX)
		require.Equal(t, 90.0, resolved.RotY)
	})

	t.Run("differing rotations are baked and global rotation zeroed", func(t *testing.T) {
		sideElem := model.ModelElement{From: model.Vector3{0, 0, 0}, To: model.Vector3{8, 8, 8}, Shade: true}
		mockBSP := &mockBlockstateParser{
			runFn: func(_, _ string, _ map[string]string, _ map[string][]string) ([]blockstateMatch, error) {
				return []blockstateMatch{
					{Model: "minecraft:block/side"},
					{Model: "minecraft:block/side", RotY: 90},
				}, nil
			},
		}
		mockMP := &mockModelParser{
			runFn: func(_ string, _ map[string][]string) (*flattenedModel, error) {
				return &flattenedModel{Elements: []model.ModelElement{sideElem}}, nil
			},
		}
		mockTR := &mockTextureResolver{
			runFn: func(textures map[string]string) map[string]string { return textures },
		}
		svc := NewBlockResolver(mockBSP, mockMP, mockTR, NewElementRotator())
		resolved, err := svc.Run("minecraft:oak_fence", "", nil, nil)
		require.NoError(t, err)
		require.Equal(t, 0.0, resolved.RotX)
		require.Equal(t, 0.0, resolved.RotY)
		require.Len(t, resolved.Elements, 2)
		require.Equal(t, model.Vector3{0, 0, 0}, resolved.Elements[0].From)
		require.Equal(t, model.Vector3{8, 8, 8}, resolved.Elements[0].To)
		require.Equal(t, model.Vector3{8, 0, 0}, resolved.Elements[1].From)
		require.Equal(t, model.Vector3{16, 8, 8}, resolved.Elements[1].To)
	})

	t.Run("nil props does not panic", func(t *testing.T) {
		mockBSP := &mockBlockstateParser{
			runFn: func(_, _ string, _ map[string]string, _ map[string][]string) ([]blockstateMatch, error) {
				return []blockstateMatch{{Model: "minecraft:block/stone"}}, nil
			},
		}
		mockMP := &mockModelParser{
			runFn: func(_ string, _ map[string][]string) (*flattenedModel, error) {
				return &flattenedModel{Elements: []model.ModelElement{{Shade: true}}}, nil
			},
		}
		mockTR := &mockTextureResolver{
			runFn: func(textures map[string]string) map[string]string { return textures },
		}
		svc := NewBlockResolver(mockBSP, mockMP, mockTR, NewElementRotator())
		resolved, err := svc.Run("minecraft:stone", "key", nil, nil)
		require.NoError(t, err)
		require.NotNil(t, resolved)
	})

	t.Run("textures passed through to result", func(t *testing.T) {
		mockBSP := &mockBlockstateParser{
			runFn: func(_, _ string, _ map[string]string, _ map[string][]string) ([]blockstateMatch, error) {
				return []blockstateMatch{{Model: "minecraft:block/cube"}}, nil
			},
		}
		mockMP := &mockModelParser{
			runFn: func(_ string, _ map[string][]string) (*flattenedModel, error) {
				return &flattenedModel{
					Elements: []model.ModelElement{{Shade: true}},
					Textures: map[string]string{"particle": "block/particle"},
				}, nil
			},
		}
		mockTR := &mockTextureResolver{
			runFn: func(textures map[string]string) map[string]string {
				result := make(map[string]string, len(textures))
				for k, v := range textures {
					result[k] = "resolved:" + v
				}
				return result
			},
		}
		svc := NewBlockResolver(mockBSP, mockMP, mockTR, NewElementRotator())
		resolved, err := svc.Run("minecraft:cube", "", nil, nil)
		require.NoError(t, err)
		require.Equal(t, "resolved:block/particle", resolved.Textures["particle"])
	})
}

func TestBlockResolver_ResolveElements(t *testing.T) {
	t.Parallel()

	elem1 := model.ModelElement{From: model.Vector3{0, 0, 0}, To: model.Vector3{16, 16, 16}, Shade: true}
	elem2 := model.ModelElement{From: model.Vector3{0, 0, 0}, To: model.Vector3{8, 16, 16}, Shade: true}

	tests := []struct {
		name    string
		matches []blockstateMatch
		modelFn func(name string, ns map[string][]string) (*flattenedModel, error)
		want    *flattenedModel
	}{
		{
			name:    "empty matches",
			matches: nil,
			modelFn: nil,
			want:    &flattenedModel{Textures: map[string]string{}},
		},
		{
			name:    "empty matches slice",
			matches: []blockstateMatch{},
			modelFn: nil,
			want:    &flattenedModel{Textures: map[string]string{}},
		},
		{
			name:    "single match resolved",
			matches: []blockstateMatch{{Model: "minecraft:block/stone"}},
			modelFn: func(name string, ns map[string][]string) (*flattenedModel, error) {
				return &flattenedModel{Elements: []model.ModelElement{elem1}}, nil
			},
			want: &flattenedModel{Elements: []model.ModelElement{elem1}, Textures: map[string]string{}},
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
			want: &flattenedModel{Elements: []model.ModelElement{elem1}, Textures: map[string]string{}},
		},
		{
			name:    "model parse error skipped",
			matches: []blockstateMatch{{Model: "minecraft:block/broken"}},
			modelFn: func(name string, ns map[string][]string) (*flattenedModel, error) {
				return nil, errors.New("parse error")
			},
			want: &flattenedModel{Textures: map[string]string{}},
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
			want: &flattenedModel{Elements: []model.ModelElement{elem1, elem2}, Textures: map[string]string{}},
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
			want: &flattenedModel{Elements: []model.ModelElement{elem1}, Textures: map[string]string{}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockMP := &mockModelParser{runFn: tt.modelFn}
			svc := &BlockResolver{modelParser: mockMP}

			got, rotX, rotY := svc.resolveFlattened(tt.matches, nil)
			require.Equal(t, tt.want, got)
			require.Equal(t, 0.0, rotX)
			require.Equal(t, 0.0, rotY)
		})
	}
}

func TestBlockResolver_RotationStrategy(t *testing.T) {
	t.Parallel()

	svc := &BlockResolver{}
	tests := []struct {
		name     string
		matches  []blockstateMatch
		wantBake bool
		wantRotX float64
		wantRotY float64
	}{
		{
			name:     "empty matches",
			matches:  nil,
			wantBake: false,
		},
		{
			name:     "no rotation",
			matches:  []blockstateMatch{{Model: "block/a"}, {Model: "block/b"}},
			wantBake: false,
		},
		{
			name:     "single rotated match",
			matches:  []blockstateMatch{{Model: "block/furnace", RotY: 180}},
			wantBake: false,
			wantRotY: 180,
		},
		{
			name: "uniform rotation kept as global",
			matches: []blockstateMatch{
				{Model: "block/a", RotY: 90},
				{Model: "block/b", RotY: 90},
			},
			wantBake: false,
			wantRotY: 90,
		},
		{
			name: "differing rotations baked and global zeroed",
			matches: []blockstateMatch{
				{Model: "block/a"},
				{Model: "block/b", RotY: 90},
			},
			wantBake: true,
		},
		{
			name: "rotX difference triggers bake",
			matches: []blockstateMatch{
				{Model: "block/a", RotX: 90},
				{Model: "block/b", RotX: 180},
			},
			wantBake: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			bake, rotX, rotY := svc.rotationStrategy(tt.matches)
			require.Equal(t, tt.wantBake, bake)
			require.Equal(t, tt.wantRotX, rotX)
			require.Equal(t, tt.wantRotY, rotY)
		})
	}
}
