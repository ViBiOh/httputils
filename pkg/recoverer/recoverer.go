package recoverer

import (
	"fmt"
	"net/http"
	"runtime"

	"github.com/ViBiOh/httputils/v4/pkg/httperror"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
)

// Middleware for request. Should be use with net/http
func Middleware(next http.Handler) http.Handler {
	if next == nil {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				output := make([]byte, 1024)
				written := runtime.Stack(output, false)

				httperror.InternalServerError(w, fmt.Errorf("recovered from panic: %s\n%s", r, output[:written]))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// LoggerRecover catch panic and output it in log
func LoggerRecover() {
	if r := recover(); r != nil {
		output := make([]byte, 1024)
		written := runtime.Stack(output, false)

		logger.Error("recovered from panic: %s\n", output[:written])
	}
}
