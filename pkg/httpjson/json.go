package httpjson

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/ViBiOh/httputils/v4/pkg/httperror"
	"github.com/ViBiOh/httputils/v4/pkg/query"
)

var (
	// ErrCannotMarshall occurs when marshaller failed
	ErrCannotMarshall = errors.New("cannot marshall json")
)

type results struct {
	Results interface{} `json:"results"`
}

type pagination struct {
	Results   interface{} `json:"results"`
	PageSize  uint        `json:"pageSize"`
	PageCount uint        `json:"pageCount"`
	Total     uint        `json:"total"`
	LastKey   string      `json:"lastKey"`
}

// IsPretty determine if pretty is defined in query params
func IsPretty(r *http.Request) bool {
	return query.GetBool(r, "pretty")
}

// Write writes marshalled obj to http.ResponseWriter with correct header
func Write(w http.ResponseWriter, status int, obj interface{}, pretty bool) {
	encoder := json.NewEncoder(w)
	if pretty {
		encoder.SetIndent("", "  ")
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(status)

	if err := encoder.Encode(obj); err != nil {
		httperror.InternalServerError(w, fmt.Errorf("%s: %w", err, ErrCannotMarshall))
	}
}

// WriteArray write marshalled obj wrapped into an object to http.ResponseWriter with correct header
func WriteArray(w http.ResponseWriter, status int, array interface{}, pretty bool) {
	Write(w, status, results{array}, pretty)
}

// WritePagination write marshalled obj wrapped into an object to http.ResponseWriter with correct header
func WritePagination(w http.ResponseWriter, status int, pageSize, total uint, lastKey string, array interface{}, pretty bool) {
	pageCount := total / pageSize
	if total%pageSize != 0 {
		pageCount++
	}

	Write(w, status, pagination{Results: array, PageSize: pageSize, PageCount: pageCount, Total: total, LastKey: lastKey}, pretty)
}

// Parse read body resquest and unmarshall it into given interface
func Parse(req *http.Request, obj interface{}, action string) (err error) {
	decoder := json.NewDecoder(req.Body)
	defer func() {
		if closeErr := req.Body.Close(); closeErr != nil {
			if err == nil {
				err = closeErr
			} else {
				err = fmt.Errorf("%s: %w", err, closeErr)
			}
		}
	}()

	if err = decoder.Decode(obj); err != nil {
		err = fmt.Errorf("unable to parse body of %s: %s", action, err)
	}

	return
}

// Read read body response and unmarshall it into given interface
func Read(resp *http.Response, obj interface{}, action string) (err error) {
	decoder := json.NewDecoder(resp.Body)
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			if err == nil {
				err = closeErr
			} else {
				err = fmt.Errorf("%s: %w", err, closeErr)
			}
		}
	}()

	if err = decoder.Decode(obj); err != nil {
		err = fmt.Errorf("unable to parse body of %s: %s", action, err)
	}

	return
}

// Stream reads io.Reader and stream array or map content to given chan
func Stream(stream io.Reader, newObj func() interface{}, output chan<- interface{}) error {
	defer close(output)
	decoder := json.NewDecoder(stream)

	if _, err := decoder.Token(); err != nil {
		return fmt.Errorf("unable to read opening token: %s", err)
	}

	for decoder.More() {
		obj := newObj()
		if err := decoder.Decode(obj); err != nil {
			return fmt.Errorf("unable to parse JSON: %s", err)
		}
		output <- obj
	}

	if _, err := decoder.Token(); err != nil {
		return fmt.Errorf("unable to read closing token: %s", err)
	}

	return nil
}
