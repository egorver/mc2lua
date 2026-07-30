package model

type Color [3]uint8

var DefaultColor = Color{191, 191, 191}

const DefaultMaterial = "Plastic"

type StyledElement struct {
	From     Vector3
	To       Vector3
	Rotation *ElementRotation
	Shade    bool
	Color    Color
	Material string
}

type StyledBlock struct {
	ID          string
	PropsKey    string
	IsFullBlock bool
	Elements    []StyledElement
}
