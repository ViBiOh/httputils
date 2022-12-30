package prometheus

import (
	"flag"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/ViBiOh/httputils/v4/pkg/request"
	"github.com/prometheus/client_golang/prometheus"
)

func TestFlags(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		want string
	}{
		"simple": {
			"Usage of simple:\n  -gzip\n    \t[prometheus] Enable gzip compression of metrics output {SIMPLE_GZIP} (default true)\n  -ignore string\n    \t[prometheus] Ignored path prefixes for metrics, comma separated {SIMPLE_IGNORE}\n",
		},
	}

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			fs := flag.NewFlagSet(intention, flag.ContinueOnError)
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
	t.Parallel()

	metricsIgnore := ""
	metricsIgnoreValue := "/api"
	gzip := true

	cases := map[string]struct {
		instance App
		requests []*http.Request
		want     string
	}{
		"golang metrics": {
			New(Config{
				ignore: &metricsIgnore,
				gzip:   &gzip,
			}),
			[]*http.Request{
				httptest.NewRequest(http.MethodGet, "/", nil),
			},
			"go_gc_duration_seconds_count",
		},
		"http metrics": {
			New(Config{
				ignore: &metricsIgnore,
				gzip:   &gzip,
			}),
			[]*http.Request{
				httptest.NewRequest(http.MethodGet, "/", nil),
			},
			"http_requests_total{code=\"204\",method=\"GET\"} 1",
		},
		"http_requests_total": {
			New(Config{
				ignore: &metricsIgnoreValue,
				gzip:   &gzip,
			}),
			[]*http.Request{
				httptest.NewRequest(http.MethodPost, "/", nil),
				httptest.NewRequest(http.MethodPost, "/", nil),
				httptest.NewRequest(http.MethodPost, "/api", nil),
			},
			"http_requests_total{code=\"204\",method=\"POST\"} 2",
		},
	}

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			handler := testCase.instance.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			}))

			for _, req := range testCase.requests {
				handler.ServeHTTP(httptest.NewRecorder(), req)
			}

			writer := httptest.NewRecorder()
			testCase.instance.Handler().ServeHTTP(writer, httptest.NewRequest(http.MethodGet, metricsEndpoint, nil))

			if result, _ := request.ReadBodyResponse(writer.Result()); !strings.Contains(string(result), testCase.want) {
				t.Errorf("Middleware() = `%s`, want `%s`", string(result), testCase.want)
			}
		})
	}
}

func TestRegisterer(t *testing.T) {
	t.Parallel()

	registry := prometheus.NewRegistry()

	cases := map[string]struct {
		instance App
		want     prometheus.Registerer
	}{
		"default": {
			App{
				registry: registry,
			},
			registry,
		},
	}

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			if result := testCase.instance.Registerer(); !reflect.DeepEqual(result, testCase.want) {
				t.Errorf("Registerer() = %#v, want %#v", result, testCase.want)
			}
		})
	}
}

func TestIsIgnored(t *testing.T) {
	t.Parallel()

	type args struct {
		path string
	}

	cases := map[string]struct {
		instance App
		args     args
		want     bool
	}{
		"empty": {
			App{},
			args{
				path: metricsEndpoint,
			},
			false,
		},
		"multiple": {
			App{
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

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			if got := testCase.instance.isIgnored(testCase.args.path); got != testCase.want {
				t.Errorf("isIgnored() = %t, want %t", got, testCase.want)
			}
		})
	}
}

func BenchmarkMiddleware(b *testing.B) {
	app := App{
		registry: prometheus.NewRegistry(),
	}

	middleware := app.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	testRequest := httptest.NewRequest(http.MethodGet, "/", nil)
	recorder := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		middleware.ServeHTTP(recorder, testRequest)
	}
}

func BenchmarkHandler(b *testing.B) {
	ignore := ""
	gzip := false

	app := New(Config{
		ignore: &ignore,
		gzip:   &gzip,
	})

	handler := app.Handler()

	testRequest := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	recorder := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(recorder, testRequest)
	}
}
