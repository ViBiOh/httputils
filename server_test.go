package httputils

import (
	"net/http"
	"testing"
)

func TestServerGracefulClose(t *testing.T) {
	var tests = []struct {
		server *http.Server
	}{
		{
			nil,
		},
		{
			&http.Server{
				Addr: `:8000`,
				Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}),
			},
		},
	}

	for _, test := range tests {
		if test.server != nil {
			go test.server.ListenAndServe()

			if _, err := GetBody(`http://localhost:8000`, ``); err != nil {
				t.Errorf(`serverGracefulClose(%v), unable to fetch started server: %v`, test.server, err)
			}
		}

		serverGracefulClose(test.server)

		if test.server != nil {
			if _, err := GetBody(`http://localhost:8000`, ``); err == nil {
				t.Errorf(`serverGracefulClose(%v), still able to fetch data`, test.server)
			}
		}
	}
}
