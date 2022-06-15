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
	cases := map[string]struct {
		want string
	}{
		"simple": {
			"Usage of simple:\n  -graceDuration duration\n    \t[http] Grace duration when SIGTERM received {SIMPLE_GRACE_DURATION} (default 30s)\n  -okStatus int\n    \t[http] Healthy HTTP Status code {SIMPLE_OK_STATUS} (default 204)\n",
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

func TestHealthHandler(t *testing.T) {
	okStatus := http.StatusNoContent
	graceDuration := time.Second
	closedChan := make(chan struct{})
	close(closedChan)

	cases := map[string]struct {
		instance   App
		request    *http.Request
		want       string
		wantStatus int
	}{
		"simple": {
			New(Config{
				okStatus:      &okStatus,
				graceDuration: &graceDuration,
			}),
			httptest.NewRequest(http.MethodGet, "/", nil),
			"",
			okStatus,
		},
	}

	for intention, tc := range cases {
		t.Run(intention, func(t *testing.T) {
			writer := httptest.NewRecorder()
			tc.instance.HealthHandler().ServeHTTP(writer, tc.request)

			if got := writer.Code; got != tc.wantStatus {
				t.Errorf("Handler = %d, want %d", got, tc.wantStatus)
			}

			if got, _ := request.ReadBodyResponse(writer.Result()); string(got) != tc.want {
				t.Errorf("Handler = `%s`, want `%s`", string(got), tc.want)
			}
		})
	}
}

func TestReadyHandler(t *testing.T) {
	okStatus := http.StatusNoContent
	graceDuration := time.Second
	closedChan := make(chan struct{})
	close(closedChan)

	cases := map[string]struct {
		instance   App
		request    *http.Request
		want       string
		wantStatus int
	}{
		"simple": {
			New(Config{
				okStatus:      &okStatus,
				graceDuration: &graceDuration,
			}),
			httptest.NewRequest(http.MethodGet, "/", nil),
			"",
			okStatus,
		},
		"shutdown": {
			App{
				okStatus:      okStatus,
				graceDuration: time.Second,
				done:          closedChan,
			},
			httptest.NewRequest(http.MethodGet, "/", nil),
			"",
			http.StatusServiceUnavailable,
		},
		"failing pinger": {
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
	}

	for intention, tc := range cases {
		t.Run(intention, func(t *testing.T) {
			writer := httptest.NewRecorder()
			tc.instance.ReadyHandler().ServeHTTP(writer, tc.request)

			if got := writer.Code; got != tc.wantStatus {
				t.Errorf("Handler = %d, want %d", got, tc.wantStatus)
			}

			if got, _ := request.ReadBodyResponse(writer.Result()); string(got) != tc.want {
				t.Errorf("Handler = `%s`, want `%s`", string(got), tc.want)
			}
		})
	}
}
