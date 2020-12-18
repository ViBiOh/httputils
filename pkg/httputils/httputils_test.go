package httputils

import (
	"flag"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ViBiOh/httputils/v3/pkg/model"
	"github.com/ViBiOh/httputils/v3/pkg/request"
)

func TestFlags(t *testing.T) {
	var cases = []struct {
		intention string
		want      string
	}{
		{
			"simple",
			"Usage of simple:\n  -address string\n    \t[http] Listen address {SIMPLE_ADDRESS}\n  -cert string\n    \t[http] Certificate file {SIMPLE_CERT}\n  -graceDuration string\n    \t[http] Grace duration when SIGTERM received {SIMPLE_GRACE_DURATION} (default \"30s\")\n  -idleTimeout string\n    \t[http] Idle Timeout {SIMPLE_IDLE_TIMEOUT} (default \"2m\")\n  -key string\n    \t[http] Key file {SIMPLE_KEY}\n  -okStatus int\n    \t[http] Healthy HTTP Status code {SIMPLE_OK_STATUS} (default 204)\n  -port uint\n    \t[http] Listen port {SIMPLE_PORT} (default 1080)\n  -readTimeout string\n    \t[http] Read Timeout {SIMPLE_READ_TIMEOUT} (default \"5s\")\n  -shutdownTimeout string\n    \t[http] Shutdown Timeout {SIMPLE_SHUTDOWN_TIMEOUT} (default \"10s\")\n  -writeTimeout string\n    \t[http] Write Timeout {SIMPLE_WRITE_TIMEOUT} (default \"10s\")\n",
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

func TestVersionHandler(t *testing.T) {
	var cases = []struct {
		intention   string
		request     *http.Request
		environment string
		want        string
		wantStatus  int
	}{
		{
			"invalid method",
			httptest.NewRequest(http.MethodOptions, "/", nil),
			"",
			"",
			http.StatusMethodNotAllowed,
		},
		{
			"empty version",
			httptest.NewRequest(http.MethodGet, "/", nil),
			"",
			"development",
			http.StatusOK,
		},
		{
			"defined version",
			httptest.NewRequest(http.MethodGet, "/", nil),
			"1234abcd",
			"1234abcd",
			http.StatusOK,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			os.Setenv("VERSION", testCase.environment)
			writer := httptest.NewRecorder()
			versionHandler().ServeHTTP(writer, testCase.request)

			if result := writer.Code; result != testCase.wantStatus {
				t.Errorf("VersionHandler = %d, want %d", result, testCase.wantStatus)
			}

			if result, _ := request.ReadBodyResponse(writer.Result()); string(result) != testCase.want {
				t.Errorf("VersionHandler = `%s`, want `%s`", string(result), testCase.want)
			}
		})
	}
}

func TestChainMiddlewares(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("handler"))
	})

	middlewareNotFound := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			next.ServeHTTP(w, r)
		})
	}

	var cases = []struct {
		intention   string
		request     *http.Request
		middlewares []model.Middleware
		want        string
		wantStatus  int
		wantHeader  http.Header
	}{
		{
			"nil chain",
			httptest.NewRequest(http.MethodGet, "/", nil),
			nil,
			"handler",
			http.StatusOK,
			http.Header{},
		},
		{
			"values",
			httptest.NewRequest(http.MethodGet, "/", nil),
			[]model.Middleware{middlewareNotFound},
			"handler",
			http.StatusNotFound,
			http.Header{},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			writer := httptest.NewRecorder()
			ChainMiddlewares(handler, testCase.middlewares...).ServeHTTP(writer, testCase.request)

			if result := writer.Code; result != testCase.wantStatus {
				t.Errorf("ChainMiddlewares = %d, want %d", result, testCase.wantStatus)
			}

			if result, _ := request.ReadBodyResponse(writer.Result()); string(result) != testCase.want {
				t.Errorf("ChainMiddlewares = `%s`, want `%s`", string(result), testCase.want)
			}

			for key := range testCase.wantHeader {
				want := testCase.wantHeader.Get(key)
				if result := writer.Header().Get(key); result != want {
					t.Errorf("`%s` Header = `%s`, want `%s`", key, result, want)
				}
			}
		})
	}
}

func TestSafeParseDuration(t *testing.T) {
	type args struct {
		name            string
		value           string
		defaultDuration time.Duration
	}

	var cases = []struct {
		intention string
		args      args
		want      time.Duration
	}{
		{
			"default",
			args{
				name:            "test",
				value:           "abcd",
				defaultDuration: time.Minute,
			},
			time.Minute,
		},
		{
			"parsed",
			args{
				name:            "test",
				value:           "5m",
				defaultDuration: time.Minute,
			},
			time.Minute * 5,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			if got := safeParseDuration(tc.args.name, tc.args.value, tc.args.defaultDuration); got != tc.want {
				t.Errorf("safeParseDuration() = %s, want %s", got, tc.want)
			}
		})
	}
}
