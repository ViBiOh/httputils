package prometheus

import (
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
)

func TestGaugeVec(t *testing.T) {
	type args struct {
		registry  *prometheus.Registry
		namespace string
		subsystem string
		name      string
		labels    []string
	}

	var cases = []struct {
		intention string
		args      args
		want      string
	}{
		{
			"nil",
			args{},
			"",
		},
		{
			"simple",
			args{
				registry:  prometheus.NewRegistry(),
				namespace: "test",
				subsystem: "gaugevec",
				name:      "item",
				labels:    []string{"item"},
			},
			"# HELP test_gaugevec_item \n# TYPE test_gaugevec_item gauge\ntest_gaugevec_item{item=\"output\"} 1\n",
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			counter := GaugeVec(tc.args.registry, tc.args.namespace, tc.args.subsystem, tc.args.name, tc.args.labels...)

			var buffer strings.Builder

			if counter != nil {
				counter.WithLabelValues("output").Inc()

				metrics, err := tc.args.registry.Gather()
				if err != nil {
					t.Errorf("unable to gather metric: %s", err)
				}

				if len(metrics) == 0 {
					t.Error("no metric gathered")
				}

				expfmt.MetricFamilyToText(&buffer, metrics[0])
			}

			if got := buffer.String(); got != tc.want {
				t.Errorf("CounterVec() = `%s`, want `%s`", got, tc.want)
			}
		})
	}
}
