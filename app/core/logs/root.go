package logs

var (
	root   = New("root")
	Debug  = root.Debug
	Debugf = root.Debugf

	Info  = root.Info
	Infof = root.Infof

	Warn  = root.Warn
	Warnf = root.Warnf

	Error  = root.Error
	Errorf = root.Errorf

	Print  = root.Print
	Printf = root.Printf
)
