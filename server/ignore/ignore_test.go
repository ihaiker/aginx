package ignore

import (
	"strconv"
	"testing"
)

func TestIgnore(t *testing.T) {
	ignore := Empty()

	for i := 0; i < 10; i++ {
		ignore.Add(strconv.Itoa(i))
	}

	for i := 0; i < 5; i++ {
		ignore.Is(strconv.Itoa(i))
	}

	for i := 0; i < 10; i++ {
		if ignore.Is(strconv.Itoa(i)) {
			t.Log("has ", i)
		}
	}

	for i := 0; i < 10; i++ {
		if ignore.Is(strconv.Itoa(i)) {
			t.Log("has ", i)
		} else {
			t.Log("no has ", i)
		}
	}
}
