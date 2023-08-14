package logger

type Provider interface {
	WithField(name string, value any) Provider
	Debug(format string, a ...any)
	Info(format string, a ...any)
	Warn(format string, a ...any)
	Error(format string, a ...any)
	Fatal(err error)
}
