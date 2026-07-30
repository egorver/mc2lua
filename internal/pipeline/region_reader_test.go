package pipeline

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"mc2lua/internal/model"
	"mc2lua/internal/runtime"

	"github.com/stretchr/testify/require"
)

type mockChunkDecoder struct {
	runFn func(data []byte, chunkX, chunkZ int) ([]model.RawBlock, error)
}

func (m *mockChunkDecoder) Run(data []byte, chunkX, chunkZ int) ([]model.RawBlock, error) {
	if m.runFn != nil {
		return m.runFn(data, chunkX, chunkZ)
	}
	return nil, nil
}

func TestRegionReader_ParseRegionCoord(t *testing.T) {
	t.Parallel()

	svc := &RegionReader{}

	tests := []struct {
		name   string
		input  string
		wantRX int
		wantRZ int
	}{
		{name: "positive coords", input: "r.0.0.mca", wantRX: 0, wantRZ: 0},
		{name: "negative rz", input: "r.-1.5.mca", wantRX: -1, wantRZ: 5},
		{name: "both negative", input: "r.-3.-7.mca", wantRX: -3, wantRZ: -7},
		{name: "large coords", input: "r.123.456.mca", wantRX: 123, wantRZ: 456},
		{name: "invalid format", input: "region.mca", wantRX: 0, wantRZ: 0},
		{name: "extra suffix", input: "r.1.2.mca.bak", wantRX: 1, wantRZ: 2},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rx, rz := svc.parseRegionCoord(tt.input)
			require.Equal(t, tt.wantRX, rx)
			require.Equal(t, tt.wantRZ, rz)
		})
	}
}

func TestRegionReader_RangeOverlap(t *testing.T) {
	t.Parallel()

	svc := &RegionReader{}

	tests := []struct {
		name       string
		min1, max1 int
		min2, max2 int
		want       bool
	}{
		{name: "full overlap", min1: 0, max1: 10, min2: 2, max2: 8, want: true},
		{name: "partial overlap", min1: 0, max1: 5, min2: 3, max2: 10, want: true},
		{name: "no overlap", min1: 0, max1: 5, min2: 6, max2: 10, want: false},
		{name: "touching at max", min1: 0, max1: 5, min2: 5, max2: 10, want: true},
		{name: "touching at min", min1: 5, max1: 10, min2: 0, max2: 5, want: true},
		{name: "equal ranges", min1: 0, max1: 10, min2: 0, max2: 10, want: true},
		{name: "completely before", min1: 0, max1: 2, min2: 5, max2: 8, want: false},
		{name: "completely after", min1: 5, max1: 8, min2: 0, max2: 2, want: false},
		{name: "negative overlap", min1: -10, max1: -5, min2: -8, max2: -3, want: true},
		{name: "negative no overlap", min1: -10, max1: -5, min2: -3, max2: 0, want: false},
		{name: "single point overlap", min1: 5, max1: 5, min2: 5, max2: 10, want: true},
		{name: "single point no overlap", min1: 5, max1: 5, min2: 6, max2: 10, want: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := svc.rangeOverlap(tt.min1, tt.max1, tt.min2, tt.max2)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestRegionReader_ListRegionFiles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		setup   func(m *runtime.FSMock)
		want    []string
		wantErr bool
	}{
		{
			name:  "empty directory",
			input: "/input",
			setup: func(m *runtime.FSMock) {},
			want:  []string{},
		},
		{
			name:  "only mca files",
			input: "/input",
			setup: func(m *runtime.FSMock) {
				m.AddFile("/input/r.0.0.mca", nil, 0644)
				m.AddFile("/input/r.-1.2.mca", nil, 0644)
			},
			want: []string{"r.0.0.mca", "r.-1.2.mca"},
		},
		{
			name:  "mixed files",
			input: "/input",
			setup: func(m *runtime.FSMock) {
				m.AddFile("/input/r.0.0.mca", nil, 0644)
				m.AddFile("/input/readme.txt", nil, 0644)
				m.AddFile("/input/r.1.1.mca", nil, 0644)
			},
			want: []string{"r.0.0.mca", "r.1.1.mca"},
		},
		{
			name:  "ignores subdirectories",
			input: "/input",
			setup: func(m *runtime.FSMock) {
				m.AddFile("/input/r.0.0.mca", nil, 0644)
				m.AddFile("/input/sub", nil, fs.ModeDir|0755)
			},
			want: []string{"r.0.0.mca"},
		},
		{
			name:  "non mca pattern",
			input: "/input",
			setup: func(m *runtime.FSMock) {
				m.AddFile("/input/region.bak", nil, 0644)
				m.AddFile("/input/data.mca.backup", nil, 0644)
				m.AddFile("/input/r.0.0.txt", nil, 0644)
			},
			want: []string{},
		},
		{
			name:    "readdir error",
			input:   "/input",
			setup:   func(m *runtime.FSMock) { m.ReadDirErrors["/input"] = fs.ErrPermission },
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := runtime.NewFSMock()
			tt.setup(m)
			svc := NewRegionReader(m, &mockChunkDecoder{})

			got, err := svc.listRegionFiles(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.ElementsMatch(t, tt.want, got)
		})
	}
}

func TestRegionReader_ProcessRegion_EmptyMCA(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	name := "r.0.0.mca"

	data := make([]byte, 8192)
	err := os.WriteFile(filepath.Join(dir, name), data, 0644)
	require.NoError(t, err)

	svc := NewRegionReader(runtime.NewFS(), &mockChunkDecoder{})

	blocks, err := svc.Run(dir, model.Bounds{
		XMin: -1000, XMax: 1000,
		YMin: -1000, YMax: 1000,
		ZMin: -1000, ZMax: 1000,
	})
	require.NoError(t, err)
	require.Empty(t, blocks)
}

func TestRegionReader_Run(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(m *runtime.FSMock)
		bounds  model.Bounds
		wantErr bool
	}{
		{
			name:  "empty input directory returns empty blocks",
			setup: func(m *runtime.FSMock) {},
			bounds: model.Bounds{
				XMin: -1000, XMax: 1000,
				YMin: -1000, YMax: 1000,
				ZMin: -1000, ZMax: 1000,
			},
		},
		{
			name: "region out of bounds returns empty blocks",
			setup: func(m *runtime.FSMock) {
				m.AddFile("/input/r.0.0.mca", nil, 0644)
			},
			bounds: model.Bounds{
				XMin: 10000, XMax: 20000,
				YMin: 0, YMax: 255,
				ZMin: 10000, ZMax: 20000,
			},
		},
		{
			name:    "readdir error returns error",
			setup:   func(m *runtime.FSMock) { m.ReadDirErrors["/input"] = fs.ErrPermission },
			bounds:  model.Bounds{XMin: -1000, XMax: 1000, YMin: -1000, YMax: 1000, ZMin: -1000, ZMax: 1000},
			wantErr: true,
		},
		{
			name: "region in bounds but file invalid returns partial result with error",
			setup: func(m *runtime.FSMock) {
				m.AddFile("/input/r.0.0.mca", []byte("not a real mca file"), 0644)
			},
			bounds: model.Bounds{
				XMin: -1000, XMax: 1000,
				YMin: -1000, YMax: 1000,
				ZMin: -1000, ZMax: 1000,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := runtime.NewFSMock()
			tt.setup(m)
			svc := NewRegionReader(m, &mockChunkDecoder{})

			blocks, err := svc.Run("/input", tt.bounds)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Empty(t, blocks)
		})
	}
}
