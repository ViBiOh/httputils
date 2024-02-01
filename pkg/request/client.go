package request

import (
	"net"
	"net/http"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/telemetry"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

var NoRedirection = func(*http.Request, []*http.Request) error {
	return http.ErrUseLastResponse
}

var defaultTransport = &http.Transport{
	Proxy: http.ProxyFromEnvironment,

	DialContext: (&net.Dialer{
		Timeout:   time.Second * 5,
		KeepAlive: time.Second * 15,
	}).DialContext,

	TLSHandshakeTimeout:   time.Second * 5,
	ExpectContinueTimeout: time.Second * 1,

	MaxConnsPerHost:     512,
	MaxIdleConns:        256,
	MaxIdleConnsPerHost: 128,
	IdleConnTimeout:     time.Second * 60,
}

func CreateClient(timeout time.Duration, onRedirect func(*http.Request, []*http.Request) error) *http.Client {
	return CreateClientWithTransport(timeout, onRedirect, defaultTransport)
}

func CreateClientWithTransport(timeout time.Duration, onRedirect func(*http.Request, []*http.Request) error, transport http.RoundTripper) *http.Client {
	return &http.Client{
		Transport:     transport,
		Timeout:       timeout,
		CheckRedirect: onRedirect,
	}
}

func AddOpenTelemetryToDefaultClient(meterProvider metric.MeterProvider, tracerProvider trace.TracerProvider) {
	defaultHTTPClient = telemetry.AddOpenTelemetryToClient(defaultHTTPClient, meterProvider, tracerProvider)
}
