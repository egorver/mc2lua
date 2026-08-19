package minecraft

import (
	"math"

	"mc2lua/internal/model"
)

const rotationPrecision = 64.0

type ElementRotator struct{}

func NewElementRotator() *ElementRotator {
	return &ElementRotator{}
}

func (svc *ElementRotator) RunStyled(elements []model.StyledElement, rotX, rotY float64) []model.StyledElement {
	if rotX == 0 && rotY == 0 {
		return elements
	}
	result := make([]model.StyledElement, len(elements))
	copy(result, elements)
	for i, elem := range result {
		from, to, rotation := svc.rotateElement(elem.From, elem.To, elem.Rotation, rotX, rotY)
		result[i].From, result[i].To, result[i].Rotation = from, to, rotation
	}
	return result
}

func (svc *ElementRotator) RunModel(elements []model.ModelElement, rotX, rotY float64) []model.ModelElement {
	if rotX == 0 && rotY == 0 {
		return elements
	}
	result := make([]model.ModelElement, len(elements))
	copy(result, elements)
	for i, elem := range result {
		from, to, rotation := svc.rotateElement(elem.From, elem.To, elem.Rotation, rotX, rotY)
		result[i].From, result[i].To, result[i].Rotation = from, to, rotation
	}
	return result
}

func (svc *ElementRotator) rotateElement(from, to model.Vector3, rotation *model.ElementRotation, rotX, rotY float64) (model.Vector3, model.Vector3, *model.ElementRotation) {
	from, to = svc.rotateBounds(from, to, rotX, rotY)
	if rotation == nil {
		return from, to, nil
	}
	return from, to, &model.ElementRotation{
		Origin:  svc.rotatePoint(rotation.Origin, rotX, rotY),
		Axis:    rotation.Axis,
		Angle:   rotation.Angle,
		Rescale: rotation.Rescale,
	}
}

func (svc *ElementRotator) rotateBounds(from, to model.Vector3, rotX, rotY float64) (model.Vector3, model.Vector3) {
	from = svc.rotatePoint(from, rotX, rotY)
	to = svc.rotatePoint(to, rotX, rotY)
	for a := 0; a < model.BlockDimensions; a++ {
		if from[a] > to[a] {
			from[a], to[a] = to[a], from[a]
		}
	}
	return from, to
}

func (svc *ElementRotator) rotatePoint(p model.Vector3, rotX, rotY float64) model.Vector3 {
	x, y, z := p[0], p[1], p[2]
	if rotX != 0 {
		rad := rotX * model.DegreesToRadians
		c, s := math.Cos(rad), math.Sin(rad)
		dy, dz := y-model.BlockCenter, z-model.BlockCenter
		y, z = model.BlockCenter+dy*c-dz*s, model.BlockCenter+dy*s+dz*c
	}
	if rotY != 0 {
		rad := rotY * model.DegreesToRadians
		c, s := math.Cos(rad), math.Sin(rad)
		dx, dz := x-model.BlockCenter, z-model.BlockCenter
		x, z = model.BlockCenter+dx*c-dz*s, model.BlockCenter+dx*s+dz*c
	}
	return model.Vector3{svc.roundCoord(x), svc.roundCoord(y), svc.roundCoord(z)}
}

func (svc *ElementRotator) roundCoord(v float64) float64 {
	return math.Round(v*rotationPrecision) / rotationPrecision
}
