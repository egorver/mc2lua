package minecraft

import (
	"encoding/json"
	"testing"

	"mc2lua/internal/model"
	"mc2lua/internal/runtime"

	"github.com/stretchr/testify/require"
)

func TestModelParserRun(t *testing.T) {
	fs, nsToRoots := setupTestFS()
	addModel(fs, "minecraft", "block/cube", []byte(testModelCube))
	addModel(fs, "minecraft", "block/grass", []byte(testModelGrass))
	addModel(fs, "minecraft", "block/grass_block", []byte(testModelGrandchild))
	addModel(fs, "minecraft", "block/leaves", []byte(testModelNoElements))
	addModel(fs, "minecraft", "block/no_textures", []byte(testModelNoTextures))
	addModel(fs, "minecraft", "block/bad_json", []byte(testModelInvalidJSON))
	addModel(fs, "minecraft", "block/implicit_shade", []byte(testModelImplicitShade))

	svc := NewModelParser(fs)

	tests := []struct {
		name      string
		modelName string
		want      *flattenedModel
		wantErr   string
	}{
		{
			name:      "simple model no parent",
			modelName: "minecraft:block/cube",
			want: &flattenedModel{
				Elements: []model.ModelElement{{From: model.Vector3{0, 0, 0}, To: model.Vector3{16, 16, 16}, Shade: true}},
				Textures: map[string]string{"particle": "block/particle"},
			},
		},
		{
			name:      "model with parent",
			modelName: "minecraft:block/grass",
			want: &flattenedModel{
				Elements: []model.ModelElement{{From: model.Vector3{0, 0, 0}, To: model.Vector3{16, 16, 16}, Shade: true}},
				Textures: map[string]string{"particle": "block/particle", "all": "block/grass_side"},
			},
		},
		{
			name:      "grandchild model",
			modelName: "minecraft:block/grass_block",
			want: &flattenedModel{
				Elements: []model.ModelElement{{From: model.Vector3{0, 0, 0}, To: model.Vector3{16, 16, 16}, Shade: true}},
				Textures: map[string]string{"particle": "block/particle", "all": "block/grass_side", "grass": "block/grass_top"},
			},
		},
		{
			name:      "no elements model",
			modelName: "minecraft:block/leaves",
			want: &flattenedModel{
				Elements: nil,
				Textures: map[string]string{"particle": "block/stone"},
			},
		},
		{
			name:      "implicit shade defaults to true",
			modelName: "minecraft:block/implicit_shade",
			want: &flattenedModel{
				Elements: []model.ModelElement{{From: model.Vector3{0, 0, 0}, To: model.Vector3{16, 16, 16}, Shade: true}},
				Textures: map[string]string{},
			},
		},
		{
			name:      "default namespace implicit",
			modelName: "block/cube",
			want: &flattenedModel{
				Elements: []model.ModelElement{{From: model.Vector3{0, 0, 0}, To: model.Vector3{16, 16, 16}, Shade: true}},
				Textures: map[string]string{"particle": "block/particle"},
			},
		},
		{
			name:      "unknown namespace",
			modelName: "unknown:block/stone",
			wantErr:   "unknown namespace",
		},
		{
			name:      "file not found",
			modelName: "minecraft:block/nonexistent",
			wantErr:   "read",
		},
		{
			name:      "invalid JSON",
			modelName: "minecraft:block/bad_json",
			wantErr:   "parse",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := svc.Run(tt.modelName, nsToRoots)
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestModelParserRunParentError(t *testing.T) {
	fs, nsToRoots := setupTestFS()
	addModel(fs, "minecraft", "block/has_parent", []byte(`{"parent":"block/nonexistent"}`))
	svc := NewModelParser(fs)

	_, err := svc.Run("minecraft:block/has_parent", nsToRoots)
	require.ErrorContains(t, err, "parent")
}

func TestModelParserRunMultipleRoots(t *testing.T) {
	fs := runtime.NewFSMock()
	fs.AddDir("assets", 0755)
	fs.AddDir("assets/mod1", 0755)
	fs.AddDir("assets/mod1/minecraft", 0755)
	fs.AddDir("assets/mod1/minecraft/models", 0755)
	fs.AddDir("assets/mod2", 0755)
	fs.AddDir("assets/mod2/minecraft", 0755)
	fs.AddDir("assets/mod2/minecraft/models", 0755)
	addModel(fs, "mod2/minecraft", "block/cube", []byte(testModelCube))

	nsToRoots := map[string][]string{
		"minecraft": {"assets/mod1/minecraft", "assets/mod2/minecraft"},
	}
	svc := NewModelParser(fs)

	model, err := svc.Run("minecraft:block/cube", nsToRoots)
	require.NoError(t, err)
	require.Len(t, model.Elements, 1)
}

func TestModelParser_ParseElements(t *testing.T) {
	t.Parallel()

	svc := &ModelParser{}

	tests := []struct {
		name string
		raws []json.RawMessage
		want []model.ModelElement
	}{
		{
			name: "nil raws returns empty slice",
			raws: nil,
			want: []model.ModelElement{},
		},
		{
			name: "empty raws",
			raws: []json.RawMessage{},
			want: []model.ModelElement{},
		},
		{
			name: "element with implicit shade defaults to true",
			raws: []json.RawMessage{
				json.RawMessage(`{"from":[0,0,0],"to":[16,16,16]}`),
			},
			want: []model.ModelElement{
				{From: model.Vector3{0, 0, 0}, To: model.Vector3{16, 16, 16}, Shade: true},
			},
		},
		{
			name: "element with explicit shade:true",
			raws: []json.RawMessage{
				json.RawMessage(`{"from":[0,0,0],"to":[16,16,16],"shade":true}`),
			},
			want: []model.ModelElement{
				{From: model.Vector3{0, 0, 0}, To: model.Vector3{16, 16, 16}, Shade: true},
			},
		},
		{
			name: "element with shade:false",
			raws: []json.RawMessage{
				json.RawMessage(`{"from":[0,0,0],"to":[16,16,16],"shade":false}`),
			},
			want: []model.ModelElement{
				{From: model.Vector3{0, 0, 0}, To: model.Vector3{16, 16, 16}, Shade: false},
			},
		},
		{
			name: "multiple elements with mixed shade",
			raws: []json.RawMessage{
				json.RawMessage(`{"from":[0,0,0],"to":[8,16,16],"shade":true}`),
				json.RawMessage(`{"from":[8,0,0],"to":[16,16,16]}`),
			},
			want: []model.ModelElement{
				{From: model.Vector3{0, 0, 0}, To: model.Vector3{8, 16, 16}, Shade: true},
				{From: model.Vector3{8, 0, 0}, To: model.Vector3{16, 16, 16}, Shade: true},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.parseElements(tt.raws)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestModelParser_ReadRaw(t *testing.T) {
	tests := []struct {
		name      string
		modelName string
		setup     func(fs *runtime.FSMock)
		wantErr   string
	}{
		{
			name:      "model found",
			modelName: "minecraft:block/cube",
			setup: func(fs *runtime.FSMock) {
				addModel(fs, "minecraft", "block/cube", []byte(testModelCube))
			},
		},
		{
			name:      "unknown namespace",
			modelName: "unknown:block/stone",
			setup:     func(fs *runtime.FSMock) {},
			wantErr:   "unknown namespace",
		},
		{
			name:      "file not found",
			modelName: "minecraft:block/nonexistent",
			setup:     func(fs *runtime.FSMock) {},
			wantErr:   "read",
		},
		{
			name:      "invalid JSON",
			modelName: "minecraft:block/bad_json",
			setup: func(fs *runtime.FSMock) {
				addModel(fs, "minecraft", "block/bad_json", []byte(`{bad json`))
			},
			wantErr: "parse",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs, nsToRoots := setupTestFS()
			tt.setup(fs)
			svc := NewModelParser(fs)
			raw, err := svc.readRaw(tt.modelName, nsToRoots)
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, raw)
			require.NotNil(t, raw.Elements)
			require.NotNil(t, raw.Textures)
		})
	}
}

func TestFlattenedModelMerge(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		base  *flattenedModel
		other *flattenedModel
		want  *flattenedModel
	}{
		{
			name:  "merge empty into empty",
			base:  &flattenedModel{Textures: map[string]string{}},
			other: &flattenedModel{Textures: map[string]string{}},
			want:  &flattenedModel{Textures: map[string]string{}},
		},
		{
			name:  "merge non-empty into empty",
			base:  &flattenedModel{Textures: make(map[string]string)},
			other: &flattenedModel{Elements: []model.ModelElement{{Shade: true}}, Textures: map[string]string{"a": "1"}},
			want:  &flattenedModel{Elements: []model.ModelElement{{Shade: true}}, Textures: map[string]string{"a": "1"}},
		},
		{
			name:  "merge into existing",
			base:  &flattenedModel{Elements: []model.ModelElement{{Shade: true}}, Textures: map[string]string{"a": "1"}},
			other: &flattenedModel{Elements: []model.ModelElement{{Shade: false}}, Textures: map[string]string{"b": "2"}},
			want:  &flattenedModel{Elements: []model.ModelElement{{Shade: true}, {Shade: false}}, Textures: map[string]string{"a": "1", "b": "2"}},
		},
		{
			name:  "texture overwrite",
			base:  &flattenedModel{Textures: map[string]string{"key": "old"}},
			other: &flattenedModel{Textures: map[string]string{"key": "new"}},
			want:  &flattenedModel{Textures: map[string]string{"key": "new"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.base.merge(tt.other)
			require.Equal(t, tt.want, tt.base)
		})
	}
}

func TestRawModelNormalize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  rawModel
		want rawModel
	}{
		{
			name: "nil elements and textures",
			raw:  rawModel{},
			want: rawModel{Elements: []json.RawMessage{}, Textures: map[string]string{}},
		},
		{
			name: "already initialized",
			raw:  rawModel{Elements: []json.RawMessage{}, Textures: map[string]string{}},
			want: rawModel{Elements: []json.RawMessage{}, Textures: map[string]string{}},
		},
		{
			name: "non-empty elements preserved",
			raw:  rawModel{Elements: []json.RawMessage{json.RawMessage(`{}`)}, Textures: map[string]string{"a": "b"}},
			want: rawModel{Elements: []json.RawMessage{json.RawMessage(`{}`)}, Textures: map[string]string{"a": "b"}},
		},
		{
			name: "nil elements only",
			raw:  rawModel{Textures: map[string]string{"k": "v"}},
			want: rawModel{Elements: []json.RawMessage{}, Textures: map[string]string{"k": "v"}},
		},
		{
			name: "nil textures only",
			raw:  rawModel{Elements: []json.RawMessage{}},
			want: rawModel{Elements: []json.RawMessage{}, Textures: map[string]string{}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.raw.normalize()
			require.Equal(t, tt.want, tt.raw)
		})
	}
}

func TestSplitModelName(t *testing.T) {
	tests := []struct {
		name      string
		modelName string
		wantNS    string
		wantPath  string
	}{
		{name: "with namespace", modelName: "minecraft:block/stone", wantNS: "minecraft", wantPath: "block/stone"},
		{name: "without namespace", modelName: "block/stone", wantNS: "minecraft", wantPath: "block/stone"},
		{name: "custom namespace", modelName: "custom:block/foo", wantNS: "custom", wantPath: "block/foo"},
		{name: "multi-colon", modelName: "a:b:c", wantNS: "a", wantPath: "b:c"},
	}
	svc := &ModelParser{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns, path := svc.splitModelName(tt.modelName)
			require.Equal(t, tt.wantNS, ns)
			require.Equal(t, tt.wantPath, path)
		})
	}
}
