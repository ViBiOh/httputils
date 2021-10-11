package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
)

// GaugeVec creates and register a gauge vector
func GaugeVec(prometheusRegisterer prometheus.Registerer, namespace, subsystem, name string, labels ...string) *prometheus.GaugeVec {
	if isNil(prometheusRegisterer) {
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
