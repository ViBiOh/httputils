package prometheus

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func getPrometheusHandlers(next http.Handler) (http.HandlerFunc, http.Handler) {
	requestsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "requests_total",
			Help: "Total number of HTTP requests made.",
		},
		[]string{"method", "code"},
	)

	duration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:        "request_duration_seconds",
		Help:        "A histogram of latencies for requests.",
		Buckets:     []float64{.25, .5, 1, 2.5, 5, 10},
		ConstLabels: prometheus.Labels{"handler": "push"},
	},
		[]string{"method"},
	)

	registry := prometheus.NewRegistry()
	registry.MustRegister(requestsTotal)
	registry.MustRegister(duration)

	return promhttp.InstrumentHandlerCounter(requestsTotal, promhttp.InstrumentHandlerDuration(duration, next)), promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
}

// NewPrometheusHandler wraps given handler into prometheus tooling and expose `/metrics` endpoints
func NewPrometheusHandler(next http.Handler) http.Handler {
	handler, metrics := getPrometheusHandlers(next)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == `/metrics` {
			metrics.ServeHTTP(w, r)
		} else {
			handler.ServeHTTP(w, r)
		}
	})
}
