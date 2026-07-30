package pipeline

import (
	"mc2lua/internal/index"
	"mc2lua/internal/model"
)

type gridAnalyzer interface {
	Run(elements []model.StyledElement) model.GridAlignment
}

type elementRotator interface {
	Run(elements []model.StyledElement, rotX, rotY float64) []model.StyledElement
}

type materialMatcher interface {
	Run(blockID string) string
}

type StyleIndexer struct {
	gridAnalyzer    gridAnalyzer
	elementRotator  elementRotator
	materialMatcher materialMatcher
}

func NewStyleIndexer(ga gridAnalyzer, er elementRotator, mm materialMatcher) *StyleIndexer {
	return &StyleIndexer{gridAnalyzer: ga, elementRotator: er, materialMatcher: mm}
}

func (svc *StyleIndexer) Run(blocks []model.ResolvedBlock) *index.StyleIndex {
	idx := index.NewStyleIndex()

	for _, b := range blocks {
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
		gridAlignment := svc.gridAnalyzer.Run(rotated)

		styled := model.StyledBlock{
			ID:            b.ID,
			PropsKey:      b.PropsKey,
			GridAlignment: gridAlignment,
			Elements:      rotated,
		}

		idx.Add(b.ID, b.PropsKey, styled)
	}

	return idx
}
