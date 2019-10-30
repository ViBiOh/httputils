package request

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	// ContentTypeHeader value
	ContentTypeHeader = "Content-Type"
)

func setHeader(headers http.Header, key, value string) http.Header {
	if headers == nil {
		headers = http.Header{}
	}

	headers.Set(key, value)

	return headers
}

// New prepare a request from given params
func New(ctx context.Context, method string, url string, body io.Reader, headers http.Header) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header = headers

	return req, nil
}

// Form prepare a Form request from given params
func Form(ctx context.Context, method string, url string, data url.Values, headers http.Header) (*http.Request, error) {
	return New(ctx, method, url, strings.NewReader(data.Encode()), setHeader(headers, ContentTypeHeader, "application/x-www-form-urlencoded"))
}

// JSON prepare a JSON request from given params
func JSON(ctx context.Context, method string, url string, body interface{}, headers http.Header) (*http.Request, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	return New(ctx, method, url, bytes.NewBuffer(jsonBody), setHeader(headers, ContentTypeHeader, "application/json"))
}
