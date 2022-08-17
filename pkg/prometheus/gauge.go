package prometheus

import (
	"github.com/ViBiOh/httputils/v4/pkg/model"
	"github.com/prometheus/client_golang/prometheus"
)

// Gauges creates and registers gauges.
func Gauges(prometheusRegisterer prometheus.Registerer, namespace, subsystem string, names ...string) map[string]prometheus.Gauge {
	if model.IsNil(prometheusRegisterer) {
		return nil
	}

	metrics := make(map[string]prometheus.Gauge)

	for _, name := range names {
		metrics[name] = Gauge(prometheusRegisterer, namespace, subsystem, name)
	}

	return metrics
}

// Gauge creates and registers a gauge.
func Gauge(prometheusRegisterer prometheus.Registerer, namespace, subsystem, name string) prometheus.Gauge {
	if model.IsNil(prometheusRegisterer) {
		return nil
	}

	metric := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      name,
	})

	prometheusRegisterer.MustRegister(metric)

	return metric
}

// GaugeVec creates and register a gauge vector.
func GaugeVec(prometheusRegisterer prometheus.Registerer, namespace, subsystem, name string, labels ...string) *prometheus.GaugeVec {
	if model.IsNil(prometheusRegisterer) {
		return nil
	}

	metric := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      name,
	}, labels)

	prometheusRegisterer.MustRegister(metric)

	return metric
}
