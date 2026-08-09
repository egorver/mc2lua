package minecraft

import (
	"encoding/json"
	"mc2lua/internal/runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBlockstateParserRun(t *testing.T) {
	fs, nsToRoots := setupTestFS()
	addBlockstate(fs, "minecraft", "furnace", []byte(testBlockstateVariants))
	addBlockstate(fs, "minecraft", "grass", []byte(testBlockstateArray))
	addBlockstate(fs, "minecraft", "default", []byte(testBlockstateDefault))
	addBlockstate(fs, "minecraft", "multipart", []byte(testBlockstateMultipart))
	addBlockstate(fs, "minecraft", "empty_variants", []byte(testBlockstateEmptyVariants))
	addBlockstate(fs, "minecraft", "bad_json", []byte(testBlockstateInvalidJSON))
	addBlockstate(fs, "minecraft", "no_variants", []byte(testBlockstateNoVariants))

	svc := NewBlockstateParser(fs, NewPropsKeyBuilder())

	tests := []struct {
		name      string
		namespace string
		blockID   string
		props     map[string]string
		want      []blockstateMatch
		wantErr   string
	}{
		{
			name:      "exact variant match",
			namespace: "minecraft", blockID: "furnace",
			props: map[string]string{"facing": "north"},
			want:  []blockstateMatch{{Model: "block/furnace"}},
		},
		{
			name:      "variant with rotation",
			namespace: "minecraft", blockID: "furnace",
			props: map[string]string{"facing": "south"},
			want:  []blockstateMatch{{Model: "block/furnace", RotY: 180}},
		},
		{
			name:      "variant multiple properties",
			namespace: "minecraft", blockID: "furnace",
			props: map[string]string{"lit": "true", "facing": "east"},
			want:  []blockstateMatch{{Model: "block/furnace_on", RotY: 90}},
		},
		{
			name:      "fallback to empty variant",
			namespace: "minecraft", blockID: "default",
			props: map[string]string{"unused": "x"},
			want:  []blockstateMatch{{Model: "block/cube"}},
		},
		{
			name:      "array variant uses first only",
			namespace: "minecraft", blockID: "grass",
			props: map[string]string{},
			want:  []blockstateMatch{{Model: "block/grass"}},
		},
		{
			name:      "no matching variant",
			namespace: "minecraft", blockID: "furnace",
			props:   map[string]string{"facing": "up"},
			wantErr: "no matching variant",
		},
		{
			name:      "unknown namespace",
			namespace: "unknown", blockID: "stone",
			props:   map[string]string{},
			wantErr: "unknown namespace",
		},
		{
			name:      "file not found",
			namespace: "minecraft", blockID: "nonexistent",
			props:   map[string]string{},
			wantErr: "not found in any mod directory",
		},
		{
			name:      "invalid JSON",
			namespace: "minecraft", blockID: "bad_json",
			props:   map[string]string{},
			wantErr: "parse blockstate",
		},
		{
			name:      "multipart without when",
			namespace: "minecraft", blockID: "multipart",
			props: map[string]string{},
			want:  []blockstateMatch{{Model: "block/block"}},
		},
		{
			name:      "no variants field",
			namespace: "minecraft", blockID: "no_variants",
			props:   map[string]string{},
			wantErr: "no variants or multipart data",
		},
		{
			name:      "empty variants object",
			namespace: "minecraft", blockID: "empty_variants",
			props:   map[string]string{},
			wantErr: "no variants or multipart data",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := svc.Run(tt.namespace, tt.blockID, tt.props, nsToRoots)
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestBlockstateParserRunMultipleRoots(t *testing.T) {
	fs := runtime.NewFSMock()
	fs.AddDir("assets", 0755)
	fs.AddDir("assets/mod1", 0755)
	fs.AddDir("assets/mod1/minecraft", 0755)
	fs.AddDir("assets/mod1/minecraft/blockstates", 0755)
	fs.AddDir("assets/mod2", 0755)
	fs.AddDir("assets/mod2/minecraft", 0755)
	fs.AddDir("assets/mod2/minecraft/blockstates", 0755)
	addBlockstate(fs, "mod2/minecraft", "stone", []byte(testBlockstateDefault))

	nsToRoots := map[string][]string{
		"minecraft": {"assets/mod1/minecraft", "assets/mod2/minecraft"},
	}
	svc := NewBlockstateParser(fs, NewPropsKeyBuilder())

	matches, err := svc.Run("minecraft", "stone", nil, nsToRoots)
	require.NoError(t, err)
	require.Equal(t, []blockstateMatch{{Model: "block/cube"}}, matches)
}

func TestBlockstateParserParseVariantValue(t *testing.T) {
	svc := NewBlockstateParser(nil, NewPropsKeyBuilder())

	tests := []struct {
		name    string
		input   []byte
		want    []blockstateMatch
		wantErr string
	}{
		{name: "single object", input: []byte(`{"model":"block/cube"}`), want: []blockstateMatch{{Model: "block/cube"}}},
		{name: "object with rotation", input: []byte(`{"model":"block/furnace","x":90,"y":180}`), want: []blockstateMatch{{Model: "block/furnace", RotX: 90, RotY: 180}}},
		{name: "array of variants uses first only", input: []byte(`[{"model":"block/stone"},{"model":"block/andesite"}]`), want: []blockstateMatch{{Model: "block/stone"}}},
		{name: "empty value", input: []byte(``), wantErr: "empty variant value"},
		{name: "whitespace only", input: []byte(`   `), wantErr: "empty variant value"},
		{name: "invalid JSON object", input: []byte(`{broken`), wantErr: "parse variant object"},
		{name: "invalid JSON array", input: []byte(`[{broken`), wantErr: "parse variant array"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := svc.parseVariantValue(tt.input)
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}



func TestBlockstateParserMatchKey(t *testing.T) {
	svc := NewBlockstateParser(nil, NewPropsKeyBuilder())

	tests := []struct {
		name  string
		key   string
		props map[string]string
		want  bool
	}{
		{name: "exact match", key: "facing=north", props: map[string]string{"facing": "north"}, want: true},
		{name: "mismatch", key: "facing=north", props: map[string]string{"facing": "south"}, want: false},
		{name: "multiple props match", key: "facing=north,lit=true", props: map[string]string{"facing": "north", "lit": "true"}, want: true},
		{name: "multiple props one mismatch", key: "facing=north,lit=true", props: map[string]string{"facing": "north", "lit": "false"}, want: false},
		{name: "missing prop", key: "facing=north", props: map[string]string{}, want: false},
		{name: "empty key returns true", key: "", props: map[string]string{"anything": "x"}, want: true},
		{name: "empty key empty props", key: "", props: map[string]string{}, want: true},
		{name: "key without equals ignored", key: "invalid", props: map[string]string{}, want: true},
		{name: "extra props ignored", key: "facing=north", props: map[string]string{"facing": "north", "unused": "x"}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.matchKey(tt.key, tt.props)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestBlockstateParser_MatchVariant(t *testing.T) {
	t.Parallel()

	svc := NewBlockstateParser(nil, NewPropsKeyBuilder())

	tests := []struct {
		name     string
		variants map[string]json.RawMessage
		props    map[string]string
		want     []blockstateMatch
		wantErr  string
	}{
		{
			name: "exact match",
			variants: map[string]json.RawMessage{
				"facing=north": json.RawMessage(`{"model":"block/furnace"}`),
			},
			props: map[string]string{"facing": "north"},
			want:  []blockstateMatch{{Model: "block/furnace"}},
		},
		{
			name: "fallback to empty variant",
			variants: map[string]json.RawMessage{
				"facing=north": json.RawMessage(`{"model":"block/furnace"}`),
				"":             json.RawMessage(`{"model":"block/default"}`),
			},
			props: map[string]string{"facing": "south"},
			want:  []blockstateMatch{{Model: "block/default"}},
		},
		{
			name: "match by key iteration",
			variants: map[string]json.RawMessage{
				"facing=north,lit=true": json.RawMessage(`{"model":"block/furnace_on"}`),
				"facing=north":          json.RawMessage(`{"model":"block/furnace"}`),
			},
			props: map[string]string{"facing": "north", "lit": "false"},
			want:  []blockstateMatch{{Model: "block/furnace"}},
		},
		{
			name: "no matching variant",
			variants: map[string]json.RawMessage{
				"facing=north": json.RawMessage(`{"model":"block/furnace"}`),
			},
			props:   map[string]string{"facing": "up"},
			wantErr: "no matching variant",
		},
		{
			name:     "empty variants map",
			variants: map[string]json.RawMessage{},
			props:    map[string]string{},
			wantErr:  "no variants defined",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := svc.matchVariant(tt.variants, tt.props)
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestBlockstateParserReadBlockstateFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		ns         string
		blockID    string
		namespaces map[string][]string
		setup      func(fs *runtime.FSMock)
		wantErr    string
	}{
		{
			name:       "unknown namespace",
			ns:         "unknown",
			blockID:    "stone",
			namespaces: map[string][]string{"minecraft": {"assets/minecraft"}},
			wantErr:    "unknown namespace",
		},
		{
			name:       "file not found",
			ns:         "minecraft",
			blockID:    "nonexistent",
			namespaces: map[string][]string{"minecraft": {"assets/minecraft"}},
			setup:      func(fs *runtime.FSMock) {},
			wantErr:    "not found in any mod directory",
		},
		{
			name:       "invalid JSON",
			ns:         "minecraft",
			blockID:    "bad",
			namespaces: map[string][]string{"minecraft": {"assets/minecraft"}},
			setup: func(fs *runtime.FSMock) {
				addBlockstate(fs, "minecraft", "bad", []byte(`{bad json`))
			},
			wantErr: "parse blockstate",
		},
		{
			name:       "success",
			ns:         "minecraft",
			blockID:    "stone",
			namespaces: map[string][]string{"minecraft": {"assets/minecraft"}},
			setup: func(fs *runtime.FSMock) {
				addBlockstate(fs, "minecraft", "stone", []byte(testBlockstateDefault))
			},
		},
		{
			name:       "multiple roots second has file",
			ns:         "minecraft",
			blockID:    "stone",
			namespaces: map[string][]string{"minecraft": {"assets/mod1/minecraft", "assets/mod2/minecraft"}},
			setup: func(fs *runtime.FSMock) {
				addBlockstate(fs, "mod2/minecraft", "stone", []byte(testBlockstateDefault))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fs := runtime.NewFSMock()
			if tt.setup != nil {
				tt.setup(fs)
			}
			svc := NewBlockstateParser(fs, NewPropsKeyBuilder())

			_, _, err := svc.readBlockstateFile(tt.ns, tt.blockID, tt.namespaces)
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestBlockstateParserMatchMultipart(t *testing.T) {
	t.Parallel()

	svc := NewBlockstateParser(nil, NewPropsKeyBuilder())

	tests := []struct {
		name    string
		parts   []json.RawMessage
		props   map[string]string
		want    []blockstateMatch
		wantErr string
	}{
		{
			name: "no when matches all parts",
			parts: []json.RawMessage{
				json.RawMessage(`{"apply":{"model":"block/a"}}`),
				json.RawMessage(`{"apply":{"model":"block/b"}}`),
			},
			want: []blockstateMatch{{Model: "block/a"}, {Model: "block/b"}},
		},
		{
			name: "simple when match",
			parts: []json.RawMessage{
				json.RawMessage(`{"when":{"north":"true"},"apply":{"model":"block/side"}}`),
			},
			props: map[string]string{"north": "true"},
			want:  []blockstateMatch{{Model: "block/side"}},
		},
		{
			name: "simple when no match",
			parts: []json.RawMessage{
				json.RawMessage(`{"when":{"north":"true"},"apply":{"model":"block/side"}}`),
			},
			props:   map[string]string{"north": "false"},
			wantErr: "no multipart conditions match",
		},
		{
			name: "AND when all match",
			parts: []json.RawMessage{
				json.RawMessage(`{"when":{"AND":[{"age":"0"},{"rooted":"false"}]},"apply":{"model":"block/seed"}}`),
			},
			props: map[string]string{"age": "0", "rooted": "false"},
			want:  []blockstateMatch{{Model: "block/seed"}},
		},
		{
			name: "AND when partial match",
			parts: []json.RawMessage{
				json.RawMessage(`{"when":{"AND":[{"age":"0"},{"rooted":"false"}]},"apply":{"model":"block/seed"}}`),
			},
			props:   map[string]string{"age": "0", "rooted": "true"},
			wantErr: "no multipart conditions match",
		},
		{
			name: "mixed when and no-when parts",
			parts: []json.RawMessage{
				json.RawMessage(`{"apply":{"model":"block/post"}}`),
				json.RawMessage(`{"when":{"north":"true"},"apply":{"model":"block/side"}}`),
				json.RawMessage(`{"when":{"north":"false"},"apply":{"model":"block/noside"}}`),
			},
			props: map[string]string{"north": "true"},
			want:  []blockstateMatch{{Model: "block/post"}, {Model: "block/side"}},
		},
		{
			name: "with rotation",
			parts: []json.RawMessage{
				json.RawMessage(`{"when":{"east":"true"},"apply":{"model":"block/side","y":90}}`),
			},
			props: map[string]string{"east": "true"},
			want:  []blockstateMatch{{Model: "block/side", RotY: 90}},
		},
		{
			name:    "empty parts",
			parts:   []json.RawMessage{},
			wantErr: "no multipart conditions match",
		},
		{
			name:    "invalid part JSON",
			parts:   []json.RawMessage{json.RawMessage(`{bad}`)},
			wantErr: "parse multipart part",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := svc.matchMultipart(tt.parts, tt.props)
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestBlockstateParserMatchWhen(t *testing.T) {
	t.Parallel()

	svc := NewBlockstateParser(nil, NewPropsKeyBuilder())

	tests := []struct {
		name  string
		when  json.RawMessage
		props map[string]string
		want  bool
	}{
		{name: "empty when always matches", when: nil, props: nil, want: true},
		{name: "simple when matches", when: json.RawMessage(`{"north":"true"}`), props: map[string]string{"north": "true"}, want: true},
		{name: "simple when mismatches", when: json.RawMessage(`{"north":"true"}`), props: map[string]string{"north": "false"}, want: false},
		{name: "AND when all match", when: json.RawMessage(`{"AND":[{"age":"0"},{"rooted":"false"}]}`), props: map[string]string{"age": "0", "rooted": "false"}, want: true},
		{name: "AND when one mismatches", when: json.RawMessage(`{"AND":[{"age":"0"},{"rooted":"false"}]}`), props: map[string]string{"age": "0", "rooted": "true"}, want: false},
		{name: "extra props ignored in when", when: json.RawMessage(`{"north":"true"}`), props: map[string]string{"north": "true", "waterlogged": "false"}, want: true},
		{name: "array when returns false", when: json.RawMessage(`[]`), props: nil, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := svc.matchWhen(tt.when, tt.props)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestBlockstateParserSortedKeys(t *testing.T) {
	svc := NewBlockstateParser(nil, NewPropsKeyBuilder())

	tests := []struct {
		name string
		m    map[string]json.RawMessage
		want []string
	}{
		{name: "nil", m: nil, want: []string{}},
		{name: "empty", m: map[string]json.RawMessage{}, want: []string{}},
		{name: "single", m: map[string]json.RawMessage{"a": nil}, want: []string{"a"}},
		{name: "sorted", m: map[string]json.RawMessage{"z": nil, "a": nil, "m": nil}, want: []string{"a", "m", "z"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.sortedKeys(tt.m)
			require.Equal(t, tt.want, got)
		})
	}
}
