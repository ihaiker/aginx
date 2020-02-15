package logs

import (
	"github.com/sirupsen/logrus"
	"testing"
)

var logger = New("test")

func TestLogger(t *testing.T) {
	logger.SetLevel(logrus.DebugLevel)
	logger.Debug("test")
	logger.Info("test")
	logger.Warn("test")
	logger.Error("test")
}
