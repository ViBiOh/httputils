package request

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

const (
	// ContentTypeHeader value
	ContentTypeHeader = `Content-Type`
)

// New prepare a request from given params
func New(method string, url string, body []byte, headers http.Header) (req *http.Request, err error) {
	req, err = http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf(`error while creating request: %v`, err)
	}
	req.Header = headers

	return
}

// JSON prepare a JSON request from given params
func JSON(method string, url string, body interface{}, headers http.Header) (*http.Request, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf(`error while marshalling body: %v`, err)
	}

	if headers == nil {
		headers = http.Header{}
	}
	headers.Set(ContentTypeHeader, `application/json`)

	return New(method, url, jsonBody, headers)
}

// Form prepare a Form request from given params
func Form(method string, url string, data url.Values, headers http.Header) (*http.Request, error) {
	if headers == nil {
		headers = http.Header{}
	}
	headers.Set(ContentTypeHeader, `application/x-www-form-urlencoded`)

	return New(method, url, []byte(data.Encode()), headers)
}

// Do send given method with given content to URL with optional headers supplied
func Do(ctx context.Context, method string, url string, body []byte, headers http.Header) ([]byte, error) {
	req, err := New(method, url, body, headers)
	if err != nil {
		return nil, err
	}

	return DoAndRead(ctx, req)
}

// DoJSON send given method with given interface{} as JSON to URL with optional headers supplied
func DoJSON(ctx context.Context, url string, data interface{}, headers http.Header, method string) ([]byte, error) {
	req, err := JSON(method, url, data, headers)
	if err != nil {
		return nil, err
	}

	return DoAndRead(ctx, req)
}

// PostForm send form via POST with urlencoded data
func PostForm(ctx context.Context, url string, data url.Values, headers http.Header) ([]byte, error) {
	req, err := Form(http.MethodPost, url, data, headers)
	if err != nil {
		return nil, err
	}

	return DoAndRead(ctx, req)
}

// Get send GET request to URL with optional headers supplied
func Get(ctx context.Context, url string, headers http.Header) ([]byte, error) {
	return Do(ctx, http.MethodGet, url, nil, headers)
}
