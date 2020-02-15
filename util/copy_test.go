package util

import (
	"github.com/ihaiker/aginx/logs"
	"github.com/sirupsen/logrus"
	"os"
	"testing"
)

func init() {
	logs.SetLevel(logrus.DebugLevel)
}

func TestCopyDir(t *testing.T) {
	_ = os.RemoveAll("/tmp/nginx")

	err := CopyDir("/usr/local/etc/nginx", "/tmp/nginx")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("OVER")
}
