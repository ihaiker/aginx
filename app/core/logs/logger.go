package logs

import (
	"github.com/sirupsen/logrus"
	"io"
	"os"
)

var loggers = make([]*logrus.Logger, 0)
var output io.Writer = os.Stdout

func GetOutput() io.Writer {
	return output
}
func SetOutput(out io.Writer) {
	output = out
	for _, log := range loggers {
		log.SetOutput(out)
	}
}

func SetLevel(level logrus.Level) {
	logrus.SetLevel(level)
	for _, logger := range loggers {
		logger.SetLevel(level)
	}
}

func NewLogger(module string, fns ...func(*FieldsHook, *logrus.Logger)) *logrus.Logger {
	logger := logrus.New()
	if module == "root" {
		logger = logrus.StandardLogger()
	}
	logger.SetReportCaller(true)
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&Formatter{
		TimestampFormat: "2006-01-02 15:04:05.000", FieldsOrder: []string{"module", "engine"},
	})
	logger.SetOutput(output)
	hook := &FieldsHook{fields: map[string]interface{}{}}
	hook.fields["module"] = module
	for _, fn := range fns {
		fn(hook, logger)
	}
	logger.AddHook(hook)
	loggers = append(loggers, logger)
	return logger
}

func New(module string, fields ...string) *logrus.Logger {
	return NewLogger(module, func(hook *FieldsHook, logger *logrus.Logger) {
		for i := 0; i < len(fields)/2; i += 2 {
			hook.fields[fields[i*2]] = fields[i*2+1]
		}
	})
}
