package pipeline

import (
	"sort"
	"strings"
)

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if len(keys[i]) != len(keys[j]) {
			return len(keys[i]) > len(keys[j])
		}
		return keys[i] < keys[j]
	})
	return keys
}

func matchPrefixes[V any](s string, sortedKeys []string, dict map[string]V) (V, bool) {
	for _, key := range sortedKeys {
		if strings.HasPrefix(s, key) {
			v, ok := dict[key]
			return v, ok
		}
	}
	var zero V
	return zero, false
}

func matchKeywords[V any](s string, sortedKeys []string, dict map[string]V) (V, bool) {
	for _, key := range sortedKeys {
		if strings.Contains(s, key) {
			v, ok := dict[key]
			return v, ok
		}
	}
	var zero V
	return zero, false
}
