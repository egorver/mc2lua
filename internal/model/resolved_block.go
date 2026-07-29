package model

type Vector3 [3]float64

type ElementRotation struct {
	Origin  Vector3 `json:"origin"`
	Axis    string  `json:"axis"`
	Angle   float64 `json:"angle"`
	Rescale bool    `json:"rescale"`
}

type ElementFace struct {
	UV      [4]float64 `json:"uv"`
	Texture string     `json:"texture"`
}

type ModelElement struct {
	From     Vector3                `json:"from"`
	To       Vector3                `json:"to"`
	Rotation *ElementRotation       `json:"rotation,omitempty"`
	Shade    bool                   `json:"shade"`
	Faces    map[string]ElementFace `json:"faces,omitempty"`
}

type ResolvedBlock struct {
	IsFullBlock bool
	Elements    []ModelElement
	Textures    map[string]string
}
