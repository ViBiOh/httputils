package recoverer

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"runtime"

	"github.com/ViBiOh/httputils/v4/pkg/httperror"
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

func Middleware(next http.Handler) http.Handler {
	if next == nil {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				httperror.InternalServerError(req.Context(), w, WithStack(formatRecover(r)))
			}
		}()

		next.ServeHTTP(w, req)
	})
}

func Error(err *error) {
	if r := recover(); r != nil {
		if err == nil {
			return
		}

		recoverErr := WithStack(formatRecover(r))

		// Don't erase a potential error already present
		if *err != nil {
			*err = errors.Join(*err, recoverErr)
		} else {
			*err = recoverErr
		}
	}
}

func Handler(handler func(error)) {
	if r := recover(); r != nil {
		if handler == nil {
			return
		}

		handler(WithStack(formatRecover(r)))
	}
}

func Logger() {
	if r := recover(); r != nil {
		slog.LogAttrs(context.Background(), slog.LevelError, fmt.Sprintf("recovered from panic: %s", r), slog.Any("error", WithStack(fmt.Errorf("%s", r))))
	}
}

func formatRecover(r any) error {
	return fmt.Errorf("recovered from panic: %s", r)
}
