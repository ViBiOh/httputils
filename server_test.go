package httputils

import (
	"errors"
	"net/http"
	"testing"
	"time"
)

func TestHttpGracefulClose(t *testing.T) {
	var cases = []struct {
		url     string
		server  *http.Server
		wait    bool
		wantErr error
	}{
		{
			``,
			nil,
			false,
			nil,
		},
		{
			`http://localhost:8000`,
			&http.Server{
				Addr: `:8000`,
				Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}),
			},
			false,
			nil,
		},
		{
			`http://localhost:8001`,
			&http.Server{
				Addr: `:8001`,
				Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path == `/long` {
						time.Sleep(time.Second * 30)
					}
					w.WriteHeader(http.StatusOK)
				}),
			},
			true,
			errors.New(`Error while shutting down http server: context deadline exceeded`),
		},
	}

	var failed bool

	for _, testCase := range cases {
		if testCase.server != nil {
			go testCase.server.ListenAndServe()
			defer testCase.server.Close()

			if _, err := GetRequest(testCase.url, nil); err != nil {
				t.Errorf(`httpGracefulClose(%v), unable to fetch started server: %v`, testCase.server, err)
			}
		}

		if testCase.wait {
			go GetRequest(testCase.url+`/long`, nil)
			time.Sleep(time.Second)
		}
		err := httpGracefulClose(testCase.server)

		if testCase.server != nil {
			if _, err := GetRequest(testCase.url, nil); err == nil {
				t.Errorf(`httpGracefulClose(%v), still able to fetch data`, testCase.server)
			}
		}

		failed = false

		if err == nil && testCase.wantErr != nil {
			failed = true
		} else if err != nil && testCase.wantErr == nil {
			failed = true
		} else if err != nil && err.Error() != testCase.wantErr.Error() {
			failed = true
		}

		if failed {
			t.Errorf(`httpGracefulClose(%v) = %v, want %v`, testCase.server, err, testCase.wantErr)
		}
	}
}

func TestGracefulClose(t *testing.T) {
	var cases = []struct {
		url      string
		server   *http.Server
		wait     bool
		callback func() error
		want     int
	}{
		{
			``,
			nil,
			false,
			nil,
			0,
		},
		{
			`http://localhost:8100`,
			&http.Server{
				Addr: `:8100`,
				Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}),
			},
			false,
			func() error {
				return nil
			},
			0,
		},
		{
			`http://localhost:8101`,
			&http.Server{
				Addr: `:8101`,
				Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path == `/long` {
						time.Sleep(time.Second * 30)
					}
					w.WriteHeader(http.StatusOK)
				}),
			},
			true,
			nil,
			1,
		},
		{
			``,
			nil,
			false,
			func() error {
				return errors.New(`Error while shutting down`)
			},
			1,
		},
	}

	for _, testCase := range cases {
		if testCase.server != nil {
			go testCase.server.ListenAndServe()
			defer testCase.server.Close()
		}

		if testCase.wait {
			go GetRequest(testCase.url+`/long`, nil)
			time.Sleep(time.Second)
		}

		if result := gracefulClose(testCase.server, testCase.callback); result != testCase.want {
			t.Errorf(`gracefulClose(%v) = %v, want %v`, testCase.server, result, testCase.want)
		}
	}
}
