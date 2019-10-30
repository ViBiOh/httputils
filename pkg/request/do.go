package request

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

var defaultHTTPClient = http.Client{
	Timeout: 30 * time.Second,
	CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

// DoWithClient execute request and return output with given client
func DoWithClient(ctx context.Context, client http.Client, req *http.Request) (body io.ReadCloser, status int, headers http.Header, err error) {
	response, doErr := client.Do(req)

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

// Do send given method with given content to URL with optional headers supplied
func Do(ctx context.Context, req *http.Request) (io.ReadCloser, int, http.Header, error) {
	return DoWithClient(ctx, defaultHTTPClient, req)
}

// Get send GET request to URL with optional headers supplied
func Get(ctx context.Context, url string, headers http.Header) (io.ReadCloser, int, http.Header, error) {
	req, err := New(ctx, http.MethodGet, url, nil, headers)
	if err != nil {
		return nil, 0, nil, err
	}

	return Do(ctx, req)
}

// Post send form via POST with urlencoded data
func Post(ctx context.Context, url string, data url.Values, headers http.Header) (io.ReadCloser, int, http.Header, error) {
	req, err := Form(ctx, http.MethodPost, url, data, headers)
	if err != nil {
		return nil, 0, nil, err
	}

	return Do(ctx, req)
}

// PostJSON send given method with given interface{} as JSON to URL with optional headers supplied
func PostJSON(ctx context.Context, url string, data interface{}, headers http.Header) (io.ReadCloser, int, http.Header, error) {
	req, err := JSON(ctx, http.MethodPost, url, data, headers)
	if err != nil {
		return nil, 0, nil, err
	}

	return Do(ctx, req)
}
