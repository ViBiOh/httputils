package request

import (
	"net"
	"net/http"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/tracer"
	"go.opentelemetry.io/otel/trace"
)

var NoRedirection = func(*http.Request, []*http.Request) error {
	return http.ErrUseLastResponse
}

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

			MaxConnsPerHost:     512,
			MaxIdleConns:        256,
			MaxIdleConnsPerHost: 128,
			IdleConnTimeout:     60 * time.Second,
		},

		Timeout: timeout,

		CheckRedirect: onRedirect,
	}
}

func AddTracerToDefaultClient(tracerProvider trace.TracerProvider) {
	defaultHTTPClient = tracer.AddTracerToClient(defaultHTTPClient, tracerProvider)
}
