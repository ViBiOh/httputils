package request

import (
	"context"
	"fmt"
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
func DoWithClient(ctx context.Context, client http.Client, req *http.Request) (*http.Response, error) {
	resp, err := client.Do(req)
	if err != nil || (resp != nil && resp.StatusCode >= http.StatusBadRequest) {
		if err == nil {
			err = fmt.Errorf("HTTP/%d", resp.StatusCode)
		}

		if payload, readErr := ReadBodyResponse(resp); readErr == nil && len(payload) > 0 {
			err = fmt.Errorf("%s\n%s", err, payload)
		}
	}

	return resp, err
}

// Do send given method with given content to URL with optional headers supplied
func Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	return DoWithClient(ctx, defaultHTTPClient, req)
}

// Get send GET request to URL with optional headers supplied
func Get(ctx context.Context, url string, headers http.Header) (*http.Response, error) {
	req, err := New(ctx, http.MethodGet, url, nil, headers)
	if err != nil {
		return nil, err
	}

	return Do(ctx, req)
}

// Post send form via POST with urlencoded data
func Post(ctx context.Context, url string, data url.Values, headers http.Header) (*http.Response, error) {
	req, err := Form(ctx, http.MethodPost, url, data, headers)
	if err != nil {
		return nil, err
	}

	return Do(ctx, req)
}

// PostJSON send given method with given interface{} as JSON to URL with optional headers supplied
func PostJSON(ctx context.Context, url string, data interface{}, headers http.Header) (*http.Response, error) {
	req, err := JSON(ctx, http.MethodPost, url, data, headers)
	if err != nil {
		return nil, err
	}

	return Do(ctx, req)
}
