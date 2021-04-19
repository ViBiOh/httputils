package logger

import (
	"time"
)

type event struct {
	fields    map[string]interface{}
	timestamp time.Time
	message   string
	level     level
}

// FieldsContext contains field context
type FieldsContext struct {
	fields   map[string]interface{}
	outputFn func(level, map[string]interface{}, string, ...interface{})
	closeFn  func()
}

// WithField add a field to current context
func (f FieldsContext) WithField(name string, value interface{}) FieldsContext {
	f.fields[name] = value

	return f
}

// Trace logs tracing message
func (f FieldsContext) Trace(format string, a ...interface{}) {
	f.outputFn(levelTrace, f.fields, format, a...)
}

// Debug logs debug message
func (f FieldsContext) Debug(format string, a ...interface{}) {
	f.outputFn(levelDebug, f.fields, format, a...)
}

// Info logs info message
func (f FieldsContext) Info(format string, a ...interface{}) {
	f.outputFn(levelInfo, f.fields, format, a...)
}

// Warn logs warning message
func (f FieldsContext) Warn(format string, a ...interface{}) {
	f.outputFn(levelWarning, f.fields, format, a...)
}

// Error logs error message
func (f FieldsContext) Error(format string, a ...interface{}) {
	f.outputFn(levelError, f.fields, format, a...)
}

// Fatal logs error message and exit with status code 1
func (f FieldsContext) Fatal(err error) {
	if err == nil {
		return
	}

	f.outputFn(levelFatal, f.fields, "%s", err)
	f.closeFn()

	exitFunc(1)
}
