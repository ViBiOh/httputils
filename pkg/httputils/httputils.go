package httputils

import (
	"log/slog"
	"net/http"

	"github.com/ViBiOh/httputils/v4/pkg/health"
	"github.com/ViBiOh/httputils/v4/pkg/httprecover"
	"github.com/ViBiOh/httputils/v4/pkg/model"
)

func Handler(handler http.Handler, healthService *health.Service, middlewares ...model.Middleware) http.Handler {
	versionHandler := versionHandler()
	HealthHandler := healthService.HealthHandler()
	readyHandler := healthService.ReadyHandler()
	appHandler := model.ChainMiddlewares(handler, append([]model.Middleware{httprecover.Middleware}, middlewares...)...)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case health.LivePath:
			HealthHandler.ServeHTTP(w, r)

		case health.ReadyPath:
			readyHandler.ServeHTTP(w, r)

		case "/version":
			versionHandler.ServeHTTP(w, r)

		default:
			appHandler.ServeHTTP(w, r)
		}
	})
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
