package recoverer

import (
	"fmt"
	"net/http"

	"github.com/ViBiOh/httputils/v4/pkg/httperror"
)

// Middleware for request. Should be use with net/http
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				httperror.InternalServerError(w, fmt.Errorf("recovered from panic: %s", r))
			}
		}()

		if next != nil {
			next.ServeHTTP(w, r)
		}
	})
}
