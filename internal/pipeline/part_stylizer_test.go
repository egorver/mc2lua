package pipeline

import (
	"testing"

	"mc2lua/internal/model"
	"mc2lua/internal/stateful"

	"github.com/stretchr/testify/require"
)

type mockPartStyleMatcher struct {
	styles map[string]model.PartStyle
}

func (m *mockPartStyleMatcher) Run(blockID string) (model.PartStyle, bool) {
	s, ok := m.styles[blockID]
	return s, ok
}

func floatPtr(v float64) *float64 {
	return &v
}

func surfPtr(texture string, color *model.Color, transparency *float64) *model.Surface {
	return &model.Surface{Texture: texture, Color: color, Transparency: transparency}
}

func testPart(blockID string, visible model.FaceMask) model.Part {
	return model.Part{
		BlockID:      blockID,
		Color:        model.Color{128, 128, 128},
		Material:     model.DefaultMaterial,
		VisibleFaces: visible,
	}
}

func indexWithRotation(id string, rotX, rotY float64) stateful.StyleIndex {
	idx := stateful.NewStyleIndex()
	idx.Add(id, "", model.StyledBlock{ID: id, PropsKey: "", RotX: rotX, RotY: rotY})
	return *idx
}

func TestPartStylizer_New(t *testing.T) {
	t.Parallel()

	svc := NewPartStylizer(&mockPartStyleMatcher{})
	require.NotNil(t, svc)
}

func TestPartStylizer_Run(t *testing.T) {
	t.Parallel()

	red := model.Color{200, 30, 30}
	green := model.Color{30, 200, 30}

	tests := []struct {
		name       string
		styles     map[string]model.PartStyle
		parts      []model.Part
		styleIndex stateful.StyleIndex
		wantCheck  func(t *testing.T, parts []model.Part)
	}{
		{
			name:   "no style leaves part unchanged",
			styles: map[string]model.PartStyle{},
			parts:  []model.Part{testPart("minecraft:stone", visibleMask())},
			wantCheck: func(t *testing.T, parts []model.Part) {
				require.Equal(t, testPart("minecraft:stone", visibleMask()), parts[0])
			},
		},
		{
			name: "applies transparency, surfaces and material",
			styles: map[string]model.PartStyle{
				"minecraft:grass_block": {
					Transparency: floatPtr(0.5),
					Top:          surfPtr("rbxassetid://1", &red, nil),
					All:          surfPtr("rbxassetid://2", &green, nil),
				},
			},
			parts: []model.Part{testPart("minecraft:grass_block", visibleMask())},
			wantCheck: func(t *testing.T, parts []model.Part) {
				p := parts[0]
				require.Equal(t, float64(0.5), *p.Transparency)
				require.Equal(t, &model.Surface{Texture: "rbxassetid://1", Color: &red}, p.Top)
				require.Equal(t, &model.Surface{Texture: "rbxassetid://2", Color: &green}, p.Bottom)
				require.Equal(t, &model.Surface{Texture: "rbxassetid://2", Color: &green}, p.Front)
				require.Equal(t, model.TexturelessMaterial, p.Material)
			},
		},
		{
			name: "transparency absent when not set in style",
			styles: map[string]model.PartStyle{
				"minecraft:stone": {All: surfPtr("rbxassetid://1", &red, nil)},
			},
			parts: []model.Part{testPart("minecraft:stone", visibleMask())},
			wantCheck: func(t *testing.T, parts []model.Part) {
				require.Nil(t, parts[0].Transparency)
			},
		},
		{
			name: "hidden faces are not styled",
			styles: map[string]model.PartStyle{
				"minecraft:stone": {
					Top: surfPtr("rbxassetid://top", &red, nil),
					All: surfPtr("rbxassetid://all", &green, nil),
				},
			},
			parts: []model.Part{testPart("minecraft:stone", model.FaceMask{false, true, true, true, true, true})},
			wantCheck: func(t *testing.T, parts []model.Part) {
				require.Nil(t, parts[0].Top)
				require.Equal(t, &model.Surface{Texture: "rbxassetid://all", Color: &green}, parts[0].Bottom)
				require.Equal(t, &model.Surface{Texture: "rbxassetid://all", Color: &green}, parts[0].Front)
			},
		},
		{
			name: "all faces hidden produces no surfaces",
			styles: map[string]model.PartStyle{
				"minecraft:stone": {All: surfPtr("rbxassetid://1", &red, nil)},
			},
			parts: []model.Part{testPart("minecraft:stone", model.FaceMask{})},
			wantCheck: func(t *testing.T, parts []model.Part) {
				p := parts[0]
				require.Nil(t, p.Top)
				require.Nil(t, p.Bottom)
				require.Nil(t, p.Front)
				require.Nil(t, p.Back)
				require.Nil(t, p.Left)
				require.Nil(t, p.Right)
				require.Equal(t, model.DefaultMaterial, p.Material)
			},
		},
		{
			name: "style without texture creates no surfaces",
			styles: map[string]model.PartStyle{
				"minecraft:stone": {Top: surfPtr("", &red, nil)},
			},
			parts: []model.Part{testPart("minecraft:stone", visibleMask())},
			wantCheck: func(t *testing.T, parts []model.Part) {
				require.Equal(t, model.DefaultMaterial, parts[0].Material)
				require.Nil(t, parts[0].Top)
			},
		},
		{
			name: "rotation maps styled face to world face",
			styles: map[string]model.PartStyle{
				"minecraft:furnace": {Front: surfPtr("rbxassetid://front", &red, nil)},
			},
			parts:      []model.Part{testPart("minecraft:furnace", visibleMask())},
			styleIndex: indexWithRotation("minecraft:furnace", 0, 90),
			wantCheck: func(t *testing.T, parts []model.Part) {
				p := parts[0]
				require.Equal(t, &model.Surface{Texture: "rbxassetid://front", Color: &red}, p.Right)
				require.Nil(t, p.Front)
				require.Nil(t, p.Back)
				require.Nil(t, p.Left)
				require.Equal(t, model.TexturelessMaterial, p.Material)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewPartStylizer(&mockPartStyleMatcher{styles: tt.styles})
			got := svc.Run(tt.parts, tt.styleIndex)
			require.Len(t, got, len(tt.parts))
			tt.wantCheck(t, got)
		})
	}
}

func TestPartStylizer_ResolveSurface(t *testing.T) {
	t.Parallel()

	red := model.Color{200, 30, 30}
	green := model.Color{30, 200, 30}
	svc := NewPartStylizer(nil)

	tests := []struct {
		name  string
		style model.PartStyle
		face  string
		want  *model.Surface
	}{
		{
			name:  "no style returns nil",
			style: model.PartStyle{},
			face:  model.FaceTop,
			want:  nil,
		},
		{
			name: "all fills every face",
			style: model.PartStyle{
				All: surfPtr("rbxassetid://a", &red, nil),
			},
			face: model.FaceTop,
			want: &model.Surface{Texture: "rbxassetid://a", Color: &red},
		},
		{
			name: "explicit face overrides all per field",
			style: model.PartStyle{
				Top: surfPtr("rbxassetid://t", &green, nil),
				All: surfPtr("rbxassetid://a", &red, floatPtr(0.2)),
			},
			face: model.FaceTop,
			want: &model.Surface{Texture: "rbxassetid://t", Color: &green, Transparency: floatPtr(0.2)},
		},
		{
			name: "sides override all on side faces only",
			style: model.PartStyle{
				Sides: surfPtr("rbxassetid://s", nil, nil),
				All:   surfPtr("rbxassetid://a", &red, nil),
			},
			face: model.FaceFront,
			want: &model.Surface{Texture: "rbxassetid://s", Color: &red},
		},
		{
			name: "sides do not affect top face",
			style: model.PartStyle{
				Sides: surfPtr("rbxassetid://s", nil, nil),
				All:   surfPtr("rbxassetid://a", &red, nil),
			},
			face: model.FaceTop,
			want: &model.Surface{Texture: "rbxassetid://a", Color: &red},
		},
		{
			name: "walls behave like sides",
			style: model.PartStyle{
				Walls: surfPtr("rbxassetid://w", nil, nil),
				All:   surfPtr("rbxassetid://a", &red, nil),
			},
			face: model.FaceLeft,
			want: &model.Surface{Texture: "rbxassetid://w", Color: &red},
		},
		{
			name: "faces group covers all faces",
			style: model.PartStyle{
				Faces: surfPtr("rbxassetid://f", nil, nil),
				All:   surfPtr("rbxassetid://a", &red, nil),
			},
			face: model.FaceBottom,
			want: &model.Surface{Texture: "rbxassetid://f", Color: &red},
		},
		{
			name: "missing fields are filled from lower priority source",
			style: model.PartStyle{
				Top: &model.Surface{Texture: "rbxassetid://t"},
				All: &model.Surface{Color: &red},
			},
			face: model.FaceTop,
			want: &model.Surface{Texture: "rbxassetid://t", Color: &red},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.resolveSurface(tt.style, tt.face)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestPartStylizer_MapFace(t *testing.T) {
	t.Parallel()

	svc := NewPartStylizer(nil)

	tests := []struct {
		name string
		face string
		rotX float64
		rotY float64
		want int
	}{
		{name: "top identity", face: model.FaceTop, want: model.FaceIndexTop},
		{name: "bottom identity", face: model.FaceBottom, want: model.FaceIndexBottom},
		{name: "front is north", face: model.FaceFront, want: model.FaceIndexBack},
		{name: "back is south", face: model.FaceBack, want: model.FaceIndexFront},
		{name: "left is west", face: model.FaceLeft, want: model.FaceIndexLeft},
		{name: "right is east", face: model.FaceRight, want: model.FaceIndexRight},
		{name: "rotY 90 front to east", face: model.FaceFront, rotY: 90, want: model.FaceIndexRight},
		{name: "rotY 90 back to west", face: model.FaceBack, rotY: 90, want: model.FaceIndexLeft},
		{name: "rotY 90 left to north", face: model.FaceLeft, rotY: 90, want: model.FaceIndexBack},
		{name: "rotY 90 right to south", face: model.FaceRight, rotY: 90, want: model.FaceIndexFront},
		{name: "rotY 180 front to south", face: model.FaceFront, rotY: 180, want: model.FaceIndexFront},
		{name: "rotY 270 front to west", face: model.FaceFront, rotY: 270, want: model.FaceIndexLeft},
		{name: "rotX 90 front to top", face: model.FaceFront, rotX: 90, want: model.FaceIndexTop},
		{name: "rotX 90 top to south", face: model.FaceTop, rotX: 90, want: model.FaceIndexFront},
		{name: "rotX 90 bottom to north", face: model.FaceBottom, rotX: 90, want: model.FaceIndexBack},
		{name: "rotX 180 top to bottom", face: model.FaceTop, rotX: 180, want: model.FaceIndexBottom},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.mapFace(tt.face, tt.rotX, tt.rotY)
			require.Equal(t, tt.want, got)
		})
	}
}
