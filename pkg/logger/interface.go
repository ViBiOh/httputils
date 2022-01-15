package logger

// Provider of logging definition
type Provider interface {
	WithField(name string, value interface{}) Provider
	Trace(format string, a ...interface{})
	Debug(format string, a ...interface{})
	Info(format string, a ...interface{})
	Warn(format string, a ...interface{})
	Error(format string, a ...interface{})
	Fatal(err error)
}
