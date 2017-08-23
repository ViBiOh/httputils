package prometheus

import (
	"flag"
	"net/http"
	"regexp"
	"runtime"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	metricsPath = flag.String(`prometheusMetricsPath`, `/metrics`, `Prometheus - Allowed regex IP to call metrics endpoint`)
	metricsIps  = flag.String(`prometheusMetricsIP`, `*`, `Prometheus - Allowed regex IP to call metrics endpoint`)
)

var metricsIpsRegex *regexp.Regexp

func goroutinesHandler(gauge prometheus.Gauge, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gauge.Set(float64(runtime.NumGoroutine()))
		next.ServeHTTP(w, r)
	})
}

func getPrometheusHandlers(prefix string, next http.Handler) (http.HandlerFunc, http.Handler) {
	requestsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: prefix + `_requests_total`,
			Help: `Total number of HTTP requests made.`,
		},
		[]string{`method`, `code`},
	)

	duration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:        prefix + `_request_duration_seconds`,
		Help:        `A histogram of latencies for requests.`,
		Buckets:     []float64{.25, .5, 1, 2.5, 5, 10},
		ConstLabels: prometheus.Labels{`handler`: `push`},
	},
		[]string{`method`},
	)

	goroutines := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: prefix + `_goroutines`,
		Help: `Number of current goroutines.`,
	})

	registry := prometheus.NewRegistry()
	registry.MustRegister(requestsTotal)
	registry.MustRegister(duration)
	registry.MustRegister(goroutines)

	return promhttp.InstrumentHandlerCounter(requestsTotal, promhttp.InstrumentHandlerDuration(duration, next)), goroutinesHandler(goroutines, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
}

// NewPrometheusHandler wraps given handler into prometheus tooling and expose `/metrics` endpoints
func NewPrometheusHandler(prefix string, next http.Handler) http.Handler {
	handler, metrics := getPrometheusHandlers(prefix, next)

	if metricsIpsRegex == nil {
		metricsIpsRegex = regexp.MustCompile(*metricsIps)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == *metricsPath {
			if metricsIpsRegex.MatchString(r.RemoteAddr) {
				metrics.ServeHTTP(w, r)
			} else {
				w.WriteHeader(http.StatusForbidden)
			}
		} else {
			handler.ServeHTTP(w, r)
		}
	})
}
