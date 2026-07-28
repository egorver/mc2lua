package app

type AppConfig struct {
	Input    string
	Output   string
	Scale    int
	NoOffset bool
	XMin, XMax int
	YMin, YMax int
	ZMin, ZMax int
}
