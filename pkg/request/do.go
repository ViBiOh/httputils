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
	CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

// DoAndReadWithClient execute request and return output with given client
func DoAndReadWithClient(ctx context.Context, client http.Client, request *http.Request) (body io.ReadCloser, status int, headers http.Header, err error) {
	response, doErr := client.Do(request)

	if response != nil {
		body = response.Body
		status = response.StatusCode
		headers = response.Header
	}

	if doErr != nil {
		err = doErr
	}

	if err != nil || status >= http.StatusBadRequest {
		if err == nil {
			err = fmt.Errorf("HTTP/%d", status)
		}

		if payload, readErr := ReadBodyResponse(response); readErr == nil && len(payload) > 0 {
			err = fmt.Errorf("%s\n%s", err, payload)
		}
	}

	return
}

// DoAndRead execute request and return output
func DoAndRead(ctx context.Context, request *http.Request) (io.ReadCloser, int, http.Header, error) {
	return DoAndReadWithClient(ctx, defaultHTTPClient, request)
}
