package prometheus

import (
	"flag"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/ViBiOh/httputils/v3/pkg/request"
	"github.com/prometheus/client_golang/prometheus"
)

func TestFlags(t *testing.T) {
	var cases = []struct {
		intention string
		want      string
	}{
		{
			"simple",
			"Usage of simple:\n  -ignore string\n    \t[prometheus] Ignored path prefixes for metrics, comma separated {SIMPLE_IGNORE}\n",
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			fs := flag.NewFlagSet(tc.intention, flag.ContinueOnError)
			Flags(fs, "")

			var writer strings.Builder
			fs.SetOutput(&writer)
			fs.Usage()

			result := writer.String()

			if result != tc.want {
				t.Errorf("Flags() = `%s`, want `%s`", result, tc.want)
			}
		})
	}
}

func TestMiddleware(t *testing.T) {
	metricsIgnore := ""
	metricsIgnoreValue := "/api"

	var cases = []struct {
		intention string
		instance  App
		requests  []*http.Request
		want      string
	}{
		{
			"golang metrics",
			New(Config{
				ignore: &metricsIgnore,
			}),
			[]*http.Request{
				httptest.NewRequest(http.MethodGet, "/", nil),
			},
			"go_gc_duration_seconds_count 0",
		},
		{
			"http metrics",
			New(Config{
				ignore: &metricsIgnore,
			}),
			[]*http.Request{
				httptest.NewRequest(http.MethodGet, "/", nil),
			},
			"http_requests_total{code=\"204\",method=\"get\"} 1",
		},
		{
			"http_requests_total",
			New(Config{
				ignore: &metricsIgnoreValue,
			}),
			[]*http.Request{
				httptest.NewRequest(http.MethodPost, "/", nil),
				httptest.NewRequest(http.MethodPost, "/", nil),
				httptest.NewRequest(http.MethodPost, "/api", nil),
			},
			"http_requests_total{code=\"204\",method=\"post\"} 2",
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			handler := tc.instance.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			}))

			for _, req := range tc.requests {
				handler.ServeHTTP(httptest.NewRecorder(), req)
			}

			writer := httptest.NewRecorder()
			tc.instance.Handler().ServeHTTP(writer, httptest.NewRequest(http.MethodGet, metricsEndpoint, nil))

			if result, _ := request.ReadBodyResponse(writer.Result()); !strings.Contains(string(result), tc.want) {
				t.Errorf("Middleware() = `%s`, want `%s`", string(result), tc.want)
			}
		})
	}
}

func TestRegisterer(t *testing.T) {
	registry := prometheus.NewRegistry()

	var cases = []struct {
		intention string
		instance  app
		want      prometheus.Registerer
	}{
		{
			"default",
			app{
				registry: registry,
			},
			registry,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			if result := tc.instance.Registerer(); !reflect.DeepEqual(result, tc.want) {
				t.Errorf("Registerer() = %#v, want %#v", result, tc.want)
			}
		})
	}
}

func TestIsIgnored(t *testing.T) {
	type args struct {
		path string
	}

	var cases = []struct {
		intention string
		instance  app
		args      args
		want      bool
	}{
		{
			"empty",
			app{},
			args{
				path: metricsEndpoint,
			},
			false,
		},
		{
			"multiple",
			app{
				ignore: []string{
					metricsEndpoint,
					"/api",
				},
			},
			args{
				path: "/api/users/1",
			},
			true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			if got := tc.instance.isIgnored(tc.args.path); got != tc.want {
				t.Errorf("isIgnored() = %t, want %t", got, tc.want)
			}
		})
	}
}
