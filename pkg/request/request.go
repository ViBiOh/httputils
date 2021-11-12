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
)

const (
	maxErrorBody = 500
)

var (
	discarder = io.Discard.(io.ReaderFrom)

	defaultHTTPClient = CreateClient(15*time.Second, NoRedirection)
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

// String representation of the request
func (r Request) String() string {
	var builder strings.Builder

	if len(r.method) != 0 {
		builder.WriteString(strings.ToUpper(r.method))
	}

	if len(r.url) != 0 {
		builder.WriteString(" ")
		builder.WriteString(r.url)
	}

	if len(r.signatureSecret) != 0 {
		builder.WriteString(", SignatureAuthorization with key `")
		builder.WriteString(r.signatureKeydID)
		builder.WriteString("`")
	} else if len(r.username) != 0 || len(r.password) != 0 {
		builder.WriteString(", BasicAuth with user `%s`")
		builder.WriteString(r.username)
	}

	for key, values := range r.header {
		builder.WriteString(", Header ")
		builder.WriteString(key)
		builder.WriteString(": `")
		builder.WriteString(strings.Join(values, ", "))
		builder.WriteString("`")
	}

	return builder.String()
}

// IsZero checks if instance is valued
func (r Request) IsZero() bool {
	return len(r.method) == 0 || len(r.url) == 0
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

// Path appends given path to the current URL
func (r Request) Path(path string) Request {
	if len(path) == 0 {
		return r
	}

	var status uint

	if strings.HasPrefix(path, "/") {
		status |= 1
	}
	if strings.HasSuffix(r.url, "/") {
		status |= 1 << 1
	}

	switch status {
	case 0:
		r.url = fmt.Sprintf("%s/%s", r.url, path)
	case 1, 2:
		r.url += path
	case 3:
		r.url += path[1:]
	}

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
	r.header = r.header.Clone()
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

// WithClient defines net/http client to use, instead of default one (15sec timeout and no redirect)
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

	resp, err := r.ContentJSON().Send(ctx, reader)

	if closeErr := reader.Close(); closeErr != nil {
		if err == nil {
			err = closeErr
		} else {
			err = fmt.Errorf("%s: %w", err, closeErr)
		}
	}

	return resp, err
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

	fmt.Fprintf(&builder, "HTTP/%d", resp.StatusCode)

	for key, value := range resp.Header {
		fmt.Fprintf(&builder, "\n%s: %s", key, strings.Join(value, ","))
	}

	if errBody, bodyErr := io.ReadAll(io.LimitReader(resp.Body, maxErrorBody)); bodyErr == nil && len(errBody) > 0 {
		fmt.Fprintf(&builder, "\n\n%s", errBody)
	}

	if err := DiscardBody(resp.Body); err != nil {
		fmt.Fprintf(&builder, "\nunable to discard body response: %s", err)
	}

	return builder.String()
}

// DiscardBody of a response
func DiscardBody(body io.ReadCloser) error {
	var err error

	if _, readErr := discarder.ReadFrom(body); readErr != nil {
		err = readErr
	}

	if closeErr := body.Close(); closeErr != nil {
		if err == nil {
			err = closeErr
		} else {
			err = fmt.Errorf("%s: %w", err, closeErr)
		}
	}

	return err
}

// Do send request with default client
func Do(req *http.Request) (*http.Response, error) {
	return DoWithClient(defaultHTTPClient, req)
}
