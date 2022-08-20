package httputils

import (
	"flag"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ViBiOh/httputils/v4/pkg/health"
	"github.com/ViBiOh/httputils/v4/pkg/request"
)

func TestHandler(t *testing.T) {
	healthApp := health.New(health.Flags(flag.NewFlagSet("TestHandler", flag.ContinueOnError), "httputils"))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("It works!")); err != nil {
			t.Error(err)
		}
	})

	t.Setenv("VERSION", "httputils/TestHandler")

	cases := map[string]struct {
		request    *http.Request
		want       string
		wantStatus int
	}{
		"simple": {
			httptest.NewRequest(http.MethodGet, "/", nil),
			"It works!",
			http.StatusOK,
		},
		"version": {
			httptest.NewRequest(http.MethodGet, "/version", nil),
			"httputils/TestHandler",
			http.StatusOK,
		},
		"health": {
			httptest.NewRequest(http.MethodGet, "/health", nil),
			"",
			http.StatusNoContent,
		},
		"ready": {
			httptest.NewRequest(http.MethodGet, "/ready", nil),
			"",
			http.StatusNoContent,
		},
	}

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			writer := httptest.NewRecorder()
			Handler(handler, healthApp).ServeHTTP(writer, testCase.request)

			if got := writer.Code; got != testCase.wantStatus {
				t.Errorf("Handler = %d, want %d", got, testCase.wantStatus)
			}

			if got, _ := request.ReadBodyResponse(writer.Result()); string(got) != testCase.want {
				t.Errorf("Handler = `%s`, want `%s`", string(got), testCase.want)
			}
		})
	}
}

func TestVersionHandler(t *testing.T) {
	cases := map[string]struct {
		request     *http.Request
		environment string
		want        string
		wantStatus  int
	}{
		"invalid method": {
			httptest.NewRequest(http.MethodOptions, "/", nil),
			"",
			"",
			http.StatusMethodNotAllowed,
		},
		"empty version": {
			httptest.NewRequest(http.MethodGet, "/", nil),
			"",
			"development",
			http.StatusOK,
		},
		"defined version": {
			httptest.NewRequest(http.MethodGet, "/", nil),
			"1234abcd",
			"1234abcd",
			http.StatusOK,
		},
	}

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
			t.Setenv("VERSION", testCase.environment)
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

func BenchmarkMux(b *testing.B) {
	fs := flag.NewFlagSet("BenchmarkMux", flag.ContinueOnError)

	healthApp := health.New(health.Flags(fs, "BenchmarkMux"))

	healthHandler := healthApp.HealthHandler()
	readyHandler := healthApp.ReadyHandler()
	versionHandler := versionHandler()
	var appHandler http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}

	mux := http.NewServeMux()
	mux.Handle(health.LivePath, healthHandler)
	mux.Handle(health.ReadyPath, readyHandler)
	mux.Handle("/version", versionHandler)
	mux.Handle("/", appHandler)

	testRequest := httptest.NewRequest(http.MethodGet, "/", nil)
	recorder := httptest.NewRecorder()

	for i := 0; i < b.N; i++ {
		mux.ServeHTTP(recorder, testRequest)
	}
}

func BenchmarkHandler(b *testing.B) {
	fs := flag.NewFlagSet("BenchmarkHandler", flag.ContinueOnError)

	healthConfig := health.Flags(fs, "BenchmarkHandler")

	var appHandler http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}

	handler := Handler(appHandler, health.New(healthConfig))

	testRequest := httptest.NewRequest(http.MethodGet, "/", nil)
	recorder := httptest.NewRecorder()

	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(recorder, testRequest)
	}
}
