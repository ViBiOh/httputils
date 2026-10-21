package recoverer

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"runtime"
)

const OutputSize = 8192

var _ error = errWithStackTrace{}

type errWithStackTrace struct {
	error
	stackTrace []byte
}

func WithStack(err error) errWithStackTrace {
	output := make([]byte, OutputSize)
	written := runtime.Stack(output, false)

	return errWithStackTrace{
		error:      err,
		stackTrace: output[:written],
	}
}

func (e errWithStackTrace) StackTrace() string {
	return string(e.stackTrace)
}

func Error(err *error) {
	recovered := recover()
	if recovered == nil {
		return
	}

	rethrowIfNeeded(recovered)

	if err == nil {
		panic(recovered)
	}

	*err = errors.Join(*err, WithStack(Format(recovered)))
}

func Handler(handler func(error)) {
	recovered := recover()
	if recovered == nil {
		return
	}

	rethrowIfNeeded(recovered)

	if handler == nil {
		return
	}

	handler(WithStack(Format(recovered)))
}

func Logger() {
	recovered := recover()
	if recovered == nil {
		return
	}

	rethrowIfNeeded(recovered)

	slog.LogAttrs(context.Background(), slog.LevelError, fmt.Sprintf("recovered from panic: %s", recovered), slog.Any("error", WithStack(fmt.Errorf("%s", recovered))))
}

func Format(r any) error {
	return fmt.Errorf("recovered from panic: %s", r)
}

func rethrowIfNeeded(recovered any) {
	if recovered == http.ErrAbortHandler {
		panic(recovered)
	}
}
