package model

type Part struct {
	Name         string
	Group        string
	GroupID      int
	BlockID      string
	PropsKey     string
	Size         Vector3
	Position     Vector3
	Color        Color
	Material     string
	VisibleFaces FaceMask
	Transparency *float64
	Top          *Surface
	Bottom       *Surface
	Front        *Surface
	Back         *Surface
	Left         *Surface
	Right        *Surface
}
