package model

type Colormap struct {
	Grass   Color
	Foliage Color
}

type TintType string

const (
	TintGrass    TintType = "grass"
	TintFoliage  TintType = "foliage"
	TintWater    TintType = "water"
	TintRedstone TintType = "redstone"
)
