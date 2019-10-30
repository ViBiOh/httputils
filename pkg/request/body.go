package request

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

// ReadContent return content of given body
func ReadContent(body io.ReadCloser) (content []byte, err error) {
	if body == nil {
		return
	}

	defer func() {
		if closeErr := body.Close(); closeErr != nil {
			if err == nil {
				err = closeErr
			} else {
				err = fmt.Errorf("%s: %w", err, closeErr)
			}
		}
	}()

	content, err = ioutil.ReadAll(body)
	return
}

// ReadBodyRequest return content of a body request (defined as a ReadCloser)
func ReadBodyRequest(r *http.Request) ([]byte, error) {
	if r == nil {
		return nil, nil
	}
	return ReadContent(r.Body)
}

// ReadBodyResponse return content of a body response (defined as a ReadCloser)
func ReadBodyResponse(r *http.Response) ([]byte, error) {
	if r == nil {
		return nil, nil
	}
	return ReadContent(r.Body)
}
