package stateful

type OccupancyIndex struct {
	cells map[[3]int]bool
}

func NewOccupancyIndex() *OccupancyIndex {
	return &OccupancyIndex{cells: make(map[[3]int]bool)}
}

func (o *OccupancyIndex) FillCell(x, y, z int, occluding bool) {
	o.cells[[3]int{x, y, z}] = occluding
}

func (o *OccupancyIndex) FillRegion(x, y, z, w, h, d int, occluding bool) {
	for dx := 0; dx < w; dx++ {
		for dy := 0; dy < h; dy++ {
			for dz := 0; dz < d; dz++ {
				o.cells[[3]int{x + dx, y + dy, z + dz}] = occluding
			}
		}
	}
}

func (o *OccupancyIndex) Occupied(x, y, z int) bool {
	_, ok := o.cells[[3]int{x, y, z}]
	return ok
}

func (o *OccupancyIndex) Occluding(x, y, z int) bool {
	return o.cells[[3]int{x, y, z}]
}

func (o *OccupancyIndex) Len() int {
	return len(o.cells)
}
