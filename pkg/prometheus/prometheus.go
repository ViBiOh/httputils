package prometheus

import (
	"flag"
	"fmt"
	"net/http"
	"strings"

	"github.com/ViBiOh/httputils/pkg/model"
	"github.com/ViBiOh/httputils/pkg/tools"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
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
	prometheusHandler := promhttp.Handler()
	instrumentedHandler := instrumentHandler(next)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == a.path {
			prometheusHandler.ServeHTTP(w, r)
		} else {
			instrumentedHandler.ServeHTTP(w, r)
		}
	})
}

func instrumentHandler(next http.Handler) http.Handler {
	instrumentedHandler := next

	instrumentedHandler = promhttp.InstrumentHandlerDuration(promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "A histogram of latencies for requests.",
			Buckets: []float64{.25, .5, 1, 2.5, 5, 10},
		},
		[]string{"code", "method"}), instrumentedHandler)

	instrumentedHandler = promhttp.InstrumentHandlerResponseSize(promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_size_bytes",
			Help:    "A histogram of response sizes for requests.",
			Buckets: []float64{200, 500, 900, 1500},
		},
		[]string{}), instrumentedHandler)

	instrumentedHandler = promhttp.InstrumentHandlerCounter(promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "A counter for requests to the wrapped handler.",
		},
		[]string{"code", "method"}), instrumentedHandler)

	return instrumentedHandler
}
