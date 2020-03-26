package logs

import (
	"bytes"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// Formatter - logrus formatter, implements logrus.Formatter
type Formatter struct {
	FieldsOrder     []string // default: fields sorted alphabetically
	TimestampFormat string   // default: time.StampMilli = "Jan _2 15:04:05.000"
	HideKeys        bool     // show [fieldValue] instead of [fieldKey:fieldValue]
	NoColors        bool     // disable colors
	NoFieldsColors  bool     // color only level, default is level + fields
	ShowFullLevel   bool     // true to show full level [WARNING] instead [WARN]
	TrimMessages    bool     // true to trim whitespace on messages
	ShowNilField    bool
}

// Format an log entry
func (f *Formatter) Format(entry *logrus.Entry) ([]byte, error) {
	levelColor := getColorByLevel(entry.Level)

	timestampFormat := f.TimestampFormat
	if timestampFormat == "" {
		timestampFormat = time.StampMilli
	}

	// output buffer
	b := &bytes.Buffer{}

	{
		// write level
		level := strings.ToUpper(entry.Level.String())
		if !f.NoColors {
			fmt.Fprintf(b, "\x1b[%dm", levelColor)
		}
		b.WriteString("[")
		if f.ShowFullLevel {
			b.WriteString(level)
		} else {
			b.WriteString(level[:4])
		}
		b.WriteString("] ")
		if !f.NoColors && !f.NoFieldsColors {
			b.WriteString("\x1b[0m")
		}
	}

	// write time
	b.WriteString(entry.Time.Format(timestampFormat))
	b.WriteString(" ")

	{
		b.WriteString(fmt.Sprintf("\x1b[36m[%v]\x1b[0m ", entry.Data["module"]))
		delete(entry.Data, "module")
	}

	if entry.HasCaller() {

		idx := strings.Index(entry.Caller.Function, ".(")
		if idx == -1 {
			fnPageName := filepath.Base(entry.Caller.Function)
			idx = strings.Index(fnPageName, ".")
			file := filepath.Dir(entry.Caller.Function) + "/" + fnPageName[0:idx] + "/" + filepath.Base(entry.Caller.File)
			fmt.Fprintf(b, "%s:%d ", file, entry.Caller.Line)
		} else {
			fn := entry.Caller.Function[idx+1:]
			file := entry.Caller.Function[0:idx] + "/" + filepath.Base(entry.Caller.File)
			fmt.Fprintf(b, "%s:%d %s() ", file, entry.Caller.Line, fn)
		}
	}

	// write fields
	if f.FieldsOrder == nil {
		f.writeFields(b, entry)
	} else {
		f.writeOrderedFields(b, entry)
	}

	// write message
	if f.TrimMessages {
		b.WriteString(strings.TrimSpace(entry.Message))
	} else {
		b.WriteString(entry.Message)
	}

	b.WriteByte('\n')

	return b.Bytes(), nil
}

func (f *Formatter) writeFields(b *bytes.Buffer, entry *logrus.Entry) {
	if len(entry.Data) != 0 {
		fields := make([]string, 0, len(entry.Data))
		for field := range entry.Data {
			fields = append(fields, field)
		}

		sort.Strings(fields)

		for _, field := range fields {
			f.writeField(b, entry, field)
		}
	}
}

func (f *Formatter) writeOrderedFields(b *bytes.Buffer, entry *logrus.Entry) {
	length := len(entry.Data)
	foundFieldsMap := map[string]bool{}
	for _, field := range f.FieldsOrder {
		if _, ok := entry.Data[field]; ok {
			foundFieldsMap[field] = true
			length--
			f.writeField(b, entry, field)
		}
	}

	if length > 0 {
		notFoundFields := make([]string, 0, length)
		for field := range entry.Data {
			if foundFieldsMap[field] == false {
				notFoundFields = append(notFoundFields, field)
			}
		}

		sort.Strings(notFoundFields)

		for _, field := range notFoundFields {
			f.writeField(b, entry, field)
		}
	}
}

func (f *Formatter) writeField(b *bytes.Buffer, entry *logrus.Entry, field string) {
	if value := entry.Data[field]; value == nil && !f.ShowNilField {
		return
	}
	if f.HideKeys {
		fmt.Fprintf(b, "(%v) ", entry.Data[field])
	} else {
		fmt.Fprintf(b, "(%s=%v) ", field, entry.Data[field])
	}
}

const (
	colorRed    = 31
	colorYellow = 33
	colorBlue   = 36
	colorGray   = 37
)

func getColorByLevel(level logrus.Level) int {
	switch level {
	case logrus.DebugLevel:
		return colorGray
	case logrus.WarnLevel:
		return colorYellow
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		return colorRed
	default:
		return colorBlue
	}
}
