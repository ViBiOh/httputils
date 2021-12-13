package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
)

// Counters creates and registers counters
func Counters(prometheusRegisterer prometheus.Registerer, namespace, subsystem string, names ...string) map[string]prometheus.Counter {
	if isNil(prometheusRegisterer) {
		return nil
	}

	metrics := make(map[string]prometheus.Counter)

	for _, name := range names {
		metrics[name] = Counter(prometheusRegisterer, namespace, subsystem, name)
	}

	return metrics
}

// Counter creates and registers a counter
func Counter(prometheusRegisterer prometheus.Registerer, namespace, subsystem, name string) prometheus.Counter {
	if isNil(prometheusRegisterer) {
		return nil
	}

	metric := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      name,
	})

	prometheusRegisterer.MustRegister(metric)

	return metric
}

// CounterVec creates and register a counter vector
func CounterVec(prometheusRegisterer prometheus.Registerer, namespace, subsystem, name string, labels ...string) *prometheus.CounterVec {
	if isNil(prometheusRegisterer) {
		return nil
	}

	metric := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      name,
	}, labels)

	prometheusRegisterer.MustRegister(metric)

	return metric
}
