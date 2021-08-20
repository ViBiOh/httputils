package prometheus

import (
	"flag"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/flags"
	"github.com/ViBiOh/httputils/v4/pkg/model"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	metricsEndpoint = "/metrics"
)

var (
	_ model.Middleware = App{}.Middleware

	durationBuckets  = []float64{0.1, 0.3, 1.2, 5}
	sizeBuckets      = []float64{200, 500, 900, 1500}
	methodLabels     = []string{"method"}
	codeMethodLabels = []string{"code", "method"}
)

// App of package
type App struct {
	registry *prometheus.Registry
	ignore   []string
	gzip     bool
}

// Config of package
type Config struct {
	ignore *string
	gzip   *bool
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) Config {
	return Config{
		ignore: flags.New(prefix, "prometheus", "Ignore").Default("", overrides).Label("Ignored path prefixes for metrics, comma separated").ToString(fs),
		gzip:   flags.New(prefix, "prometheus", "Gzip").Default(true, overrides).Label("Enable gzip compression of metrics output").ToBool(fs),
	}
}

// New creates new App from Config
func New(config Config) App {
	var ignore []string
	ignoredPaths := *config.ignore
	if len(ignoredPaths) != 0 {
		ignore = strings.Split(ignoredPaths, ",")
	}

	registry := prometheus.NewRegistry()
	registry.MustRegister(collectors.NewBuildInfoCollector())
	registry.MustRegister(collectors.NewGoCollector())
	registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	return App{
		ignore:   ignore,
		gzip:     *config.gzip,
		registry: registry,
	}
}

// Handler for request. Should be use with net/http
func (a App) Handler() http.Handler {
	instrumentHandler := promhttp.HandlerFor(a.registry, promhttp.HandlerOpts{
		DisableCompression: !a.gzip,
	})

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
func (a App) Middleware(next http.Handler) http.Handler {
	if next == nil {
		return next
	}

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
func (a App) Registerer() prometheus.Registerer {
	return a.registry
}

func (a App) instrumentHandler(next http.Handler) http.Handler {
	durationVec := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "A histogram of latencies for requests.",
		Buckets: durationBuckets,
	}, methodLabels)
	a.registry.MustRegister(durationVec)

	sizeVec := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "http_response_size_bytes",
		Help:    "A histogram of response sizes for requests.",
		Buckets: sizeBuckets,
	})
	a.registry.MustRegister(sizeVec)

	counterVec := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "A counter for requests to the wrapped handler.",
	}, codeMethodLabels)
	a.registry.MustRegister(counterVec)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()

		d := newDelegator(w)
		next.ServeHTTP(d, r)

		durationVec.WithLabelValues(r.Method).Observe(time.Since(now).Seconds())
		counterVec.WithLabelValues(strconv.Itoa(d.Status()), r.Method).Inc()
		sizeVec.Observe(float64(d.Written()))
	})
}

func (a App) isIgnored(path string) bool {
	for _, prefix := range a.ignore {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}

	return false
}
