package matcher

import (
	"fmt"

	"mc2lua/internal/model"

	"gopkg.in/yaml.v3"
)

type ColorMatcher struct {
	colors map[string]model.Color
}

func NewColorMatcher(fs fileReader, path string) (*ColorMatcher, error) {
	data, err := fs.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read colors config: %w", err)
	}

	var raw struct {
		Colors map[string][3]uint8 `yaml:"colors"`
	}

	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse colors config: %w", err)
	}

	colors := make(map[string]model.Color, len(raw.Colors))
	for id, c := range raw.Colors {
		colors[id] = model.Color{c[0], c[1], c[2]}
	}

	return &ColorMatcher{colors: colors}, nil
}

func (svc *ColorMatcher) Run(blockID string) (model.Color, bool) {
	c, ok := svc.colors[blockID]
	return c, ok
}
