package minecraft

import (
	"bytes"
	"compress/zlib"
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

func TestChunkDecoder_DecodeSection(t *testing.T) {
	t.Parallel()

	svc := &ChunkDecoder{}

	tests := []struct {
		name    string
		section save.Section
		wantNil bool
		wantLen int
		wantY   int
		wantID  string
	}{
		{
			name: "empty palette returns nil",
			section: save.Section{
				Y: 0,
				BlockStates: save.PaletteContainer[save.BlockState]{
					Palette: []save.BlockState{},
				},
			},
			wantNil: true,
		},
		{
			name: "all air blocks returns nil",
			section: save.Section{
				Y: 0,
				BlockStates: save.PaletteContainer[save.BlockState]{
					Palette: []save.BlockState{{Name: "minecraft:air"}},
					Data:    []uint64{},
				},
			},
			wantNil: true,
		},
		{
			name: "cave_air blocks returns nil",
			section: save.Section{
				Y: 0,
				BlockStates: save.PaletteContainer[save.BlockState]{
					Palette: []save.BlockState{{Name: "minecraft:cave_air"}},
					Data:    []uint64{},
				},
			},
			wantNil: true,
		},
		{
			name: "void_air blocks returns nil",
			section: save.Section{
				Y: 0,
				BlockStates: save.PaletteContainer[save.BlockState]{
					Palette: []save.BlockState{{Name: "minecraft:void_air"}},
					Data:    []uint64{},
				},
			},
			wantNil: true,
		},
		{
			name: "stone blocks at Y=4",
			section: save.Section{
				Y: 4,
				BlockStates: save.PaletteContainer[save.BlockState]{
					Palette: []save.BlockState{{Name: "minecraft:stone"}},
					Data:    []uint64{},
				},
			},
			wantLen: 4096,
			wantY:   64,
			wantID:  "minecraft:stone",
		},
		{
			name: "stone blocks at negative Y",
			section: save.Section{
				Y: -4,
				BlockStates: save.PaletteContainer[save.BlockState]{
					Palette: []save.BlockState{{Name: "minecraft:stone"}},
					Data:    []uint64{},
				},
			},
			wantLen: 4096,
			wantY:   -64,
			wantID:  "minecraft:stone",
		},
		{
			name: "blocks with properties",
			section: save.Section{
				Y: 0,
				BlockStates: save.PaletteContainer[save.BlockState]{
					Palette: []save.BlockState{{
						Name: "minecraft:oak_log",
						Properties: nbt.RawMessage{
							Type: 0x0A,
							Data: []byte{
								0x08, 0x00, 0x04, 'a', 'x', 'i', 's', 0x00, 0x01, 'y',
								0x00,
							},
						},
					}},
					Data: []uint64{},
				},
			},
			wantLen: 4096,
			wantY:   0,
			wantID:  "minecraft:oak_log",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			blocks := svc.decodeSection(tt.section, 0, 0)

			if tt.wantNil {
				require.Nil(t, blocks)
				return
			}
			require.Len(t, blocks, tt.wantLen)
			require.Equal(t, tt.wantID, blocks[0].ID)
			require.Equal(t, tt.wantY, blocks[0].Y)
		})
	}
}

func TestChunkDecoder_Run_InvalidData(t *testing.T) {
	t.Parallel()

	svc := NewChunkDecoder()

	_, err := svc.Run([]byte{0xFF, 0xFF, 0xFF}, 0, 0)
	require.Error(t, err)
}

func encodeChunkNBT(chunk map[string]interface{}) ([]byte, error) {
	raw, err := nbt.Marshal(chunk)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteByte(2) // zlib compression
	w := zlib.NewWriter(&buf)
	if _, err := w.Write(raw); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func TestChunkDecoder_Run_NonFullStatus(t *testing.T) {
	t.Parallel()

	data, err := encodeChunkNBT(map[string]interface{}{
		"Status": "partial",
		"sections": []interface{}{
			map[string]interface{}{
				"Y": int8(0),
				"block_states": map[string]interface{}{
					"palette": []interface{}{
						map[string]interface{}{
							"Name": "minecraft:stone",
						},
					},
					"data": []int64{},
				},
			},
		},
	})
	require.NoError(t, err)

	svc := NewChunkDecoder()
	blocks, err := svc.Run(data, 0, 0)
	require.NoError(t, err)
	require.Nil(t, blocks)
}

func TestChunkDecoder_Run_FullChunk(t *testing.T) {
	t.Parallel()

	data, err := encodeChunkNBT(map[string]interface{}{
		"Status": "full",
		"sections": []interface{}{
			map[string]interface{}{
				"Y": int8(4),
				"block_states": map[string]interface{}{
					"palette": []interface{}{
						map[string]interface{}{
							"Name": "minecraft:stone",
						},
					},
					"data": []int64{},
				},
			},
		},
	})
	require.NoError(t, err)

	svc := NewChunkDecoder()
	blocks, err := svc.Run(data, 0, 0)
	require.NoError(t, err)
	require.NotNil(t, blocks)
	require.Len(t, blocks, 4096)
	require.Equal(t, "minecraft:stone", blocks[0].ID)
	require.Equal(t, 64, blocks[0].Y)
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
