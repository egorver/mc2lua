package matcher

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

const defaultBrightness = 1.0

type BrightnessMatcher struct {
	factors map[string]float64
}

func NewBrightnessMatcher(fs fileReader, path string) (*BrightnessMatcher, error) {
	data, err := fs.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read brightness config: %w", err)
	}

	var raw struct {
		Brightness map[string]float64 `yaml:"brightness"`
	}

	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse brightness config: %w", err)
	}

	return &BrightnessMatcher{factors: raw.Brightness}, nil
}

func (svc *BrightnessMatcher) Run(material string) float64 {
	if f, ok := svc.factors[material]; ok {
		return f
	}
	return defaultBrightness
}
