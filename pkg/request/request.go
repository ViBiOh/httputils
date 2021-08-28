package request

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/logger"
)

const (
	maxErrorBody = 500
)

var (
	defaultHTTPClient = &http.Client{
		Timeout:   15 * time.Second,
		Transport: http.DefaultTransport,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
)

// Request describe a complete request
type Request struct {
	client *http.Client
	header http.Header

	method   string
	url      string
	username string
	password string

	signatureKeydID string
	signatureSecret []byte
}

// New create a new Request
func New() Request {
	return Request{
		method: http.MethodGet,
		header: http.Header{},
		client: defaultHTTPClient,
	}
}

// IsZero checks if instance is valued
func (r Request) IsZero() bool {
	return len(r.method) == 0
}

// Method set method of Request
func (r Request) Method(method string) Request {
	r.method = method

	return r
}

// URL set URL of Request
func (r Request) URL(url string) Request {
	r.url = url

	return r
}

// Get set GET to given url
func (r Request) Get(url string) Request {
	return r.Method(http.MethodGet).URL(url)
}

// Post set POST to given url
func (r Request) Post(url string) Request {
	return r.Method(http.MethodPost).URL(url)
}

// Put set PUT to given url
func (r Request) Put(url string) Request {
	return r.Method(http.MethodPut).URL(url)
}

// Patch set PATCH to given url
func (r Request) Patch(url string) Request {
	return r.Method(http.MethodPatch).URL(url)
}

// Delete set DELETE to given url
func (r Request) Delete(url string) Request {
	return r.Method(http.MethodDelete).URL(url)
}

// BasicAuth add Basic Auth header
func (r Request) BasicAuth(username, password string) Request {
	r.username = username
	r.password = password

	return r
}

// Header add header to request
func (r Request) Header(name, value string) Request {
	r.header.Add(name, value)

	return r
}

// ContentType set Content-Type header
func (r Request) ContentType(contentType string) Request {
	return r.Header("Content-Type", contentType)
}

// ContentForm set Content-Type header to application/x-www-form-urlencoded
func (r Request) ContentForm() Request {
	return r.ContentType("application/x-www-form-urlencoded")
}

// ContentJSON set Content-Type header to application/json
func (r Request) ContentJSON() Request {
	return r.ContentType("application/json")
}

// WithClient defines net/http client to use, instead of default one (30sec timeout and no redirect)
func (r Request) WithClient(client *http.Client) Request {
	r.client = client

	return r
}

// WithSignatureAuthorization add Authorization header when sending request by calculating digest and HMAC signature header
func (r Request) WithSignatureAuthorization(keyID string, secret []byte) Request {
	r.signatureKeydID = keyID
	r.signatureSecret = secret

	return r
}

// Build create request for given context and payload
func (r Request) Build(ctx context.Context, payload io.ReadCloser) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, r.method, r.url, payload)
	if err != nil {
		return nil, err
	}

	req.Header = r.header

	if len(r.signatureSecret) != 0 {
		body, err := readContent(payload)
		if err != nil {
			return nil, fmt.Errorf("unable to read content for signature: %s", err)
		}

		AddSignature(req, r.signatureKeydID, r.signatureSecret, body)
		req.Body = io.NopCloser(bytes.NewBuffer(body))
	} else if len(r.username) != 0 || len(r.password) != 0 {
		req.SetBasicAuth(r.username, r.password)
	}

	return req, nil
}

// Send build request and send it with defined client
func (r Request) Send(ctx context.Context, payload io.ReadCloser) (*http.Response, error) {
	req, err := r.Build(ctx, payload)
	if err != nil {
		return nil, fmt.Errorf("unable to build request: %s", err)
	}

	return DoWithClient(r.client, req)
}

// Form send request with given context and url.Values as payload
func (r Request) Form(ctx context.Context, data url.Values) (*http.Response, error) {
	return r.ContentForm().Send(ctx, io.NopCloser(strings.NewReader(data.Encode())))
}

// JSON send request with given context and given interface as JSON payload
func (r Request) JSON(ctx context.Context, body interface{}) (*http.Response, error) {
	reader, writer := io.Pipe()

	go func() {
		// CloseWithError always return nil
		_ = writer.CloseWithError(json.NewEncoder(writer).Encode(body))
	}()
	defer loggedClose(reader)

	return r.ContentJSON().Send(ctx, reader)
}

// DoWithClient send request with given client
func DoWithClient(client *http.Client, req *http.Request) (*http.Response, error) {
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= http.StatusBadRequest {
		err = errors.New(convertResponseError(resp))
	}

	return resp, err
}

func convertResponseError(resp *http.Response) string {
	builder := strings.Builder{}

	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			logger.Error("unable to close error body: %s", closeErr)
		}
	}()

	fmt.Fprintf(&builder, "HTTP/%d", resp.StatusCode)

	for key, value := range resp.Header {
		fmt.Fprintf(&builder, "\n%s: %s", key, strings.Join(value, ","))
	}

	if errBody, bodyErr := io.ReadAll(io.LimitReader(resp.Body, maxErrorBody)); bodyErr == nil && len(errBody) > 0 {
		fmt.Fprintf(&builder, "\n\n%s", errBody)
	}

	// Discard remaining content
	if _, err := io.Copy(io.Discard, resp.Body); err != nil {
		logger.Error("unable to discard error body response: %s", err)
	}

	return builder.String()
}

// Do send request with default client
func Do(req *http.Request) (*http.Response, error) {
	return DoWithClient(defaultHTTPClient, req)
}

func loggedClose(closer io.Closer) {
	if err := closer.Close(); err != nil {
		logger.Error("unable to close: %s", err)
	}
}
