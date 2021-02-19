package httputils

import (
	"flag"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/ViBiOh/httputils/v3/pkg/health"
	"github.com/ViBiOh/httputils/v3/pkg/request"
)

func TestHandler(t *testing.T) {
	healthApp := health.New(health.Flags(flag.NewFlagSet("TestHandler", flag.ContinueOnError), "httputils"))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("It works!")); err != nil {
			t.Error(err)
		}
	})

	os.Setenv("VERSION", "httputils/TestHandler")

	var cases = []struct {
		intention  string
		request    *http.Request
		want       string
		wantStatus int
	}{
		{
			"simple",
			httptest.NewRequest(http.MethodGet, "/", nil),
			"It works!",
			http.StatusOK,
		},
		{
			"version",
			httptest.NewRequest(http.MethodGet, "/version", nil),
			"httputils/TestHandler",
			http.StatusOK,
		},
		{
			"health",
			httptest.NewRequest(http.MethodGet, "/health", nil),
			"",
			http.StatusNoContent,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			writer := httptest.NewRecorder()
			Handler(handler, healthApp).ServeHTTP(writer, tc.request)

			if got := writer.Code; got != tc.wantStatus {
				t.Errorf("Handler = %d, want %d", got, tc.wantStatus)
			}

			if got, _ := request.ReadBodyResponse(writer.Result()); string(got) != tc.want {
				t.Errorf("Handler = `%s`, want `%s`", string(got), tc.want)
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
