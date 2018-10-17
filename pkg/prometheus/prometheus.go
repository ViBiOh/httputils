package prometheus

import (
	"net/http"

	"github.com/ViBiOh/httputils/pkg/model"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var _ model.Middleware = &App{}

// App stores informations
type App struct {
}

// NewApp creates new App from Flags' config
func NewApp(_ map[string]*string) *App {
	return &App{}
}

// Flags adds flags for given prefix
func Flags(_ string) map[string]*string {
	return nil
}

// Handler for net/http
func (a App) Handler(next http.Handler) http.Handler {
	prometheusHandler := promhttp.Handler()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == `/metrics` {
			prometheusHandler.ServeHTTP(w, r)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
