package prometheus

import (
	"github.com/ViBiOh/httputils/v4/pkg/model"
	"github.com/prometheus/client_golang/prometheus"
)

func Counters(prometheusRegisterer prometheus.Registerer, namespace, subsystem string, names ...string) map[string]prometheus.Counter {
	if model.IsNil(prometheusRegisterer) {
		return nil
	}

	metrics := make(map[string]prometheus.Counter)

	for _, name := range names {
		metrics[name] = Counter(prometheusRegisterer, namespace, subsystem, name)
	}

	return metrics
}

func Counter(prometheusRegisterer prometheus.Registerer, namespace, subsystem, name string) prometheus.Counter {
	if model.IsNil(prometheusRegisterer) {
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

func CounterVec(prometheusRegisterer prometheus.Registerer, namespace, subsystem, name string, labels ...string) *prometheus.CounterVec {
	if model.IsNil(prometheusRegisterer) {
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
