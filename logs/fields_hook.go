package logs

import (
	"github.com/sirupsen/logrus"
)

type FieldsHook struct {
	fields logrus.Fields
}

func (h *FieldsHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *FieldsHook) Fire(e *logrus.Entry) error {
	for s, i := range h.fields {
		e.Data[s] = i
	}
	return nil
}
