package request

import (
	"fmt"
	"io"
	"net/http"
)

func readContent(body io.ReadCloser) ([]byte, error) {
	if body == nil {
		return nil, nil
	}

	content, err := io.ReadAll(body)

	if closeErr := body.Close(); closeErr != nil {
		if err == nil {
			err = closeErr
		} else {
			err = fmt.Errorf("%s: %w", err, closeErr)
		}
	}

	return content, err
}

// ReadBodyRequest return content of a body request (defined as a ReadCloser)
func ReadBodyRequest(r *http.Request) ([]byte, error) {
	if r == nil {
		return nil, nil
	}
	return readContent(r.Body)
}

// ReadBodyResponse return content of a body response (defined as a ReadCloser)
func ReadBodyResponse(r *http.Response) ([]byte, error) {
	if r == nil {
		return nil, nil
	}
	return readContent(r.Body)
}
