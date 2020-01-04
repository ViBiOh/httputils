package prometheus

import (
	"flag"
	"fmt"
	"net/http"
	"strings"

	"github.com/ViBiOh/httputils/v3/pkg/flags"
	"github.com/ViBiOh/httputils/v3/pkg/model"
	"github.com/ViBiOh/httputils/v3/pkg/swagger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	_ model.Middleware = (&app{}).Middleware
	_ swagger.Provider = (&app{}).Swagger
)

// App of package
type App interface {
	Middleware(http.Handler) http.Handler
	Registerer() prometheus.Registerer
	Swagger() (swagger.Configuration, error)
}

// Config of package
type Config struct {
	path *string
}

type app struct {
	path string

	registry *prometheus.Registry
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
		path:     strings.TrimSpace(*config.path),
		registry: prometheus.NewRegistry(),
	}
}

// Middleware for net/http
func (a app) Middleware(next http.Handler) http.Handler {
	a.registry.MustRegister(prometheus.NewGoCollector())

	prometheusHandler := promhttp.InstrumentMetricHandler(
		a.registry, promhttp.HandlerFor(a.registry, promhttp.HandlerOpts{}),
	)
	instrumentedHandler := a.instrumentHandler(next)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == a.path {
			prometheusHandler.ServeHTTP(w, r)
		} else {
			instrumentedHandler.ServeHTTP(w, r)
		}
	})
}

// Registerer return served registerer
func (a app) Registerer() prometheus.Registerer {
	return a.registry
}

// Registerer return served registerer
func (a app) Swagger() (swagger.Configuration, error) {
	return swagger.Configuration{
		Paths: fmt.Sprintf(`%s:
  get:
    description: Retrieves metrics of app

    responses:
      200:
        description: Metrics of app
        content:
          text/plain:
            schema:
              type: string`, a.path),
	}, nil
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
	},
		[]string{})
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
