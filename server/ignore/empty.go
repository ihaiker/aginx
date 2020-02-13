package ignore

type emptyIgnore struct{}

func Empty() *emptyIgnore {
	return &emptyIgnore{}
}

func (ignore *emptyIgnore) Add(files ...string) {}

func (ignore *emptyIgnore) Is(path string) bool {
	return false
}

func (ignore *emptyIgnore) IfNotIsAdd(path string) bool {
	return false
}
