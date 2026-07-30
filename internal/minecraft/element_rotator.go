package minecraft

import (
	"math"

	"mc2lua/internal/model"
)

type ElementRotator struct{}

func NewElementRotator() *ElementRotator {
	return &ElementRotator{}
}

func (svc *ElementRotator) Run(elements []model.StyledElement, rotX, rotY float64) []model.StyledElement {
	if rotX == 0 && rotY == 0 {
		return elements
	}
	result := make([]model.StyledElement, len(elements))
	copy(result, elements)
	for i, elem := range result {
		from := svc.rotatePoint(elem.From, rotX, rotY)
		to := svc.rotatePoint(elem.To, rotX, rotY)
		for a := 0; a < 3; a++ {
			if from[a] > to[a] {
				from[a], to[a] = to[a], from[a]
			}
		}
		result[i].From, result[i].To = from, to
		if elem.Rotation != nil {
			newOrigin := svc.rotatePoint(elem.Rotation.Origin, rotX, rotY)
			result[i].Rotation = &model.ElementRotation{
				Origin:  newOrigin,
				Axis:    elem.Rotation.Axis,
				Angle:   elem.Rotation.Angle,
				Rescale: elem.Rotation.Rescale,
			}
		}
	}
	return result
}

func (svc *ElementRotator) rotatePoint(p model.Vector3, rotX, rotY float64) model.Vector3 {
	x, y, z := p[0], p[1], p[2]
	if rotY != 0 {
		rad := rotY * math.Pi / 180
		c, s := math.Cos(rad), math.Sin(rad)
		dx, dz := x-8, z-8
		x, z = 8+dx*c-dz*s, 8+dx*s+dz*c
	}
	if rotX != 0 {
		rad := rotX * math.Pi / 180
		c, s := math.Cos(rad), math.Sin(rad)
		dy, dz := y-8, z-8
		y, z = 8+dy*c-dz*s, 8+dy*s+dz*c
	}
	return model.Vector3{svc.roundCoord(x), svc.roundCoord(y), svc.roundCoord(z)}
}

func (svc *ElementRotator) roundCoord(v float64) float64 {
	return math.Round(v*64) / 64
}
