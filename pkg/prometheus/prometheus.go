package prometheus

import (
	"flag"
	"fmt"
	"net/http"
	"strings"

	"github.com/ViBiOh/httputils/pkg/model"
	"github.com/ViBiOh/httputils/pkg/tools"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var _ model.Middleware = &App{}

// Config of package
type Config struct {
	path *string
}

// App of package
type App struct {
	path string
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string) Config {
	docPrefix := prefix
	if prefix == "" {
		docPrefix = "prometheus"
	}

	return Config{
		path: fs.String(tools.ToCamel(fmt.Sprintf("%sPath", prefix)), "/metrics", fmt.Sprintf("[%s] Path for exposing metrics", docPrefix)),
	}
}

// New creates new App from Config
func New(config Config) *App {
	return &App{
		path: strings.TrimSpace(*config.path),
	}
}

// Handler for net/http
func (a App) Handler(next http.Handler) http.Handler {
	instrumentedHandler := promhttp.InstrumentMetricHandler(
		prometheus.DefaultRegisterer, next,
	)
	prometheusHandler := promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{})

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == a.path {
			prometheusHandler.ServeHTTP(w, r)
		} else {
			instrumentedHandler.ServeHTTP(w, r)
		}
	})
}
