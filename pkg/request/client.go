package request

import (
	"net"
	"net/http"
	"time"
)

// NoRedirection discard redirection
var NoRedirection = func(*http.Request, []*http.Request) error {
	return http.ErrUseLastResponse
}

// CreateClient creates http client with given timeout and redirection handling
func CreateClient(timeout time.Duration, onRedirect func(*http.Request, []*http.Request) error) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,

			DialContext: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 15 * time.Second,
			}).DialContext,

			TLSHandshakeTimeout:   5 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,

			MaxConnsPerHost:     100,
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     60 * time.Second,
		},

		Timeout: timeout,

		CheckRedirect: onRedirect,
	}
}
