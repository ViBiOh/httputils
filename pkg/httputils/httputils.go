package httputils

import (
	"net/http"
	"os"

	"github.com/ViBiOh/httputils/v4/pkg/health"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/httputils/v4/pkg/model"
)

// Handler creates the handler for default httputils behavior
func Handler(handler http.Handler, healthApp health.App, middlewares ...model.Middleware) http.Handler {
	versionHandler := versionHandler()
	HealthHandler := healthApp.HealthHandler()
	readyHandler := healthApp.ReadyHandler()
	apphandler := model.ChainMiddlewares(handler, middlewares...)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case health.LivePath:
			HealthHandler.ServeHTTP(w, r)

		case health.ReadyPath:
			readyHandler.ServeHTTP(w, r)

		case "/version":
			versionHandler.ServeHTTP(w, r)

		default:
			apphandler.ServeHTTP(w, r)
		}
	})
}

func versionHandler() http.Handler {
	versionValue := os.Getenv("VERSION")
	if len(versionValue) == 0 {
		versionValue = "development"
	}
	version := []byte(versionValue)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if _, err := w.Write(version); err != nil {
			logger.Error("%s", err)
		}
	})
}
