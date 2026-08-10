package matcher

type fileReader interface {
	ReadFile(name string) ([]byte, error)
}
