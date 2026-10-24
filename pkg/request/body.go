package request

import (
	"errors"
	"io"
	"net/http"

	"github.com/ViBiOh/httputils/v4/pkg/model"
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

func ReadBodyRequest(r *http.Request, maxSize int64) ([]byte, error) {
	if r == nil || r.Body == nil {
		return nil, nil
	}

	content, err := io.ReadAll(http.MaxBytesReader(nil, r.Body, maxSize))
	if err != nil {
		if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
			return nil, model.WrapInvalid(err)
		}

		return content, err
	}

	return content, nil
}

func ReadBodyResponse(r *http.Response) ([]byte, error) {
	if r == nil {
		return nil, nil
	}

	return readContent(r.Body)
}
