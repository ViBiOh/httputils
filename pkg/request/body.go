package request

import (
	"errors"
	"io"
	"net/http"
)

func readContent(body io.ReadCloser) ([]byte, error) {
	if body == nil {
		return nil, nil
	}

	content, err := io.ReadAll(body)

	if closeErr := body.Close(); closeErr != nil {
		err = errors.Join(err, closeErr)
	}

	return content, err
}

func ReadBodyRequest(r *http.Request) ([]byte, error) {
	if r == nil {
		return nil, nil
	}

	return io.ReadAll(r.Body)
}

func ReadBodyResponse(r *http.Response) ([]byte, error) {
	if r == nil {
		return nil, nil
	}

	return readContent(r.Body)
}
