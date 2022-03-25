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
	cases := map[string]struct {
		want string
	}{
		"simple": {
			"Usage of simple:\n  -gzip\n    \t[prometheus] Enable gzip compression of metrics output {SIMPLE_GZIP} (default true)\n  -ignore string\n    \t[prometheus] Ignored path prefixes for metrics, comma separated {SIMPLE_IGNORE}\n",
		},
	}

	for intention, tc := range cases {
		t.Run(intention, func(t *testing.T) {
			fs := flag.NewFlagSet(intention, flag.ContinueOnError)
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
			"go_gc_duration_seconds_count 0",
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

	for intention, tc := range cases {
		t.Run(intention, func(t *testing.T) {
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

	for intention, tc := range cases {
		t.Run(intention, func(t *testing.T) {
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

	for intention, tc := range cases {
		t.Run(intention, func(t *testing.T) {
			if got := tc.instance.isIgnored(tc.args.path); got != tc.want {
				t.Errorf("isIgnored() = %t, want %t", got, tc.want)
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

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	recorder := httptest.NewRecorder()

	for i := 0; i < b.N; i++ {
		middleware.ServeHTTP(recorder, request)
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

	request := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	recorder := httptest.NewRecorder()

	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(recorder, request)
	}
}
