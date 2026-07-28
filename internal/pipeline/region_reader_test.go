package pipeline

import (
	"io/fs"
	"testing"

	"github.com/Tnze/go-mc/nbt"
	"github.com/Tnze/go-mc/save"
	"github.com/stretchr/testify/require"

	"mc2lua/internal/model"
	"mc2lua/internal/runtime"
)

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
		name          string
		min1, max1    int
		min2, max2    int
		want          bool
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

func TestRegionReader_NewBitStorage(t *testing.T) {
	t.Parallel()

	svc := &RegionReader{}

	tests := []struct {
		name    string
		palSize int
		data    []uint64
		wantLen int
	}{
		{name: "palette size 1", palSize: 1, data: nil, wantLen: 4096},
		{name: "palette size 2", palSize: 2, data: nil, wantLen: 4096},
		{name: "palette size 4", palSize: 4, data: nil, wantLen: 4096},
		{name: "palette size 16", palSize: 16, data: nil, wantLen: 4096},
		{name: "palette size 17", palSize: 17, data: nil, wantLen: 4096},
		{name: "palette size 256", palSize: 256, data: nil, wantLen: 4096},
		{name: "palette size 257", palSize: 257, data: nil, wantLen: 4096},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			bs := svc.newBitStorage(tt.palSize, tt.data)
			require.NotNil(t, bs)
			require.Equal(t, tt.wantLen, bs.Len())
		})
	}
}

func TestRegionReader_DecodeBlock(t *testing.T) {
	t.Parallel()

	svc := &RegionReader{}
	defaultBounds := model.Bounds{
		XMin: -1000, XMax: 1000,
		YMin: -1000, YMax: 1000,
		ZMin: -1000, ZMax: 1000,
	}
	stonePalette := []save.BlockState{{Name: "minecraft:stone"}}

	tests := []struct {
		name     string
		palette  []save.BlockState
		palSize  int
		data     []uint64
		j        int
		bounds   model.Bounds
		wantNil  bool
		wantID   string
		wantProp map[string]string
	}{
		{
			name:    "index out of range",
			palette: []save.BlockState{{Name: "minecraft:stone"}},
			palSize: 2,
			data: func() []uint64 {
				d := make([]uint64, 256)
				d[0] = 1
				return d
			}(),
			j:       0,
			bounds:  defaultBounds,
			wantNil: true,
		},
		{
			name:    "air block",
			palette: []save.BlockState{{Name: "minecraft:air"}},
			palSize: 1,
			j:       0,
			bounds:  defaultBounds,
			wantNil: true,
		},
		{
			name:    "cave_air block",
			palette: []save.BlockState{{Name: "minecraft:cave_air"}},
			palSize: 1,
			j:       0,
			bounds:  defaultBounds,
			wantNil: true,
		},
		{
			name:    "void_air block",
			palette: []save.BlockState{{Name: "minecraft:void_air"}},
			palSize: 1,
			j:       0,
			bounds:  defaultBounds,
			wantNil: true,
		},
		{
			name:    "block outside X bounds",
			palette: stonePalette,
			palSize: 1,
			j:       0,
			bounds:  model.Bounds{XMin: 1, XMax: 10, YMin: -1000, YMax: 1000, ZMin: -1000, ZMax: 1000},
			wantNil: true,
		},
		{
			name:    "block outside Y bounds",
			palette: stonePalette,
			palSize: 1,
			j:       0,
			bounds:  model.Bounds{XMin: -1000, XMax: 1000, YMin: 1, YMax: 10, ZMin: -1000, ZMax: 1000},
			wantNil: true,
		},
		{
			name:    "block outside Z bounds",
			palette: stonePalette,
			palSize: 1,
			j:       0,
			bounds:  model.Bounds{XMin: -1000, XMax: 1000, YMin: -1000, YMax: 1000, ZMin: 1, ZMax: 10},
			wantNil: true,
		},
		{
			name:    "stone inside bounds no properties",
			palette: stonePalette,
			palSize: 1,
			j:       0,
			bounds:  defaultBounds,
			wantNil: false,
			wantID:  "minecraft:stone",
		},
		{
			name: "block with properties",
			palette: []save.BlockState{{
				Name: "minecraft:stairs",
				Properties: nbt.RawMessage{
					Type: 0x0A,
					Data: []byte{
						0x08, 0x00, 0x06, 'f', 'a', 'c', 'i', 'n', 'g', 0x00, 0x05, 'n', 'o', 'r', 't', 'h',
						0x00,
					},
				},
			}},
			palSize: 1,
			j:       0,
			bounds:  defaultBounds,
			wantNil: false,
			wantID:  "minecraft:stairs",
			wantProp: map[string]string{"facing": "north"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			bs := svc.newBitStorage(tt.palSize, tt.data)
			block := svc.decodeBlock(bs, tt.j, tt.palette, 0, 0, 0, tt.bounds)

			if tt.wantNil {
				require.Nil(t, block)
				return
			}
			require.NotNil(t, block)
			require.Equal(t, tt.wantID, block.ID)
			if tt.wantProp != nil {
				require.Equal(t, tt.wantProp, block.Properties)
			} else {
				require.Nil(t, block.Properties)
			}
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
			svc := NewRegionReader(m)

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
			name: "region in bounds but file invalid returns empty blocks",
			setup: func(m *runtime.FSMock) {
				m.AddFile("/input/r.0.0.mca", []byte("not a real mca file"), 0644)
			},
			bounds: model.Bounds{
				XMin: -1000, XMax: 1000,
				YMin: -1000, YMax: 1000,
				ZMin: -1000, ZMax: 1000,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := runtime.NewFSMock()
			tt.setup(m)
			svc := NewRegionReader(m)

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
