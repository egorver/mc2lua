package pipeline

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

type MaterialMatcher struct {
	mappings   map[string]string
	sortedKeys []string
	suffixes   []string
	overrides  map[string]string
}

func NewMaterialMatcher(fs fsApi, path string) (*MaterialMatcher, error) {
	data, err := fs.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read materials config: %w", err)
	}

	var raw struct {
		Mappings  map[string]string `yaml:"mappings"`
		Suffixes  []string          `yaml:"suffixes"`
		Overrides map[string]string `yaml:"overrides"`
	}

	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse materials config: %w", err)
	}

	m := &MaterialMatcher{
		mappings:  raw.Mappings,
		suffixes:  raw.Suffixes,
		overrides: raw.Overrides,
	}
	m.sortedKeys = sortedKeys(m.mappings)

	return m, nil
}

func (svc *MaterialMatcher) Run(blockID string) string {
	if m, ok := svc.overrides[blockID]; ok {
		return m
	}

	block := strings.TrimPrefix(blockID, "minecraft:")

	if m, ok := matchKeywords(block, svc.sortedKeys, svc.mappings); ok {
		return m
	}

	if base, ok := svc.findSuffix(block); ok {
		if m, ok := matchKeywords(base, svc.sortedKeys, svc.mappings); ok {
			return m
		}
		if strings.HasSuffix(base, "_planks") {
			return "Wood"
		}
	}

	return "SmoothPlastic"
}

func (svc *MaterialMatcher) findSuffix(block string) (string, bool) {
	for _, suffix := range svc.suffixes {
		if strings.HasSuffix(block, suffix) {
			return strings.TrimSuffix(block, suffix), true
		}
	}
	return "", false
}
