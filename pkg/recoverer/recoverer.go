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

func Middleware(next http.Handler) http.Handler {
	if next == nil {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				output := make([]byte, OutputSize)
				written := runtime.Stack(output, false)

				httperror.InternalServerError(w, fmt.Errorf("recovered from panic: %s\n%s", r, output[:written]))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func Error(err *error) {
	if r := recover(); r != nil {
		output := make([]byte, 1024)
		written := runtime.Stack(output, false)

		if err == nil {
			return
		}

		*err = errors.Join(*err, fmt.Errorf("recovered from panic: %s", output[:written]))
	}
}

func Logger() {
	if r := recover(); r != nil {
		output := make([]byte, 1024)
		written := runtime.Stack(output, false)

		slog.Error("recovered from panic", "stacktrace", string(output[:written]))
	}
}
