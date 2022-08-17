package logger

import (
	"time"
)

type field struct {
	value any
	name  string
}

type event struct {
	timestamp time.Time
	message   string
	fields    []field
	level     level
}

// FieldsContext contains field context.
type FieldsContext struct {
	outputFn func(level, []field, string, ...any)
	closeFn  func()
	fields   []field
}

// WithField add a field to current context.
func (f FieldsContext) WithField(name string, value any) Provider {
	f.fields = append(f.fields, field{
		name:  name,
		value: value,
	})

	return f
}

// Trace logs tracing message.
func (f FieldsContext) Trace(format string, a ...any) {
	f.outputFn(levelTrace, f.fields, format, a...)
}

// Debug logs debug message.
func (f FieldsContext) Debug(format string, a ...any) {
	f.outputFn(levelDebug, f.fields, format, a...)
}

// Info logs info message.
func (f FieldsContext) Info(format string, a ...any) {
	f.outputFn(levelInfo, f.fields, format, a...)
}

// Warn logs warning message.
func (f FieldsContext) Warn(format string, a ...any) {
	f.outputFn(levelWarning, f.fields, format, a...)
}

// Error logs error message.
func (f FieldsContext) Error(format string, a ...any) {
	f.outputFn(levelError, f.fields, format, a...)
}

// Fatal logs error message and exit with status code 1.
func (f FieldsContext) Fatal(err error) {
	if err == nil {
		return
	}

	f.outputFn(levelFatal, f.fields, "%s", err)
	f.closeFn()

	exitFunc(1)
}
