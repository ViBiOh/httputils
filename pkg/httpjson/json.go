package httpjson

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ViBiOh/httputils/v4/pkg/httperror"
	"github.com/ViBiOh/httputils/v4/pkg/query"
)

var (
	// ErrCannotMarshal occurs when marshaller failed
	ErrCannotMarshal = errors.New("cannot marshall json")

	headers = http.Header{}
)

func init() {
	headers.Add("Content-Type", "application/json; charset=utf-8")
	headers.Add("Cache-Control", "no-cache")
}

type items struct {
	Items interface{} `json:"items"`
}

type pagination struct {
	Items     interface{} `json:"items"`
	Last      string      `json:"last"`
	PageSize  uint        `json:"pageSize"`
	PageCount uint        `json:"pageCount"`
	Total     uint        `json:"total"`
}

// IsPretty determine if pretty is defined in query params
func IsPretty(r *http.Request) bool {
	return query.GetBool(r, "pretty")
}

// RawWrite writes marshalled obj to io.Writer
func RawWrite(w io.Writer, obj interface{}, pretty bool) error {
	encoder := json.NewEncoder(w)
	if pretty {
		encoder.SetIndent("", "  ")
	}

	if err := encoder.Encode(obj); err != nil {
		return fmt.Errorf("%s: %w", err, ErrCannotMarshal)
	}
	return nil
}

// Write writes marshalled obj to http.ResponseWriter with correct header
func Write(w http.ResponseWriter, status int, obj interface{}, pretty bool) {
	for key, value := range headers {
		w.Header()[key] = value
	}
	w.WriteHeader(status)

	if err := RawWrite(w, obj, pretty); err != nil {
		httperror.InternalServerError(w, err)
	}
}

// WriteArray write marshalled obj wrapped into an object to http.ResponseWriter with correct header
func WriteArray(w http.ResponseWriter, status int, array interface{}, pretty bool) {
	Write(w, status, items{array}, pretty)
}

// WritePagination write marshalled obj wrapped into an object to http.ResponseWriter with correct header
func WritePagination(w http.ResponseWriter, status int, pageSize, total uint, last string, array interface{}, pretty bool) {
	pageCount := total / pageSize
	if total%pageSize != 0 {
		pageCount++
	}

	Write(w, status, pagination{Items: array, PageSize: pageSize, PageCount: pageCount, Total: total, Last: last}, pretty)
}

// Parse read body resquest and unmarshal it into given interface
func Parse(req *http.Request, obj interface{}) (err error) {
	return decode(req.Body, obj)
}

// Read read body response and unmarshal it into given interface
func Read(resp *http.Response, obj interface{}) (err error) {
	return decode(resp.Body, obj)
}

func decode(input io.Reader, obj interface{}) (err error) {
	decoder := json.NewDecoder(input)

	if closer, ok := input.(io.Closer); ok {
		defer func() {
			if closeErr := closer.Close(); closeErr != nil {
				if err == nil {
					err = closeErr
				}

				err = fmt.Errorf("%s: %w", err, closeErr)
			}
		}()
	}

	if err = decoder.Decode(obj); err != nil {
		err = fmt.Errorf("unable to parse JSON body: %s", err)
	}

	return
}

// Stream reads io.Reader and stream array or map content to given chan
func Stream(stream io.Reader, newObj func() interface{}, output chan<- interface{}, key string) error {
	defer close(output)
	decoder := json.NewDecoder(stream)

	var token json.Token
	var err error
	for !strings.EqualFold(fmt.Sprintf("%s", token), key) {
		token, err = decoder.Token()
		if err != nil {
			return fmt.Errorf("unable to read token: %s", err)
		}
	}

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
