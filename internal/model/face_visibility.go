package model

type FaceMask [6]bool

const (
	FaceIndexTop = iota
	FaceIndexBottom
	FaceIndexFront
	FaceIndexBack
	FaceIndexLeft
	FaceIndexRight
)

type FaceVisibility struct {
	BlockFaces   []FaceMask
	MicroFaces   []FaceMask
	ComplexFaces []FaceMask
}
