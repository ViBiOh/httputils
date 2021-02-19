package httputils

import (
	"net/http"
	"os"

	"github.com/ViBiOh/httputils/v3/pkg/health"
	"github.com/ViBiOh/httputils/v3/pkg/logger"
	"github.com/ViBiOh/httputils/v3/pkg/model"
)

// App of package
type App interface {
	Handler(http.Handler, health.App, ...model.Middleware) http.Handler
}

// Handler creates the handler for default httputils behavior
func Handler(handler http.Handler, healthApp health.App, middlewares ...model.Middleware) http.Handler {
	versionHandler := versionHandler()
	healthHandler := healthApp.Handler()
	apphandler := model.ChainMiddlewares(handler, middlewares...)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/health":
			healthHandler.ServeHTTP(w, r)

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
