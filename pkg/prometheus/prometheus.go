package prometheus

import (
	"flag"
	"net/http"
	"strings"

	"github.com/ViBiOh/httputils/v3/pkg/flags"
	"github.com/ViBiOh/httputils/v3/pkg/model"
	"github.com/ViBiOh/httputils/v3/pkg/server"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	metricsEndpoint = "/metrics"
)

var (
	_ model.Middleware = (app{}).Middleware
)

// App of package
type App interface {
	Middleware(http.Handler) http.Handler
	Registerer() prometheus.Registerer
	Handler() http.Handler
}

// Config of package
type Config struct {
	serverConfig server.Config
	ignore       *string
	port         *uint
}

type app struct {
	serverApp server.App
	registry  *prometheus.Registry
	ignore    []string
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) Config {
	return Config{
		ignore: flags.New(prefix, "prometheus").Name("Ignore").Default(flags.Default("Ignore", "", overrides)).Label("Ignored path prefixes for metrics, comma separated").ToString(fs),
	}
}

// New creates new App from Config
func New(config Config) App {
	var ignore []string
	ignoredPaths := strings.TrimSpace(*config.ignore)
	if len(ignoredPaths) != 0 {
		ignore = strings.Split(ignoredPaths, ",")
	}

	return app{
		ignore:   ignore,
		registry: prometheus.NewRegistry(),
	}
}

// Handler for request. Should be use with net/http
func (a app) Handler() http.Handler {
	instrumentHandler := promhttp.InstrumentMetricHandler(
		a.registry, promhttp.HandlerFor(a.registry, promhttp.HandlerOpts{}),
	)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case metricsEndpoint:
			instrumentHandler.ServeHTTP(w, r)

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
}

// Middleware for net/http
func (a app) Middleware(next http.Handler) http.Handler {
	a.registry.MustRegister(prometheus.NewBuildInfoCollector())
	a.registry.MustRegister(prometheus.NewGoCollector())
	a.registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))

	instrumentedHandler := a.instrumentHandler(next)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if a.isIgnored(r.URL.Path) {
			next.ServeHTTP(w, r)
		} else {
			instrumentedHandler.ServeHTTP(w, r)
		}
	})
}

// Registerer return served registerer
func (a app) Registerer() prometheus.Registerer {
	return a.registry
}

func (a app) instrumentHandler(next http.Handler) http.Handler {
	instrumentedHandler := next

	durationVec := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "A histogram of latencies for requests.",
		Buckets: []float64{0.25, 0.5, 1, 2.5, 5, 10},
	}, []string{"code", "method"})
	a.registry.MustRegister(durationVec)
	instrumentedHandler = promhttp.InstrumentHandlerDuration(durationVec, instrumentedHandler)

	sizeVec := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_response_size_bytes",
		Help:    "A histogram of response sizes for requests.",
		Buckets: []float64{200, 500, 900, 1500},
	}, nil)
	a.registry.MustRegister(sizeVec)
	instrumentedHandler = promhttp.InstrumentHandlerResponseSize(sizeVec, instrumentedHandler)

	counterVec := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "A counter for requests to the wrapped handler.",
	},
		[]string{"code", "method"})
	a.registry.MustRegister(counterVec)
	instrumentedHandler = promhttp.InstrumentHandlerCounter(counterVec, instrumentedHandler)

	return instrumentedHandler
}

func (a app) isIgnored(path string) bool {
	for _, prefix := range a.ignore {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}

	return false
}
