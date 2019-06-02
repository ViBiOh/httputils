package request

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/ViBiOh/httputils/pkg/errors"
	"github.com/opentracing-contrib/go-stdlib/nethttp"
	opentracing "github.com/opentracing/opentracing-go"
)

var defaultHTTPClient = http.Client{
	Timeout:   30 * time.Second,
	Transport: &nethttp.Transport{},
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

// DoAndReadWithClient execute request and return output with given client
func DoAndReadWithClient(ctx context.Context, client http.Client, request *http.Request) (io.ReadCloser, int, http.Header, error) {
	if ctx != nil {
		var netTracer *nethttp.Tracer

		request = request.WithContext(ctx)
		request, netTracer = nethttp.TraceRequest(opentracing.GlobalTracer(), request)
		defer netTracer.Finish()
	}

	response, err := client.Do(request)
	if err != nil {
		if response != nil {
			if closeErr := response.Body.Close(); closeErr != nil {
				err = errors.New("%#v, and also %#v", err, closeErr)
			}
		}
		return nil, 0, nil, errors.WithStack(err)
	}

	if response.StatusCode >= http.StatusBadRequest {
		err = errors.New("error status %d", response.StatusCode)
	}

	return response.Body, response.StatusCode, response.Header, err
}

// DoAndRead execute request and return output
func DoAndRead(ctx context.Context, request *http.Request) (io.ReadCloser, int, http.Header, error) {
	return DoAndReadWithClient(ctx, defaultHTTPClient, request)
}
