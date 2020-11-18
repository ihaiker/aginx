package logs_test

import (
	"bytes"
	"github.com/ihaiker/aginx/v2/core/logs"
	"github.com/sirupsen/logrus"
	"testing"
)

func TestLogger(t *testing.T) {
	var logger = logs.New("test")
	logger.SetLevel(logrus.DebugLevel)
	logger.Debug("test")
	logger.Info("test")
	logger.Warn("test")
	logger.Error("test")
}

func TestOut(t *testing.T) {
	out := bytes.NewBuffer([]byte{})
	logs.SetOutput(out)
	logs.SetLevel(logrus.DebugLevel)

	logs.Debug("test")
	logs.Info("test")
	logs.Warn("test")
	logs.Error("test")

	t.Log(out.String())
}
