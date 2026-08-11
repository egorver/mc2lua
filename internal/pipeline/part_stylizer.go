package pipeline

import (
	"math"

	"mc2lua/internal/model"
	"mc2lua/internal/stateful"
)

type partStyleMatcher interface {
	Run(blockID string) (model.PartStyle, bool)
}

var namedFaces = []string{
	model.FaceTop, model.FaceBottom,
	model.FaceFront, model.FaceBack,
	model.FaceLeft, model.FaceRight,
}

type PartStylizer struct {
	partStyleMatcher partStyleMatcher
}

func NewPartStylizer(psm partStyleMatcher) *PartStylizer {
	return &PartStylizer{partStyleMatcher: psm}
}

func (svc *PartStylizer) Run(parts []model.Part, styleIndex stateful.StyleIndex) []model.Part {
	styled := make([]model.Part, len(parts))
	for i, p := range parts {
		partStyle, ok := svc.partStyleMatcher.Run(p.BlockID)
		if !ok {
			styled[i] = p
			continue
		}
		rotX, rotY := svc.blockRotation(styleIndex, p)
		styled[i] = svc.applyPartStyle(p, partStyle, rotX, rotY)
	}
	return styled
}

func (svc *PartStylizer) blockRotation(styleIndex stateful.StyleIndex, p model.Part) (float64, float64) {
	styled, ok := styleIndex.Get(p.BlockID, p.PropsKey)
	if !ok {
		return 0, 0
	}
	return styled.RotX, styled.RotY
}

func (svc *PartStylizer) applyPartStyle(p model.Part, style model.PartStyle, rotX, rotY float64) model.Part {
	if style.Transparency != nil {
		p.Transparency = style.Transparency
	}

	hasTexture := false
	for _, face := range namedFaces {
		surf := svc.resolveSurface(style, face)
		if surf == nil {
			continue
		}
		worldFace := svc.mapFace(face, rotX, rotY)
		if !p.VisibleFaces[worldFace] {
			continue
		}
		svc.setSurface(&p, worldFace, surf)
		if surf.Texture != "" {
			hasTexture = true
		}
	}

	if hasTexture {
		p.Material = model.TexturelessMaterial
	}
	return p
}

func (svc *PartStylizer) resolveSurface(style model.PartStyle, face string) *model.Surface {
	var merged model.Surface
	have := false

	for _, name := range svc.surfaceSources(face) {
		src := svc.surfaceByKey(style, name)
		if src == nil {
			continue
		}
		if src.Texture != "" {
			merged.Texture = src.Texture
			have = true
		}
		if src.Color != nil {
			merged.Color = src.Color
			have = true
		}
		if src.Transparency != nil {
			merged.Transparency = src.Transparency
			have = true
		}
	}

	if !have {
		return nil
	}
	return &merged
}

func (svc *PartStylizer) surfaceSources(face string) []string {
	sources := []string{model.FaceAll, model.FaceFaces}
	switch face {
	case model.FaceFront, model.FaceBack, model.FaceLeft, model.FaceRight:
		sources = append(sources, model.FaceSides, model.FaceWalls)
	}
	return append(sources, face)
}

func (svc *PartStylizer) surfaceByKey(style model.PartStyle, name string) *model.Surface {
	switch name {
	case model.FaceTop:
		return style.Top
	case model.FaceBottom:
		return style.Bottom
	case model.FaceFront:
		return style.Front
	case model.FaceBack:
		return style.Back
	case model.FaceLeft:
		return style.Left
	case model.FaceRight:
		return style.Right
	case model.FaceSides:
		return style.Sides
	case model.FaceFaces:
		return style.Faces
	case model.FaceWalls:
		return style.Walls
	case model.FaceAll:
		return style.All
	}
	return nil
}

func (svc *PartStylizer) setSurface(p *model.Part, worldFace int, surf *model.Surface) {
	switch worldFace {
	case model.FaceIndexTop:
		p.Top = surf
	case model.FaceIndexBottom:
		p.Bottom = surf
	case model.FaceIndexFront:
		p.Front = surf
	case model.FaceIndexBack:
		p.Back = surf
	case model.FaceIndexLeft:
		p.Left = surf
	case model.FaceIndexRight:
		p.Right = surf
	}
}

func (svc *PartStylizer) mapFace(face string, rotX, rotY float64) int {
	dir := baseFaceDir(face)
	if rotY != 0 {
		rad := rotY * math.Pi / 180
		c, s := math.Cos(rad), math.Sin(rad)
		x, z := dir[0], dir[2]
		dir[0] = x*c - z*s
		dir[2] = x*s + z*c
	}
	if rotX != 0 {
		rad := rotX * math.Pi / 180
		c, s := math.Cos(rad), math.Sin(rad)
		y, z := dir[1], dir[2]
		dir[1] = y*c - z*s
		dir[2] = y*s + z*c
	}
	return directionToFace(dir)
}

func baseFaceDir(face string) model.Vector3 {
	switch face {
	case model.FaceTop:
		return model.Vector3{0, 1, 0}
	case model.FaceBottom:
		return model.Vector3{0, -1, 0}
	case model.FaceFront:
		return model.Vector3{0, 0, -1}
	case model.FaceBack:
		return model.Vector3{0, 0, 1}
	case model.FaceLeft:
		return model.Vector3{-1, 0, 0}
	case model.FaceRight:
		return model.Vector3{1, 0, 0}
	}
	return model.Vector3{}
}

func directionToFace(dir model.Vector3) int {
	x := roundAxis(dir[0])
	y := roundAxis(dir[1])
	z := roundAxis(dir[2])
	switch {
	case y == 1:
		return model.FaceIndexTop
	case y == -1:
		return model.FaceIndexBottom
	case z == 1:
		return model.FaceIndexFront
	case z == -1:
		return model.FaceIndexBack
	case x == -1:
		return model.FaceIndexLeft
	case x == 1:
		return model.FaceIndexRight
	}
	return model.FaceIndexTop
}

func roundAxis(v float64) int {
	switch {
	case v > 0.5:
		return 1
	case v < -0.5:
		return -1
	}
	return 0
}
