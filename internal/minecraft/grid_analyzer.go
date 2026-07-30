package minecraft

import "mc2lua/internal/model"

type GridAnalyzer struct{}

func NewGridAnalyzer() *GridAnalyzer {
	return &GridAnalyzer{}
}

func (svc *GridAnalyzer) Run(elements []model.StyledElement) model.GridAlignment {
	if len(elements) == 0 {
		return model.GridNotAligned
	}

	for _, elem := range elements {
		if elem.Rotation != nil || !elem.Shade {
			return model.GridNotAligned
		}
		if !svc.isGridAligned(elem.From[:]) || !svc.isGridAligned(elem.To[:]) {
			return model.GridNotAligned
		}
	}

	if len(elements) == 1 {
		elem := elements[0]
		if svc.isZero(elem.From[:], 0, 0, 0) && svc.isZero(elem.To[:], 16, 16, 16) {
			return model.GridFullBlock
		}
	}

	return model.GridSubBlock
}

func (svc *GridAnalyzer) isGridAligned(v []float64) bool {
	return int(v[0])%4 == 0 && int(v[1])%4 == 0 && int(v[2])%4 == 0
}

func (svc *GridAnalyzer) isZero(v []float64, x, y, z float64) bool {
	return v[0] == x && v[1] == y && v[2] == z
}
