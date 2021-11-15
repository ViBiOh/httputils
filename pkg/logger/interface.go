package logger

// Provider definition
type Provider interface {
	WithField(name string, value interface{}) FieldsContext
	Trace(format string, a ...interface{})
	Debug(format string, a ...interface{})
	Info(format string, a ...interface{})
	Warn(format string, a ...interface{})
	Error(format string, a ...interface{})
	Fatal(err error)
}
