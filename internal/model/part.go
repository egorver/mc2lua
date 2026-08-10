package model

type Part struct {
	Name     string
	Group    string
	GroupID  int
	BlockID  string
	PropsKey string
	Size     Vector3
	Position Vector3
	Color    Color
	Material string
}
