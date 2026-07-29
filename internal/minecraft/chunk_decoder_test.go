package minecraft

import (
	"testing"

	"github.com/Tnze/go-mc/nbt"
	"github.com/Tnze/go-mc/save"
	"github.com/stretchr/testify/require"
)

func TestChunkDecoder_NewBitStorage(t *testing.T) {
	t.Parallel()

	svc := &ChunkDecoder{}

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

func TestChunkDecoder_DecodeBlock(t *testing.T) {
	t.Parallel()

	svc := &ChunkDecoder{}
	stonePalette := []save.BlockState{{Name: "minecraft:stone"}}

	tests := []struct {
		name     string
		palette  []save.BlockState
		palSize  int
		data     []uint64
		j        int
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
			wantNil: true,
		},
		{
			name:    "air block",
			palette: []save.BlockState{{Name: "minecraft:air"}},
			palSize: 1,
			j:       0,
			wantNil: true,
		},
		{
			name:    "cave_air block",
			palette: []save.BlockState{{Name: "minecraft:cave_air"}},
			palSize: 1,
			j:       0,
			wantNil: true,
		},
		{
			name:    "void_air block",
			palette: []save.BlockState{{Name: "minecraft:void_air"}},
			palSize: 1,
			j:       0,
			wantNil: true,
		},
		{
			name:    "stone block",
			palette: stonePalette,
			palSize: 1,
			j:       0,
			wantNil: false,
			wantID:  "minecraft:stone",
		},
		{
			name: "block with zero properties type returns nil props",
			palette: []save.BlockState{{
				Name:       "minecraft:stone",
				Properties: nbt.RawMessage{Type: 0, Data: nil},
			}},
			palSize: 1,
			j:       0,
			wantNil: false,
			wantID:  "minecraft:stone",
		},
		{
			name: "block with invalid properties data returns nil props",
			palette: []save.BlockState{{
				Name: "minecraft:stone",
				Properties: nbt.RawMessage{
					Type: 0x0A,
					Data: []byte{0xFF, 0xFF, 0xFF},
				},
			}},
			palSize: 1,
			j:       0,
			wantNil: false,
			wantID:  "minecraft:stone",
		},
		{
			name: "block with empty valid properties returns nil props",
			palette: []save.BlockState{{
				Name: "minecraft:stone",
				Properties: nbt.RawMessage{
					Type: 0x0A,
					Data: []byte{0x00},
				},
			}},
			palSize: 1,
			j:       0,
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
			block := svc.decodeBlock(bs, tt.j, tt.palette, 0, 0, 0)

			if tt.wantNil {
				require.Nil(t, block)
				return
			}
			require.NotNil(t, block)
			require.Equal(t, tt.wantID, block.ID)
			if tt.wantProp != nil {
				require.Equal(t, tt.wantProp, block.Props)
			} else {
				require.Nil(t, block.Props)
			}
		})
	}
}
