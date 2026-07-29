package minecraft

import "mc2lua/internal/model"

type ModelAnalyzer struct{}

func NewModelAnalyzer() *ModelAnalyzer {
	return &ModelAnalyzer{}
}

func (svc *ModelAnalyzer) Run(elements []model.ModelElement) bool {
	if len(elements) != 1 {
		return false
	}

	elem := elements[0]

	if !svc.isZero(elem.From[:], 0, 0, 0) {
		return false
	}
	if !svc.isZero(elem.To[:], 16, 16, 16) {
		return false
	}
	if elem.Rotation != nil {
		return false
	}
	if !elem.Shade {
		return false
	}

	return true
}

func (svc *ModelAnalyzer) isZero(v []float64, x, y, z float64) bool {
	return v[0] == x && v[1] == y && v[2] == z
}
