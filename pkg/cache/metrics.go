package cache

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	cacheNamespace = "cache"
)

func createMetrics(prometheusRegisterer prometheus.Registerer, names ...string) (map[string]prometheus.Counter, error) {
	if prometheusRegisterer == nil {
		return nil, nil
	}

	metrics := make(map[string]prometheus.Counter)

	for _, name := range names {
		metric, err := createMetric(prometheusRegisterer, name)
		if err != nil {
			return nil, err
		}

		metrics[name] = metric
	}

	return metrics, nil
}

func createMetric(prometheusRegisterer prometheus.Registerer, name string) (prometheus.Counter, error) {
	counter := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: cacheNamespace,
		Name:      name,
	})

	if err := prometheusRegisterer.Register(counter); err != nil {
		return nil, fmt.Errorf("unable to register `%s` metric: %s", name, err)
	}

	return counter, nil
}

func (a *App) increase(name string) {
	if gauge, ok := a.metrics[name]; ok {
		gauge.Inc()
	}
}
