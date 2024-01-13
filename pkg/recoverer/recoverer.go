package recoverer

import (
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
				httperror.InternalServerError(req.Context(), w, WithStack(fmt.Errorf("recovered from panic: %s", r)))
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

		*err = errors.Join(*err, WithStack(fmt.Errorf("recovered from panic: %s", r)))
	}
}

func Handler(handler func(error)) {
	if r := recover(); r != nil {
		if handler == nil {
			return
		}

		handler(WithStack(fmt.Errorf("recovered from panic: %s", r)))
	}
}

func Logger() {
	if r := recover(); r != nil {
		slog.Error("recovered from panic", "error", WithStack(fmt.Errorf("%s", r)))
	}
}
