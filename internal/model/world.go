package model

type Block struct {
	ID         string
	Properties map[string]string
	X, Y, Z    int
}

type World struct {
	Blocks      []Block
	Lookup      map[[3]int]*Block
	MinX, MinY, MinZ int
	MaxX, MaxY, MaxZ int
}
