package prometheus

import (
	"flag"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ViBiOh/httputils/v3/pkg/request"
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
				t.Errorf("Flags() = %s, want %s", result, testCase.want)
			}
		})
	}
}

func TestHandler(t *testing.T) {
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
			handler := testCase.instance.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			}))

			for _, req := range testCase.requests {
				handler.ServeHTTP(httptest.NewRecorder(), req)
			}

			writer := httptest.NewRecorder()
			handler.ServeHTTP(writer, httptest.NewRequest(http.MethodGet, "/metrics", nil))

			if result, _ := request.ReadBodyResponse(writer.Result()); !strings.Contains(string(result), testCase.want) {
				t.Errorf("Handler() = `%s`, want `%s`", string(result), testCase.want)
			}
		})
	}
}
