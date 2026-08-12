package matcher

import (
	"fmt"

	"mc2lua/internal/model"

	"gopkg.in/yaml.v3"
)

type TintMatcher struct {
	tints map[string]model.TintType
}

func NewTintMatcher(fs fileReader, path string) (*TintMatcher, error) {
	data, err := fs.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read tints config: %w", err)
	}

	var raw struct {
		Tints map[string]model.TintType `yaml:"tints"`
	}

	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse tints config: %w", err)
	}

	svc := &TintMatcher{tints: raw.Tints}
	for blockID, tintType := range raw.Tints {
		if !svc.isValidTintType(tintType) {
			return nil, fmt.Errorf("tints config: unknown tint type %q for block %s", tintType, blockID)
		}
	}

	return svc, nil
}

func (svc *TintMatcher) Run(blockID string) (model.TintType, bool) {
	tintType, ok := svc.tints[blockID]
	return tintType, ok
}

func (svc *TintMatcher) isValidTintType(t model.TintType) bool {
	switch t {
	case model.TintGrass, model.TintFoliage, model.TintWater, model.TintRedstone:
		return true
	}
	return false
}
