package pipeline

import (
	"mc2lua/internal/model"
)

type indexerModelAnalyzer interface {
	Run(elements []model.ModelElement) bool
}

type StyleIndexer struct {
	modelAnalyzer indexerModelAnalyzer
}

func NewStyleIndexer(ma indexerModelAnalyzer) *StyleIndexer {
	return &StyleIndexer{modelAnalyzer: ma}
}

func (svc *StyleIndexer) Run(blocks []model.ResolvedBlock) *model.StyleIndex {
	idx := model.NewStyleIndex()

	for _, b := range blocks {
		isFullBlock := svc.modelAnalyzer.Run(b.Elements)

		elements := make([]model.StyledElement, len(b.Elements))
		for i, el := range b.Elements {
			elements[i] = model.StyledElement{
				From:     el.From,
				To:       el.To,
				Rotation: el.Rotation,
				Shade:    el.Shade,
			}
		}

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
