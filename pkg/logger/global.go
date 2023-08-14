package logger

func Global(l *Logger) {
	defer logger.Close()
	logger = l
}

func GetGlobal() *Logger {
	return logger
}

func Close() {
	logger.Close()
}

func WithField(name string, value any) Provider {
	return logger.WithField(name, value)
}

func Debug(format string, a ...any) {
	logger.Debug(format, a...)
}

func Info(format string, a ...any) {
	logger.Info(format, a...)
}

func Warn(format string, a ...any) {
	logger.Warn(format, a...)
}

func Error(format string, a ...any) {
	logger.Error(format, a...)
}

func Fatal(err error) {
	logger.Fatal(err)
}
