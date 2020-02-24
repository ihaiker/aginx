package file

import (
	"github.com/ihaiker/aginx/util"
	"testing"
)

var engine = MustSystem()

func TestStorage(t *testing.T) {
	files, err := engine.List()
	util.PanicIfError(err)

	for _, file := range files {
		t.Log(file)
	}
}
