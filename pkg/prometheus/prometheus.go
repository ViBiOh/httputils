package prometheus

import (
	"flag"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/model"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	metricsNamespace = "http"
	metricsEndpoint  = "/metrics"
)

var (
	_ model.Middleware = App{}.Middleware

	durationBuckets  = []float64{0.1, 0.3, 1.2, 5}
	sizeBuckets      = []float64{200, 500, 900, 1500}
	methodLabels     = []string{"method"}
	codeMethodLabels = []string{"code", "method"}
)

type App struct {
	registry *prometheus.Registry
	ignore   []string
	gzip     bool
}

type Config struct {
	ignore *[]string
	gzip   *bool
}

func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) Config {
	return Config{
		ignore: flags.StringSlice(fs, prefix, "prometheus", "Ignore", "Ignored path prefixe for metrics", nil, overrides),
		gzip:   flags.Bool(fs, prefix, "prometheus", "Gzip", "Enable gzip compression of metrics output", true, overrides),
	}
}

func New(config Config) App {
	registry := prometheus.NewRegistry()
	registry.MustRegister(collectors.NewBuildInfoCollector())
	registry.MustRegister(collectors.NewGoCollector())
	registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	return App{
		ignore:   *config.ignore,
		gzip:     *config.gzip,
		registry: registry,
	}
}

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

func (a App) Registerer() prometheus.Registerer {
	return a.registry
}

func (a App) instrumentHandler(next http.Handler) http.Handler {
	durationVec := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: metricsNamespace,
		Subsystem: "request",
		Name:      "duration_seconds",
		Help:      "A histogram of latencies for requests.",
		Buckets:   durationBuckets,
	}, methodLabels)
	a.registry.MustRegister(durationVec)

	sizeVec := prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: metricsNamespace,
		Subsystem: "response",
		Name:      "size_bytes",
		Help:      "A histogram of response sizes for requests.",
		Buckets:   sizeBuckets,
	})
	a.registry.MustRegister(sizeVec)

	counterVec := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: metricsNamespace,
		Subsystem: "requests",
		Name:      "total",
		Help:      "A counter for requests to the wrapped handler.",
	}, codeMethodLabels)
	a.registry.MustRegister(counterVec)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()

		d := newDelegator(w)
		next.ServeHTTP(d, r)

		durationVec.WithLabelValues(r.Method).Observe(d.Time().Sub(now).Seconds())
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
