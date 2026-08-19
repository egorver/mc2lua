package minecraft

import (
	"math"
	"testing"

	"mc2lua/internal/model"

	"github.com/stretchr/testify/require"
)

func TestElementRotator_New(t *testing.T) {
	t.Parallel()

	r := NewElementRotator()
	require.NotNil(t, r)
}

func TestElementRotator_RoundCoord(t *testing.T) {
	t.Parallel()

	svc := NewElementRotator()
	tests := []struct {
		name string
		v    float64
		want float64
	}{
		{name: "zero", v: 0, want: 0},
		{name: "one", v: 1, want: 1},
		{name: "half", v: 0.5, want: 0.5},
		{name: "quarter", v: 0.25, want: 0.25},
		{name: "negative", v: -0.5, want: -0.5},
		{name: "sixteenth", v: 1.0 / 16, want: 0.0625},
		{name: "sixtyfourth", v: 1.0 / 64, want: 0.015625},
		{name: "three sixtyfourths", v: 3.0 / 64, want: 0.046875},
		{name: "pi precision", v: math.Pi, want: math.Round(math.Pi*64) / 64},
		{name: "large", v: 16, want: 16},
		{name: "large negative", v: -16, want: -16},
		{name: "sub 64th rounds down", v: 0.001, want: 0},
		{name: "sub 64th rounds up", v: 0.02, want: 0.015625},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := svc.roundCoord(tt.v)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestElementRotator_RotatePoint(t *testing.T) {
	t.Parallel()

	svc := NewElementRotator()
	tests := []struct {
		name string
		p    model.Vector3
		rotX float64
		rotY float64
		want model.Vector3
	}{
		{
			name: "no rotation",
			p:    model.Vector3{4, 4, 4},
			want: model.Vector3{4, 4, 4},
		},
		{
			name: "rotY 90 around center",
			p:    model.Vector3{12, 8, 8},
			rotY: 90,
			want: model.Vector3{8, 8, 12},
		},
		{
			name: "rotY 180 around center",
			p:    model.Vector3{12, 8, 8},
			rotY: 180,
			want: model.Vector3{4, 8, 8},
		},
		{
			name: "rotY -90 around center",
			p:    model.Vector3{12, 8, 8},
			rotY: -90,
			want: model.Vector3{8, 8, 4},
		},
		{
			name: "rotX 90 around center",
			p:    model.Vector3{8, 12, 8},
			rotX: 90,
			want: model.Vector3{8, 8, 12},
		},
		{
			name: "rotX 180 around center",
			p:    model.Vector3{8, 12, 8},
			rotX: 180,
			want: model.Vector3{8, 4, 8},
		},
		{
			name: "point at center unchanged",
			p:    model.Vector3{8, 8, 8},
			rotX: 90,
			rotY: 90,
			want: model.Vector3{8, 8, 8},
		},
		{
			name: "both axes",
			p:    model.Vector3{12, 12, 4},
			rotX: 90,
			rotY: 90,
			want: model.Vector3{4, 12, 12},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := svc.rotatePoint(tt.p, tt.rotX, tt.rotY)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestElementRotator_RunStyled(t *testing.T) {
	t.Parallel()

	svc := NewElementRotator()
	elem := model.StyledElement{
		From:  model.Vector3{0, 0, 0},
		To:    model.Vector3{16, 16, 16},
		Shade: true,
	}
	elemWithRotation := model.StyledElement{
		From:  model.Vector3{0, 0, 0},
		To:    model.Vector3{16, 16, 16},
		Shade: true,
		Rotation: &model.ElementRotation{
			Origin: model.Vector3{4, 8, 8},
			Axis:   "y",
			Angle:  45,
		},
	}

	tests := []struct {
		name     string
		elements []model.StyledElement
		rotX     float64
		rotY     float64
		wantLen  int
		check    func(t *testing.T, result []model.StyledElement)
	}{
		{
			name:     "empty elements",
			elements: nil,
			wantLen:  0,
		},
		{
			name:     "no rotation returns copy",
			elements: []model.StyledElement{elem},
			wantLen:  1,
			check: func(t *testing.T, result []model.StyledElement) {
				require.Equal(t, model.Vector3{0, 0, 0}, result[0].From)
				require.Equal(t, model.Vector3{16, 16, 16}, result[0].To)
			},
		},
		{
			name:     "rotY 90 swaps x and z",
			elements: []model.StyledElement{elem},
			rotY:     90,
			wantLen:  1,
			check: func(t *testing.T, result []model.StyledElement) {
				require.Equal(t, model.Vector3{0, 0, 0}, result[0].From)
				require.Equal(t, model.Vector3{16, 16, 16}, result[0].To)
			},
		},
		{
			name: "from and to swap after rotation",
			elements: []model.StyledElement{
				{From: model.Vector3{12, 0, 0}, To: model.Vector3{16, 16, 16}, Shade: true},
			},
			rotY:    90,
			wantLen: 1,
			check: func(t *testing.T, result []model.StyledElement) {
				require.Equal(t, model.Vector3{0, 0, 12}, result[0].From)
				require.Equal(t, model.Vector3{16, 16, 16}, result[0].To)
			},
		},
		{
			name:     "element with sub-rotation origin updates",
			elements: []model.StyledElement{elemWithRotation},
			rotY:     90,
			wantLen:  1,
			check: func(t *testing.T, result []model.StyledElement) {
				require.NotNil(t, result[0].Rotation)
				require.Equal(t, "y", result[0].Rotation.Axis)
				require.Equal(t, 45.0, result[0].Rotation.Angle)
			},
		},
		{
			name:     "rotX 90 changes y and z",
			elements: []model.StyledElement{elem},
			rotX:     90,
			wantLen:  1,
			check: func(t *testing.T, result []model.StyledElement) {
				require.Equal(t, model.Vector3{0, 0, 0}, result[0].From)
				require.Equal(t, model.Vector3{16, 16, 16}, result[0].To)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := svc.RunStyled(tt.elements, tt.rotX, tt.rotY)
			require.Len(t, got, tt.wantLen)
			if tt.check != nil {
				tt.check(t, got)
			}
		})
	}
}

func TestElementRotator_RunModel(t *testing.T) {
	t.Parallel()

	svc := NewElementRotator()
	asymmetric := model.ModelElement{From: model.Vector3{0, 0, 0}, To: model.Vector3{8, 8, 8}, Shade: true}
	elemWithRotation := model.ModelElement{
		From:  model.Vector3{0, 0, 0},
		To:    model.Vector3{8, 8, 8},
		Shade: true,
		Rotation: &model.ElementRotation{
			Origin: model.Vector3{4, 8, 8},
			Axis:   "y",
			Angle:  45,
		},
	}

	tests := []struct {
		name     string
		elements []model.ModelElement
		rotX     float64
		rotY     float64
		wantLen  int
		check    func(t *testing.T, result []model.ModelElement)
	}{
		{
			name:     "empty elements",
			elements: nil,
			wantLen:  0,
		},
		{
			name:     "no rotation returns original slice",
			elements: []model.ModelElement{asymmetric},
			wantLen:  1,
			check: func(t *testing.T, result []model.ModelElement) {
				require.Equal(t, model.Vector3{0, 0, 0}, result[0].From)
				require.Equal(t, model.Vector3{8, 8, 8}, result[0].To)
			},
		},
		{
			name:     "rotY 90 rotates from and to",
			elements: []model.ModelElement{asymmetric},
			rotY:     90,
			wantLen:  1,
			check: func(t *testing.T, result []model.ModelElement) {
				require.Equal(t, model.Vector3{8, 0, 0}, result[0].From)
				require.Equal(t, model.Vector3{16, 8, 8}, result[0].To)
			},
		},
		{
			name:     "rotX 90 rotates from and to",
			elements: []model.ModelElement{asymmetric},
			rotX:     90,
			wantLen:  1,
			check: func(t *testing.T, result []model.ModelElement) {
				require.Equal(t, model.Vector3{0, 8, 0}, result[0].From)
				require.Equal(t, model.Vector3{8, 16, 8}, result[0].To)
			},
		},
		{
			name:     "sub-rotation origin transformed",
			elements: []model.ModelElement{elemWithRotation},
			rotY:     90,
			wantLen:  1,
			check: func(t *testing.T, result []model.ModelElement) {
				require.NotNil(t, result[0].Rotation)
				require.Equal(t, model.Vector3{8, 8, 4}, result[0].Rotation.Origin)
				require.Equal(t, "y", result[0].Rotation.Axis)
				require.Equal(t, 45.0, result[0].Rotation.Angle)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := svc.RunModel(tt.elements, tt.rotX, tt.rotY)
			require.Len(t, got, tt.wantLen)
			if tt.check != nil {
				tt.check(t, got)
			}
		})
	}
}
