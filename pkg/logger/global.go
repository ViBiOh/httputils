package logger

// Global sets global logger.
func Global(l Logger) {
	defer logger.Close()
	logger = l
}

// GetGlobal returns global logger.
func GetGlobal() Logger {
	return logger
}

// Close ends logger gracefully.
func Close() {
	logger.Close()
}

// WithField create context for logging.
func WithField(name string, value any) Provider {
	return logger.WithField(name, value)
}

// Trace logs tracing message.
func Trace(format string, a ...any) {
	logger.Trace(format, a...)
}

// Debug logs debug message.
func Debug(format string, a ...any) {
	logger.Debug(format, a...)
}

// Info logs info message.
func Info(format string, a ...any) {
	logger.Info(format, a...)
}

// Warn logs warning message.
func Warn(format string, a ...any) {
	logger.Warn(format, a...)
}

// Error logs error message.
func Error(format string, a ...any) {
	logger.Error(format, a...)
}

// Fatal logs error message and exit with status code 1.
func Fatal(err error) {
	logger.Fatal(err)
}
