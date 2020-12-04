package request

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ViBiOh/httputils/v3/pkg/logger"
)

var (
	defaultHTTPClient = http.Client{
		Timeout: 15 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	defaultRetryCount uint = 3
)

// Request describe a complete request
type Request struct {
	method   string
	url      string
	username string
	password string

	header http.Header
	client http.Client

	retry bool
}

// New create a new Request
func New() *Request {
	return &Request{
		retry:  true,
		method: http.MethodGet,
		header: http.Header{},
		client: defaultHTTPClient,
	}
}

// NoRetry deactivates retry on error
func (r *Request) NoRetry() *Request {
	r.retry = false

	return r
}

// Method set method of Request
func (r *Request) Method(method string) *Request {
	r.method = method

	return r
}

// URL set URL of Request
func (r *Request) URL(url string) *Request {
	r.url = url

	return r
}

// Get set GET to given url
func (r *Request) Get(url string) *Request {
	return r.Method(http.MethodGet).URL(url)
}

// Post set POST to given url
func (r *Request) Post(url string) *Request {
	return r.Method(http.MethodPost).URL(url)
}

// Put set PUT to given url
func (r *Request) Put(url string) *Request {
	return r.Method(http.MethodPut).URL(url)
}

// Patch set PATCH to given url
func (r *Request) Patch(url string) *Request {
	return r.Method(http.MethodPatch).URL(url)
}

// Delete set DELETE to given url
func (r *Request) Delete(url string) *Request {
	return r.Method(http.MethodDelete).URL(url)
}

// BasicAuth add Basic Auth header
func (r *Request) BasicAuth(username, password string) *Request {
	r.username = username
	r.password = password

	return r
}

// Header add header to request
func (r *Request) Header(name, value string) *Request {
	r.header.Set(name, value)

	return r
}

// ContentType set Content-Type header
func (r *Request) ContentType(contentType string) *Request {
	return r.Header("Content-Type", contentType)
}

// ContentForm set Content-Type header to application/x-www-form-urlencoded
func (r *Request) ContentForm() *Request {
	return r.ContentType("application/x-www-form-urlencoded")
}

// ContentJSON set Content-Type header to application/json
func (r *Request) ContentJSON() *Request {
	return r.ContentType("application/json")
}

// WithClient defines net/http client to use, instead of default one (30sec timeout and no redirect)
func (r *Request) WithClient(client http.Client) *Request {
	r.client = client

	return r
}

// Build create request for given context and payload
func (r *Request) Build(ctx context.Context, payload io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, r.method, r.url, payload)
	if err != nil {
		return nil, err
	}

	req.Header = r.header
	if len(r.username) != 0 || len(r.password) != 0 {
		req.SetBasicAuth(r.username, r.password)
	}

	return req, nil
}

// Send build request and send it with defined client
func (r *Request) Send(ctx context.Context, payload io.Reader) (*http.Response, error) {
	req, err := r.Build(ctx, payload)
	if err != nil {
		return nil, err
	}

	return DoWithClientAndRetry(r.client, req, defaultRetryCount)
}

// Form send request with given context and url.Values as payload
func (r *Request) Form(ctx context.Context, data url.Values) (*http.Response, error) {
	return r.ContentForm().Send(ctx, strings.NewReader(data.Encode()))
}

// JSON send request with given context and given interface as JSON payload
func (r *Request) JSON(ctx context.Context, body interface{}) (*http.Response, error) {
	reader, writer := io.Pipe()

	go func() {
		if err := json.NewEncoder(writer).Encode(body); err != nil {
			logger.Error("unable to send json: %s", err)
		}

		if err := writer.Close(); err != nil {
			logger.Error("unable to close json writer: %s", err)
		}
	}()

	return r.ContentJSON().Send(ctx, reader)
}

// DoWithClientAndRetry send request with given client and retry for specific HTTP status
func DoWithClientAndRetry(client http.Client, req *http.Request, retry uint) (*http.Response, error) {
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode >= http.StatusBadRequest {
		if resp != nil && retry > 0 && CanRetry(req, resp) {
			time.Sleep(time.Second)
			return DoWithClientAndRetry(client, req, retry-1)
		}

		if err == nil {
			err = fmt.Errorf("HTTP/%d", resp.StatusCode)
		}

		if errBody, bodyErr := ReadBodyResponse(resp); bodyErr == nil && len(errBody) > 0 {
			err = fmt.Errorf("%s\n%s", err, errBody)
		}
	}

	return resp, err
}

// Do send request with default client
func Do(req *http.Request) (*http.Response, error) {
	return DoWithClientAndRetry(defaultHTTPClient, req, defaultRetryCount)
}

// CanRetry evaluates request and
func CanRetry(r *http.Request, resp *http.Response) bool {
	if r.Method != http.MethodGet && r.Method != http.MethodHead && r.Method != http.MethodOptions {
		return false
	}

	return resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusInternalServerError || resp.StatusCode == http.StatusBadGateway || resp.StatusCode == http.StatusServiceUnavailable
}
