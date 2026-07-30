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

type materialMatcher interface {
	Run(blockID string) string
}

type StyleIndexer struct {
	modelAnalyzer   modelAnalyzer
	elementRotator  elementRotator
	materialMatcher materialMatcher
}

func NewStyleIndexer(ma modelAnalyzer, er elementRotator, mm materialMatcher) *StyleIndexer {
	return &StyleIndexer{modelAnalyzer: ma, elementRotator: er, materialMatcher: mm}
}

func (svc *StyleIndexer) Run(blocks []model.ResolvedBlock) *index.StyleIndex {
	idx := index.NewStyleIndex()

	for _, b := range blocks {
		isFullBlock := svc.modelAnalyzer.Run(b.Elements)
		material := svc.materialMatcher.Run(b.ID)

		elements := make([]model.StyledElement, len(b.Elements))
		for i, el := range b.Elements {
			elements[i] = model.StyledElement{
				From:     el.From,
				To:       el.To,
				Rotation: el.Rotation,
				Shade:    el.Shade,
				Color:    model.DefaultColor,
				Material: material,
			}
		}

		rotated := svc.elementRotator.Run(elements, b.RotX, b.RotY)

		styled := model.StyledBlock{
			ID:          b.ID,
			PropsKey:    b.PropsKey,
			IsFullBlock: isFullBlock,
			Elements:    rotated,
		}

		idx.Add(b.ID, b.PropsKey, styled)
	}

	return idx
}
