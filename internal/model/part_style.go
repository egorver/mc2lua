package model

const (
	FaceTop    = "top"
	FaceBottom = "bottom"
	FaceFront  = "front"
	FaceBack   = "back"
	FaceLeft   = "left"
	FaceRight  = "right"
	FaceSides  = "sides"
	FaceFaces  = "faces"
	FaceWalls  = "walls"
	FaceAll    = "all"
)

type Surface struct {
	Texture      string   `yaml:"texture,omitempty"`
	Color        *Color   `yaml:"color,omitempty"`
	Transparency *float64 `yaml:"transparency,omitempty"`
	StudsPerTile *float64 `yaml:"studs_per_tile,omitempty"`
}

type PartStyle struct {
	Color        *Color   `yaml:"color,omitempty"`
	Transparency *float64 `yaml:"transparency,omitempty"`
	Top          *Surface `yaml:"top,omitempty"`
	Bottom       *Surface `yaml:"bottom,omitempty"`
	Front        *Surface `yaml:"front,omitempty"`
	Back         *Surface `yaml:"back,omitempty"`
	Left         *Surface `yaml:"left,omitempty"`
	Right        *Surface `yaml:"right,omitempty"`
	Sides        *Surface `yaml:"sides,omitempty"`
	Faces        *Surface `yaml:"faces,omitempty"`
	Walls        *Surface `yaml:"walls,omitempty"`
	All          *Surface `yaml:"all,omitempty"`
}
