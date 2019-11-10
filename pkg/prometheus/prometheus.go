package prometheus

import (
	"flag"
	"net/http"
	"strings"

	"github.com/ViBiOh/httputils/v3/pkg/flags"
	"github.com/ViBiOh/httputils/v3/pkg/model"
	"github.com/prometheus/client_golang/prometheus"
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
	registry := prometheus.NewRegistry()
	registry.MustRegister(prometheus.NewGoCollector())

	prometheusHandler := promhttp.InstrumentMetricHandler(
		registry, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}),
	)
	instrumentedHandler := instrumentHandler(registry, next)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == a.path {
			prometheusHandler.ServeHTTP(w, r)
		} else {
			instrumentedHandler.ServeHTTP(w, r)
		}
	})
}

func instrumentHandler(registerer prometheus.Registerer, next http.Handler) http.Handler {
	instrumentedHandler := next

	durationVec := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "A histogram of latencies for requests.",
		Buckets: []float64{0.25, 0.5, 1, 2.5, 5, 10},
	}, []string{"code", "method"})
	registerer.MustRegister(durationVec)
	instrumentedHandler = promhttp.InstrumentHandlerDuration(durationVec, instrumentedHandler)

	sizeVec := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_response_size_bytes",
		Help:    "A histogram of response sizes for requests.",
		Buckets: []float64{200, 500, 900, 1500},
	},
		[]string{})
	registerer.MustRegister(sizeVec)
	instrumentedHandler = promhttp.InstrumentHandlerResponseSize(sizeVec, instrumentedHandler)

	counterVec := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "A counter for requests to the wrapped handler.",
	},
		[]string{"code", "method"})
	registerer.MustRegister(counterVec)
	instrumentedHandler = promhttp.InstrumentHandlerCounter(counterVec, instrumentedHandler)

	return instrumentedHandler
}
