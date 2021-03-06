package health

import (
	"errors"
	"flag"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/request"
)

func TestFlags(t *testing.T) {
	var cases = []struct {
		intention string
		want      string
	}{
		{
			"simple",
			"Usage of simple:\n  -graceDuration string\n    \t[http] Grace duration when SIGTERM received {SIMPLE_GRACE_DURATION} (default \"30s\")\n  -okStatus int\n    \t[http] Healthy HTTP Status code {SIMPLE_OK_STATUS} (default 204)\n",
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

func TestHandler(t *testing.T) {
	okStatus := http.StatusNoContent
	graceDuration := "1s"
	closedChan := make(chan struct{})
	close(closedChan)

	var cases = []struct {
		intention  string
		instance   App
		request    *http.Request
		want       string
		wantStatus int
	}{
		{
			"wrong method",
			New(Config{
				okStatus:      &okStatus,
				graceDuration: &graceDuration,
			}),
			httptest.NewRequest(http.MethodHead, "/", nil),
			"",
			http.StatusMethodNotAllowed,
		},
		{
			"simple",
			New(Config{
				okStatus:      &okStatus,
				graceDuration: &graceDuration,
			}),
			httptest.NewRequest(http.MethodGet, "/", nil),
			"",
			okStatus,
		},
		{
			"shutdown",
			app{
				okStatus:      okStatus,
				graceDuration: time.Second,
				done:          closedChan,
			},
			httptest.NewRequest(http.MethodGet, "/", nil),
			"",
			http.StatusServiceUnavailable,
		},
		{
			"failing pinger",
			New(Config{
				okStatus:      &okStatus,
				graceDuration: &graceDuration,
			}, func() error {
				return errors.New("boom")
			}),
			httptest.NewRequest(http.MethodGet, "/ready", nil),
			"",
			http.StatusServiceUnavailable,
		},
		{
			"failing pinger on health",
			New(Config{
				okStatus:      &okStatus,
				graceDuration: &graceDuration,
			}, func() error {
				return errors.New("boom")
			}),
			httptest.NewRequest(http.MethodGet, "/health", nil),
			"",
			http.StatusNoContent,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			writer := httptest.NewRecorder()
			tc.instance.Handler().ServeHTTP(writer, tc.request)

			if got := writer.Code; got != tc.wantStatus {
				t.Errorf("Handler = %d, want %d", got, tc.wantStatus)
			}

			if got, _ := request.ReadBodyResponse(writer.Result()); string(got) != tc.want {
				t.Errorf("Handler = `%s`, want `%s`", string(got), tc.want)
			}
		})
	}
}
