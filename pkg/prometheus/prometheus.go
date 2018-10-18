package prometheus

import (
	"flag"
	"fmt"
	"net/http"
	"strings"

	"github.com/ViBiOh/httputils/pkg/model"
	"github.com/ViBiOh/httputils/pkg/tools"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var _ model.Middleware = &App{}

// App stores informations
type App struct {
	path string
}

// NewApp creates new App from Flags' config
func NewApp(config map[string]*string) *App {
	return &App{
		path: strings.TrimSpace(*config[`path`]),
	}
}

// Flags adds flags for given prefix
func Flags(prefix string) map[string]*string {
	return map[string]*string{
		`path`: flag.String(tools.ToCamel(fmt.Sprintf(`%sPath`, prefix)), `/metrics`, `[prometheus] Path for exposing metrics`),
	}
}

// Handler for net/http
func (a App) Handler(next http.Handler) http.Handler {
	prometheusHandler := promhttp.Handler()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == a.path {
			prometheusHandler.ServeHTTP(w, r)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
