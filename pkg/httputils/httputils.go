package httputils

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/ViBiOh/httputils/v4/pkg/health"
	"github.com/ViBiOh/httputils/v4/pkg/httprecover"
	"github.com/ViBiOh/httputils/v4/pkg/model"
)

func Handler(handler http.Handler, healthService *health.Service, middlewares ...model.Middleware) http.Handler {
	mux := http.NewServeMux()

	mux.Handle(fmt.Sprintf("GET %s", health.LivePath), healthService.HealthHandler())
	mux.Handle(fmt.Sprintf("GET %s", health.ReadyPath), healthService.ReadyHandler())

	mux.Handle("GET /version", versionHandler())
	mux.Handle("/", model.ChainMiddlewares(handler, append([]model.Middleware{httprecover.Middleware}, middlewares...)...))

	return mux
}

func versionHandler() http.Handler {
	version := []byte(model.Version())

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)

			return
		}

		if _, err := w.Write(version); err != nil {
			slog.LogAttrs(r.Context(), slog.LevelError, "write", slog.Any("error", err))
		}
	})
}
