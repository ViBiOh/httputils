package prometheus

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func BenchmarkCounterVec(b *testing.B) {
	registry := prometheus.NewRegistry()
	counter := CounterVec(registry, "benchmark", "prometheus", "vector", "state")

	for i := 0; i < b.N; i++ {
		counter.WithLabelValues("valid").Inc()
	}
}

func BenchmarkCounter(b *testing.B) {
	registry := prometheus.NewRegistry()
	counter := createCounter(registry, "benchmark", "prometheus", "counter")

	for i := 0; i < b.N; i++ {
		counter.Inc()
	}
}
