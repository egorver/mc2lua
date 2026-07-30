package pipeline

import (
	"mc2lua/internal/index"
	"mc2lua/internal/model"
)

type modelAnalyzer interface {
	Run(elements []model.ModelElement) bool
}

type elementRotator interface {
	Run(elements []model.StyledElement, rotX, rotY float64) []model.StyledElement
}

type StyleIndexer struct {
	modelAnalyzer  modelAnalyzer
	elementRotator elementRotator
}

func NewStyleIndexer(ma modelAnalyzer, er elementRotator) *StyleIndexer {
	return &StyleIndexer{modelAnalyzer: ma, elementRotator: er}
}

func (svc *StyleIndexer) Run(blocks []model.ResolvedBlock) *index.StyleIndex {
	idx := index.NewStyleIndex()

	for _, b := range blocks {
		isFullBlock := svc.modelAnalyzer.Run(b.Elements)

		elements := make([]model.StyledElement, len(b.Elements))
		for i, el := range b.Elements {
			elements[i] = model.StyledElement{
				From:     el.From,
				To:       el.To,
				Rotation: el.Rotation,
				Shade:    el.Shade,
				Color:    model.DefaultColor,
				Material: model.DefaultMaterial,
			}
		}

		elements = svc.elementRotator.Run(elements, b.RotX, b.RotY)

		styled := model.StyledBlock{
			ID:          b.ID,
			PropsKey:    b.PropsKey,
			IsFullBlock: isFullBlock,
			Elements:    elements,
		}

		idx.Add(b.ID, b.PropsKey, styled)
	}

	return idx
}
