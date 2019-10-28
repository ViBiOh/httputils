package request

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

var defaultHTTPClient = http.Client{
	Timeout: 30 * time.Second,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

// DoAndReadWithClient execute request and return output with given client
func DoAndReadWithClient(ctx context.Context, client http.Client, request *http.Request) (io.ReadCloser, int, http.Header, error) {
	response, err := client.Do(request)
	if err != nil {
		if payload, readErr := ReadBodyResponse(response); readErr != nil {
			err = fmt.Errorf("%s\n%s: %w", err, payload, readErr)
		}
		return nil, 0, nil, err
	}

	if response.StatusCode >= http.StatusBadRequest {
		err = fmt.Errorf("error status %d", response.StatusCode)
	}

	return response.Body, response.StatusCode, response.Header, err
}

// DoAndRead execute request and return output
func DoAndRead(ctx context.Context, request *http.Request) (io.ReadCloser, int, http.Header, error) {
	return DoAndReadWithClient(ctx, defaultHTTPClient, request)
}
