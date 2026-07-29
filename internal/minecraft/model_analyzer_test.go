package minecraft

import (
	"testing"

	"mc2lua/internal/model"
	"github.com/stretchr/testify/require"
)

func TestModelAnalyzer_IsZero(t *testing.T) {
	t.Parallel()

	svc := &ModelAnalyzer{}

	tests := []struct {
		name     string
		v        []float64
		x, y, z  float64
		want     bool
	}{
		{name: "all match", v: []float64{0, 0, 0}, x: 0, y: 0, z: 0, want: true},
		{name: "x mismatch", v: []float64{1, 0, 0}, x: 0, y: 0, z: 0, want: false},
		{name: "y mismatch", v: []float64{0, 1, 0}, x: 0, y: 0, z: 0, want: false},
		{name: "z mismatch", v: []float64{0, 0, 1}, x: 0, y: 0, z: 0, want: false},
		{name: "all 16", v: []float64{16, 16, 16}, x: 16, y: 16, z: 16, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := svc.isZero(tt.v, tt.x, tt.y, tt.z)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestModelAnalyzerRun(t *testing.T) {
	svc := NewModelAnalyzer()

	tests := []struct {
		name     string
		elements []model.ModelElement
		want     bool
	}{
		{
			name:     "full block",
			elements: []model.ModelElement{{From: model.Vector3{0, 0, 0}, To: model.Vector3{16, 16, 16}, Shade: true}},
			want:     true,
		},
		{
			name:     "zero elements",
			elements: []model.ModelElement{},
			want:     false,
		},
		{
			name:     "nil elements",
			elements: nil,
			want:     false,
		},
		{
			name:     "two elements",
			elements: []model.ModelElement{{}, {}},
			want:     false,
		},
		{
			name:     "from not zero",
			elements: []model.ModelElement{{From: model.Vector3{1, 0, 0}, To: model.Vector3{16, 16, 16}, Shade: true}},
			want:     false,
		},
		{
			name:     "to not 16",
			elements: []model.ModelElement{{From: model.Vector3{0, 0, 0}, To: model.Vector3{15, 16, 16}, Shade: true}},
			want:     false,
		},
		{
			name:     "has rotation",
			elements: []model.ModelElement{{From: model.Vector3{0, 0, 0}, To: model.Vector3{16, 16, 16}, Rotation: &model.ElementRotation{}, Shade: true}},
			want:     false,
		},
		{
			name:     "shade false",
			elements: []model.ModelElement{{From: model.Vector3{0, 0, 0}, To: model.Vector3{16, 16, 16}, Shade: false}},
			want:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.Run(tt.elements)
			require.Equal(t, tt.want, got)
		})
	}
}
