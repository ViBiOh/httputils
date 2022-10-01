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

type FieldsContext struct {
	outputFn func(level, []field, string, ...any)
	closeFn  func()
	fields   []field
}

func (f FieldsContext) WithField(name string, value any) Provider {
	f.fields = append(f.fields, field{
		name:  name,
		value: value,
	})

	return f
}

func (f FieldsContext) Trace(format string, a ...any) {
	f.outputFn(levelTrace, f.fields, format, a...)
}

func (f FieldsContext) Debug(format string, a ...any) {
	f.outputFn(levelDebug, f.fields, format, a...)
}

func (f FieldsContext) Info(format string, a ...any) {
	f.outputFn(levelInfo, f.fields, format, a...)
}

func (f FieldsContext) Warn(format string, a ...any) {
	f.outputFn(levelWarning, f.fields, format, a...)
}

func (f FieldsContext) Error(format string, a ...any) {
	f.outputFn(levelError, f.fields, format, a...)
}

func (f FieldsContext) Fatal(err error) {
	if err == nil {
		return
	}

	f.outputFn(levelFatal, f.fields, "%s", err)
	f.closeFn()

	exitFunc(1)
}
