package prometheus

import (
	"flag"
	"fmt"
	"net/http"
	"strings"

	"github.com/ViBiOh/httputils/pkg/tools"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// App stores informations
type App struct {
	endpoint string
}

// NewApp creates new App from Flags' config
func NewApp(config map[string]*string) *App {
	return &App{
		endpoint: *config[`endpoint`],
	}
}

// Flags adds flags for given prefix
func Flags(prefix string) map[string]*string {
	return map[string]*string{
		`endpoint`: flag.String(tools.ToCamel(fmt.Sprintf(`%sEndpoint`, prefix)), `/metrics`, `[prometheus] Metrics endpoint`),
	}
}

// Handler for request. Should be use with net/http
func (a App) Handler(next http.Handler) http.Handler {
	handler := promhttp.Handler()
	if next == nil {
		return handler
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, a.endpoint) {
			handler.ServeHTTP(w, r)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
