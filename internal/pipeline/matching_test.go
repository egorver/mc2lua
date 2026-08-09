package pipeline

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSortedKeys(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		m    map[string]int
		want []string
	}{
		{
			name: "nil map",
			m:    nil,
			want: []string{},
		},
		{
			name: "empty map",
			m:    map[string]int{},
			want: []string{},
		},
		{
			name: "single key",
			m:    map[string]int{"a": 1},
			want: []string{"a"},
		},
		{
			name: "different lengths longer first",
			m:    map[string]int{"a": 1, "bb": 2, "ccc": 3},
			want: []string{"ccc", "bb", "a"},
		},
		{
			name: "same length alphabetical",
			m:    map[string]int{"z": 1, "a": 2, "m": 3},
			want: []string{"a", "m", "z"},
		},
		{
			name: "mixed lengths with tie break",
			m:    map[string]int{"ab": 1, "aa": 2, "b": 3},
			want: []string{"aa", "ab", "b"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := sortedKeys(tt.m)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestSortedKeys_StringToString(t *testing.T) {
	t.Parallel()

	m := map[string]string{"z": "last", "m": "middle", "a": "first"}
	got := sortedKeys(m)
	require.Equal(t, []string{"a", "m", "z"}, got)
}

func TestMatchPrefixes(t *testing.T) {
	t.Parallel()

	dict := map[string]int{
		"apple":       1,
		"application": 2,
		"banana":      3,
	}
	keys := sortedKeys(dict)

	tests := []struct {
		name string
		s    string
		want int
		ok   bool
	}{
		{name: "exact match", s: "apple", want: 1, ok: true},
		{name: "prefix match", s: "application_suffix", want: 2, ok: true},
		{name: "longer prefix matched first", s: "application_abc", want: 2, ok: true},
		{name: "no match", s: "grape", want: 0, ok: false},
		{name: "empty string", s: "", want: 0, ok: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, ok := matchPrefixes(tt.s, keys, dict)
			require.Equal(t, tt.ok, ok)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestMatchPrefixes_EmptyDict(t *testing.T) {
	t.Parallel()

	dict := map[string]int{}
	keys := sortedKeys(dict)
	got, ok := matchPrefixes("anything", keys, dict)
	require.False(t, ok)
	require.Equal(t, 0, got)
}

func TestMatchKeywords(t *testing.T) {
	t.Parallel()

	dict := map[string]int{
		"stone": 1,
		"wood":  2,
		"glass": 3,
	}
	keys := sortedKeys(dict)

	tests := []struct {
		name string
		s    string
		want int
		ok   bool
	}{
		{name: "contains keyword", s: "minecraft:stone_stairs", want: 1, ok: true},
		{name: "exact match (also contains)", s: "stone", want: 1, ok: true},
		{name: "no match", s: "minecraft:dirt", want: 0, ok: false},
		{name: "empty string", s: "", want: 0, ok: false},
		{name: "multiple keywords picks first sorted", s: "wood_stone", want: 1, ok: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, ok := matchKeywords(tt.s, keys, dict)
			require.Equal(t, tt.ok, ok)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestMatchKeywords_EmptyDict(t *testing.T) {
	t.Parallel()

	dict := map[string]int{}
	keys := sortedKeys(dict)
	got, ok := matchKeywords("anything", keys, dict)
	require.False(t, ok)
	require.Equal(t, 0, got)
}
