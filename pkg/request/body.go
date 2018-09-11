package request

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

func readBody(body io.ReadCloser) (content []byte, err error) {
	defer func() {
		closeErr := body.Close()

		if closeErr != nil {
			if err != nil {
				err = fmt.Errorf(`, and also error while closing body: %v`, closeErr)
			} else {
				err = closeErr
			}
		}
	}()

	content, err = ioutil.ReadAll(body)
	return
}

// ReadBodyRequest return content of a body request (defined as a ReadCloser)
func ReadBodyRequest(r *http.Request) ([]byte, error) {
	return readBody(r.Body)
}

// ReadBodyResponse return content of a body response (defined as a ReadCloser)
func ReadBodyResponse(r *http.Response) ([]byte, error) {
	return readBody(r.Body)
}
