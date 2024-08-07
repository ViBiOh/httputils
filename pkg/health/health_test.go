package health

import (
	"context"
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
	t.Parallel()

	cases := map[string]struct {
		want string
	}{
		"simple": {
			"Usage of simple:\n  -graceDuration duration\n    \t[http] Grace duration when signal received ${SIMPLE_GRACE_DURATION} (default 30s)\n  -okStatus int\n    \t[http] Healthy HTTP Status code ${SIMPLE_OK_STATUS} (default 204)\n",
		},
	}

	for intention, testCase := range cases {
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

func TestHealthHandler(t *testing.T) {
	t.Parallel()

	closedChan := make(chan struct{})
	close(closedChan)

	cases := map[string]struct {
		instance   *Service
		request    *http.Request
		want       string
		wantStatus int
	}{
		"simple": {
			New(context.Background(), &Config{
				OkStatus:      http.StatusNoContent,
				GraceDuration: time.Second,
			}),
			httptest.NewRequest(http.MethodGet, "/", nil),
			"",
			http.StatusNoContent,
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			writer := httptest.NewRecorder()
			testCase.instance.HealthHandler().ServeHTTP(writer, testCase.request)

			if got := writer.Code; got != testCase.wantStatus {
				t.Errorf("Handler = %d, want %d", got, testCase.wantStatus)
			}

			if got, _ := request.ReadBodyResponse(writer.Result()); string(got) != testCase.want {
				t.Errorf("Handler = `%s`, want `%s`", string(got), testCase.want)
			}
		})
	}
}

func TestReadyHandler(t *testing.T) {
	t.Parallel()

	doneCtx, doneCancel := context.WithCancel(context.Background())
	doneCancel()

	cases := map[string]struct {
		instance   *Service
		request    *http.Request
		want       string
		wantStatus int
	}{
		"simple": {
			New(context.Background(), &Config{
				OkStatus:      http.StatusNoContent,
				GraceDuration: time.Second,
			}),
			httptest.NewRequest(http.MethodGet, "/", nil),
			"",
			http.StatusNoContent,
		},
		"shutdown": {
			&Service{
				okStatus:      http.StatusNoContent,
				graceDuration: time.Second,
				doneCtx:       doneCtx,
				doneCancel:    doneCancel,
			},
			httptest.NewRequest(http.MethodGet, "/", nil),
			"",
			http.StatusServiceUnavailable,
		},
		"failing pinger": {
			New(context.Background(), &Config{
				OkStatus:      http.StatusNoContent,
				GraceDuration: time.Second,
			}, func(_ context.Context) error {
				return errors.New("boom")
			}),
			httptest.NewRequest(http.MethodGet, "/ready", nil),
			"",
			http.StatusServiceUnavailable,
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			writer := httptest.NewRecorder()
			testCase.instance.ReadyHandler().ServeHTTP(writer, testCase.request)

			if got := writer.Code; got != testCase.wantStatus {
				t.Errorf("Handler = %d, want %d", got, testCase.wantStatus)
			}

			if got, _ := request.ReadBodyResponse(writer.Result()); string(got) != testCase.want {
				t.Errorf("Handler = `%s`, want `%s`", string(got), testCase.want)
			}
		})
	}
}
