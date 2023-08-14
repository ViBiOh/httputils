package logger

import (
	"log/slog"
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
	level     slog.Level
}

type FieldsContext struct {
	outputFn func(slog.Level, []field, string, ...any)
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

func (f FieldsContext) Debug(format string, a ...any) {
	f.outputFn(slog.LevelDebug, f.fields, format, a...)
}

func (f FieldsContext) Info(format string, a ...any) {
	f.outputFn(slog.LevelInfo, f.fields, format, a...)
}

func (f FieldsContext) Warn(format string, a ...any) {
	f.outputFn(slog.LevelWarn, f.fields, format, a...)
}

func (f FieldsContext) Error(format string, a ...any) {
	f.outputFn(slog.LevelError, f.fields, format, a...)
}

func (f FieldsContext) Fatal(err error) {
	if err == nil {
		return
	}

	f.outputFn(slog.LevelError, f.fields, "%s", err)
	f.closeFn()

	exitFunc(1)
}
