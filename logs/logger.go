package logs

import (
	"github.com/sirupsen/logrus"
	"os"
)

var loggers = make([]*logrus.Logger, 0)

var STD = New("root")

func SetLevel(level logrus.Level) {
	logrus.SetLevel(level)
	for _, logger := range loggers {
		logger.SetLevel(level)
	}
}

func SetLogger(debug bool, level string) error {
	if debug {
		SetLevel(logrus.DebugLevel)
	} else if logrusLevel, err := logrus.ParseLevel(level); err != nil {
		return err
	} else {
		SetLevel(logrusLevel)
	}
	return nil
}

func NewLogger(module string, fns ...func(*FieldsHook, *logrus.Logger)) *logrus.Logger {
	logger := logrus.New()
	if module == "root" {
		logger = logrus.StandardLogger()
	}
	logger.SetReportCaller(true)
	logger.SetLevel(logrus.GetLevel())
	logger.SetFormatter(&Formatter{
		TimestampFormat: "2006-01-02 15:04:05.000", FieldsOrder: []string{"module", "engine"},
	})
	logger.SetOutput(os.Stdout)
	//logger.SetReportCaller(true)
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
