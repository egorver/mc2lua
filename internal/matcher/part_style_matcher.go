package matcher

import (
	"fmt"

	"mc2lua/internal/model"

	"gopkg.in/yaml.v3"
)

type PartStyleMatcher struct {
	styles map[string]model.PartStyle
}

func NewPartStyleMatcher(fs fileReader, path string) (*PartStyleMatcher, error) {
	data, err := fs.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read parts config: %w", err)
	}

	var raw struct {
		Parts map[string]model.PartStyle `yaml:"parts"`
	}

	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse parts config: %w", err)
	}

	return &PartStyleMatcher{styles: raw.Parts}, nil
}

func (svc *PartStyleMatcher) Run(blockID string) (model.PartStyle, bool) {
	s, ok := svc.styles[blockID]
	return s, ok
}
