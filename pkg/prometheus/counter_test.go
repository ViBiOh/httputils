package prometheus

import (
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
)

func TestCounterVec(t *testing.T) {
	t.Parallel()

	type args struct {
		registry  *prometheus.Registry
		namespace string
		subsystem string
		name      string
		labels    []string
	}

	cases := map[string]struct {
		args args
		want string
	}{
		"nil": {
			args{},
			"",
		},
		"simple": {
			args{
				registry:  prometheus.NewRegistry(),
				namespace: "test",
				subsystem: "countervec",
				name:      "item",
				labels:    []string{"item"},
			},
			"# HELP test_countervec_item \n# TYPE test_countervec_item counter\ntest_countervec_item{item=\"output\"} 1\n",
		},
	}

	for intention, testCase := range cases {
		intention := intention
		testCase := testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			counter := CounterVec(testCase.args.registry, testCase.args.namespace, testCase.args.subsystem, testCase.args.name, testCase.args.labels...)

			var buffer strings.Builder

			if counter != nil {
				counter.WithLabelValues("output").Inc()

				metrics, err := testCase.args.registry.Gather()
				if err != nil {
					t.Errorf("gather metric: %s", err)
				}

				if len(metrics) == 0 {
					t.Error("no metric gathered")
				}

				if _, err = expfmt.MetricFamilyToText(&buffer, metrics[0]); err != nil {
					t.Errorf("format metric: %s", err)
				}
			}

			if got := buffer.String(); got != testCase.want {
				t.Errorf("CounterVec() = `%s`, want `%s`", got, testCase.want)
			}
		})
	}
}

func BenchmarkCounterVec(b *testing.B) {
	registry := prometheus.NewRegistry()
	counter := CounterVec(registry, "benchmark", "prometheus", "vector", "state")

	for i := 0; i < b.N; i++ {
		counter.WithLabelValues("valid").Inc()
	}
}

func BenchmarkCounter(b *testing.B) {
	registry := prometheus.NewRegistry()
	counter := Counter(registry, "benchmark", "prometheus", "counter")

	for i := 0; i < b.N; i++ {
		counter.Inc()
	}
}
