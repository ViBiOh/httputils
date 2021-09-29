package prometheus

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

// Counters creates and registers counters
func Counters(prometheusRegisterer prometheus.Registerer, namespace, subsystem string, names ...string) (map[string]prometheus.Counter, error) {
	if prometheusRegisterer == nil {
		return nil, nil
	}

	metrics := make(map[string]prometheus.Counter)

	for _, name := range names {
		metric, err := createCounter(prometheusRegisterer, namespace, subsystem, name)
		if err != nil {
			return nil, err
		}

		metrics[name] = metric
	}

	return metrics, nil
}

func createCounter(prometheusRegisterer prometheus.Registerer, namespace, subsystem, name string) (prometheus.Counter, error) {
	counter := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      name,
	})

	if err := prometheusRegisterer.Register(counter); err != nil {
		return nil, fmt.Errorf("unable to register `%s` metric: %s", name, err)
	}

	return counter, nil
}
