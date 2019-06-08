package server

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/ViBiOh/httputils/pkg/healthcheck"
	"github.com/ViBiOh/httputils/pkg/request"
)

func TestHttpGracefulClose(t *testing.T) {
	var cases = []struct {
		intention string
		url       string
		server    *http.Server
		wait      bool
		wantErr   error
	}{
		{
			"nothing if no server",
			"",
			nil,
			false,
			nil,
		},
		{
			"shutdown http quickly",
			"http://localhost:8000",
			&http.Server{
				Addr: ":8000",
				Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}),
			},
			false,
			nil,
		},
		{
			"http://localhost:8001",
			"http://localhost:8001",
			&http.Server{
				Addr: ":8001",
				Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path == "/long_request" {
						time.Sleep(time.Second * 30)
					}
					w.WriteHeader(http.StatusOK)
				}),
			},
			true,
			errors.New("context deadline exceeded"),
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			if testCase.server != nil {
				go testCase.server.ListenAndServe()

				if _, _, _, err := request.Get(nil, testCase.url, nil); err != nil {
					t.Errorf("unable to fetch server: %#v", err)
				}
			}

			if testCase.wait {
				go request.Get(nil, fmt.Sprintf("%s/long_request", testCase.url), nil)
				time.Sleep(time.Second)
			}

			err := httpGracefulClose(testCase.server)

			if testCase.server != nil {
				if _, _, _, err := request.Get(nil, testCase.url, nil); err == nil {
					t.Errorf("still able to fetch server")
				}
			}

			failed := false

			if err == nil && testCase.wantErr != nil {
				failed = true
			} else if err != nil && testCase.wantErr == nil {
				failed = true
			} else if err != nil && err.Error() != testCase.wantErr.Error() {
				failed = true
			}

			if failed {
				t.Errorf("httpGracefulClose(%#v) = %#v, want %#v", testCase.server, err, testCase.wantErr)
			}

			if testCase.server != nil {
				testCase.server.Close()
			}
		})
	}
}

func TestGracefulClose(t *testing.T) {
	var cases = []struct {
		intention        string
		url              string
		server           *http.Server
		gracefulDuration time.Duration
		wait             bool
		healthcheckApp   *healthcheck.App
		want             int
	}{
		{
			"nothing if no server",
			"",
			nil,
			0,
			false,
			nil,
			0,
		},
		{
			"nominal case",
			"http://localhost:8100",
			&http.Server{
				Addr: ":8100",
				Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}),
			},
			time.Second,
			false,
			healthcheck.New(),
			0,
		},
		{
			"fail on long request",
			"http://localhost:8101",
			&http.Server{
				Addr: ":8101",
				Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path == "/long_request" {
						time.Sleep(time.Second * 30)
					}
					w.WriteHeader(http.StatusOK)
				}),
			},
			time.Second * 2,
			true,
			nil,
			1,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			if testCase.server != nil {
				go testCase.server.ListenAndServe()
			}

			if testCase.wait {
				go request.Get(nil, fmt.Sprintf("%s/long_request", testCase.url), nil)
				time.Sleep(time.Second)
			}

			if result := gracefulClose(testCase.server, testCase.gracefulDuration, testCase.healthcheckApp); result != testCase.want {
				t.Errorf("gracefulClose() = %d, want %d", result, testCase.want)
			}

			if testCase.server != nil {
				testCase.server.Close()
			}
		})
	}
}
