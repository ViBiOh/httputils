package recoverer

import (
	"fmt"
	"net/http"
	"runtime"

	"github.com/ViBiOh/httputils/v4/pkg/httperror"
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
				runtime.Stack(output, false)

				httperror.InternalServerError(w, fmt.Errorf("recovered from panic: %s\n%s", r, output))
			}
		}()

		next.ServeHTTP(w, r)
	})
}
