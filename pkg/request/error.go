package request

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

var _ error = RequestError{}

type RequestError struct {
	Header     http.Header
	Body       []byte
	StatusCode int
}

func NewResponseError(resp *http.Response) RequestError {
	errBody, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBody))

	_ = DiscardBody(resp.Body)

	return RequestError{
		StatusCode: resp.StatusCode,
		Header:     resp.Header,
		Body:       errBody,
	}
}

func (re RequestError) Error() string {
	builder := strings.Builder{}

	_, _ = fmt.Fprintf(&builder, "HTTP/%d", re.StatusCode)

	for key, value := range re.Header {
		_, _ = fmt.Fprintf(&builder, "\n%s: %s", key, strings.Join(value, ","))
	}

	if len(re.Body) > 0 {
		_, _ = fmt.Fprintf(&builder, "\n\n%s", re.Body)
	}

	return builder.String()
}
