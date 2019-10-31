package prometheus

import (
	"flag"
	"net/http"
	"strings"

	"github.com/ViBiOh/httputils/v2/pkg/flags"
	"github.com/ViBiOh/httputils/v2/pkg/model"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	_ model.Middleware = &app{}
)

// App of package
type App interface {
	Handler(http.Handler) http.Handler
}

// Config of package
type Config struct {
	path *string
}

type app struct {
	path string
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string) Config {
	return Config{
		path: flags.New(prefix, "prometheus").Name("Path").Default("/metrics").Label("Path for exposing metrics").ToString(fs),
	}
}

// New creates new App from Config
func New(config Config) App {
	return &app{
		path: strings.TrimSpace(*config.path),
	}
}

// Handler for net/http
func (a *app) Handler(next http.Handler) http.Handler {
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
			Buckets: []float64{0.25, 0.5, 1, 2.5, 5, 10},
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
