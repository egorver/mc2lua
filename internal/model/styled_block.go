package model

type Color [3]uint8

var DefaultColor = Color{191, 191, 191}

const (
	DefaultMaterial     = "Plastic"
	TexturelessMaterial = "SmoothPlastic"
)

type GridAlignment int

const (
	GridNotAligned GridAlignment = iota
	GridFullBlock
	GridSubBlock
)

type StyledElement struct {
	From     Vector3
	To       Vector3
	Rotation *ElementRotation
	Shade    bool
	Color    Color
	Material string
}

type StyledBlock struct {
	ID            string
	PropsKey      string
	GridAlignment GridAlignment
	RotX          float64
	RotY          float64
	Elements      []StyledElement
	Transparent   bool
}
