package app

type AppConfig struct {
	Input         string
	AssetsDir     string
	Output        string
	Scale         int
	NoOffset      bool
	MaterialsPath string
	ColorsPath    string
	XMin, XMax    int
	YMin, YMax    int
	ZMin, ZMax    int
}
