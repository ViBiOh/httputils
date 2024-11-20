package request

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

var _ error = Error{}

type Error struct {
	Header     http.Header
	Body       []byte
	StatusCode int
}

func NewResponseError(resp *http.Response) Error {
	errBody, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBody))

	_ = DiscardBody(resp.Body)

	return Error{
		StatusCode: resp.StatusCode,
		Header:     resp.Header,
		Body:       errBody,
	}
}

func (re Error) Error() string {
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
