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
			"Usage of simple:\n  -path string\n    \t[prometheus] Path for exposing metrics {SIMPLE_PATH} (default \"/metrics\")\n",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			fs := flag.NewFlagSet(testCase.intention, flag.ContinueOnError)
			Flags(fs, "")

			var writer strings.Builder
			fs.SetOutput(&writer)
			fs.Usage()

			result := writer.String()

			if result != testCase.want {
				t.Errorf("Flags() = `%s`, want `%s`", result, testCase.want)
			}
		})
	}
}

func TestMiddleware(t *testing.T) {
	metricsPath := "/metrics"

	var cases = []struct {
		intention string
		instance  App
		requests  []*http.Request
		want      string
	}{
		{
			"golang metrics",
			New(Config{
				path: &metricsPath,
			}),
			[]*http.Request{
				httptest.NewRequest(http.MethodGet, "/", nil),
			},
			"go_gc_duration_seconds_count 0",
		},
		{
			"http metrics",
			New(Config{
				path: &metricsPath,
			}),
			[]*http.Request{
				httptest.NewRequest(http.MethodGet, "/", nil),
			},
			"http_requests_total{code=\"204\",method=\"get\"} 1",
		},
		{
			"http_requests_total",
			New(Config{
				path: &metricsPath,
			}),
			[]*http.Request{
				httptest.NewRequest(http.MethodPost, "/", nil),
				httptest.NewRequest(http.MethodPost, "/", nil),
				httptest.NewRequest(http.MethodPost, "/", nil),
			},
			"http_requests_total{code=\"204\",method=\"post\"} 3",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			handler := testCase.instance.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			}))

			for _, req := range testCase.requests {
				handler.ServeHTTP(httptest.NewRecorder(), req)
			}

			writer := httptest.NewRecorder()
			handler.ServeHTTP(writer, httptest.NewRequest(http.MethodGet, "/metrics", nil))

			if result, _ := request.ReadBodyResponse(writer.Result()); !strings.Contains(string(result), testCase.want) {
				t.Errorf("Middleware() = `%s`, want `%s`", string(result), testCase.want)
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

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			if result := testCase.instance.Registerer(); !reflect.DeepEqual(result, testCase.want) {
				t.Errorf("Registerer() = %#v, want %#v", result, testCase.want)
			}
		})
	}
}

func TestSwagger(t *testing.T) {
	var cases = []struct {
		intention string
		instance  app
		want      string
	}{
		{
			"simple",
			app{
				path: "/metrics",
			},
			`/metrics:
  get:
    description: Retrieves metrics of app

    responses:
      200:
        description: Metrics of app
        content:
          text/plain:
            schema:
              type: string`,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			if result, _ := testCase.instance.Swagger(); result.Paths != testCase.want {
				t.Errorf("Swagger() = `%s`, want `%s`", result.Paths, testCase.want)
			}
		})
	}
}
