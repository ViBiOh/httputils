package logger

// Close ends logger gracefully
func Close() {
	logger.Close()
}

// Trace logs tracing message
func Trace(format string, a ...interface{}) {
	logger.Trace(format, a...)
}

// Debug logs debug message
func Debug(format string, a ...interface{}) {
	logger.Debug(format, a...)
}

// Info logs info message
func Info(format string, a ...interface{}) {
	logger.Info(format, a...)
}

// Warn logs warning message
func Warn(format string, a ...interface{}) {
	logger.Warn(format, a...)
}

// Error logs error message
func Error(format string, a ...interface{}) {
	logger.Error(format, a...)
}

// Fatal logs error message and exit with status code 1
func Fatal(err error) {
	logger.Fatal(err)
}
