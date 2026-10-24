package httprecover

import (
	"net/http"

	"github.com/ViBiOh/httputils/v4/pkg/httperror"
	"github.com/ViBiOh/httputils/v4/pkg/recoverer"
)

func Middleware(next http.Handler) http.Handler {
	if next == nil {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				if recovered == http.ErrAbortHandler {
					panic(recovered)
				}

				httperror.InternalServerError(req.Context(), w, recoverer.WithStack(recoverer.Format(recovered)))
			}
		}()

		next.ServeHTTP(w, req)
	})
}
