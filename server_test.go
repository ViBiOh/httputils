package httputils

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestHttpGracefulClose(t *testing.T) {
	var tests = []struct {
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
			fmt.Errorf(`Error while shutting down http server: context deadline exceeded`),
		},
	}

	var failed bool

	for _, test := range tests {
		if test.server != nil {
			go test.server.ListenAndServe()
			defer test.server.Close()

			if _, err := GetBody(test.url, ``, false); err != nil {
				t.Errorf(`httpGracefulClose(%v), unable to fetch started server: %v`, test.server, err)
			}
		}

		if test.wait {
			go GetBody(test.url+`/long`, ``, false)
			time.Sleep(time.Second)
		}
		err := httpGracefulClose(test.server)

		if test.server != nil {
			if _, err := GetBody(test.url, ``, false); err == nil {
				t.Errorf(`httpGracefulClose(%v), still able to fetch data`, test.server)
			}
		}

		failed = false

		if err == nil && test.wantErr != nil {
			failed = true
		} else if err != nil && test.wantErr == nil {
			failed = true
		} else if err != nil && err.Error() != test.wantErr.Error() {
			failed = true
		}

		if failed {
			t.Errorf(`httpGracefulClose(%v) = %v, want %v`, test.server, err, test.wantErr)
		}
	}
}

func TestGracefulClose(t *testing.T) {
	var tests = []struct {
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
				return fmt.Errorf(`Error while shutting down`)
			},
			1,
		},
	}

	for _, test := range tests {
		if test.server != nil {
			go test.server.ListenAndServe()
			defer test.server.Close()
		}

		if test.wait {
			go GetBody(test.url+`/long`, ``, false)
			time.Sleep(time.Second)
		}

		if result := gracefulClose(test.server, test.callback); result != test.want {
			t.Errorf(`gracefulClose(%v) = %v, want %v`, test.server, result, test.want)
		}
	}
}
