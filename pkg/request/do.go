package request

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/opentracing-contrib/go-stdlib/nethttp"
	opentracing "github.com/opentracing/opentracing-go"
)

var defaultHTTPClient = http.Client{
	Timeout:   30 * time.Second,
	Transport: &nethttp.Transport{},
}

// DoAndReadWithClient execute request and return output with given client
func DoAndReadWithClient(ctx context.Context, client http.Client, request *http.Request) ([]byte, error) {
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
				err = fmt.Errorf(`, and also error while closing body: %v`, closeErr)
			}
		}
		return nil, fmt.Errorf(`error while processing request: %v`, err)
	}

	responseBody, err := ReadBodyResponse(response)
	if err != nil {
		return nil, fmt.Errorf(`error while reading body: %v`, err)
	}

	if response.StatusCode >= http.StatusBadRequest {
		return responseBody, fmt.Errorf(`error status %d`, response.StatusCode)
	}

	return responseBody, nil
}

// DoAndRead execute request and return output
func DoAndRead(ctx context.Context, request *http.Request) ([]byte, error) {
	return DoAndReadWithClient(ctx, defaultHTTPClient, request)
}
