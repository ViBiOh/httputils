package httpjson

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ViBiOh/httputils/v4/pkg/httperror"
	"github.com/ViBiOh/httputils/v4/pkg/model"
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
	Items any `json:"items"`
}

type pagination struct {
	Items     any    `json:"items"`
	Last      string `json:"last"`
	PageSize  uint   `json:"pageSize"`
	PageCount uint   `json:"pageCount"`
	Total     uint   `json:"total"`
}

// RawWrite writes marshalled obj to io.Writer
func RawWrite(w io.Writer, obj any) error {
	if err := json.NewEncoder(w).Encode(obj); err != nil {
		return fmt.Errorf("%s: %w", err, ErrCannotMarshal)
	}

	return nil
}

// Write writes marshalled obj to http.ResponseWriter with correct header
func Write(w http.ResponseWriter, status int, obj any) {
	for key, value := range headers {
		w.Header()[key] = value
	}
	w.WriteHeader(status)

	if err := RawWrite(w, obj); err != nil {
		httperror.InternalServerError(w, err)
	}
}

// WriteArray write marshalled obj wrapped into an object to http.ResponseWriter with correct header
func WriteArray(w http.ResponseWriter, status int, array any) {
	Write(w, status, items{array})
}

// WritePagination write marshalled obj wrapped into an object to http.ResponseWriter with correct header
func WritePagination(w http.ResponseWriter, status int, pageSize, total uint, last string, array any) {
	pageCount := total / pageSize
	if total%pageSize != 0 {
		pageCount++
	}

	Write(w, status, pagination{Items: array, PageSize: pageSize, PageCount: pageCount, Total: total, Last: last})
}

// Parse read body resquest and unmarshal it into given interface
func Parse(req *http.Request, obj any) error {
	if err := json.NewDecoder(req.Body).Decode(obj); err != nil {
		return fmt.Errorf("parse JSON: %w", err)
	}

	return nil
}

// Read body response and unmarshal it into given interface
func Read(resp *http.Response, obj any) error {
	var err error

	if err = json.NewDecoder(resp.Body).Decode(obj); err != nil {
		err = fmt.Errorf("read JSON: %w", err)
	}

	if closeErr := resp.Body.Close(); closeErr != nil {
		return model.WrapError(err, closeErr)
	}

	return err
}

// Stream reads io.Reader and stream content to given chan
func Stream[T any](stream io.Reader, output chan<- T, key string, closeChan bool) error {
	if closeChan {
		defer close(output)
	}
	decoder := json.NewDecoder(stream)

	if len(key) > 0 {
		var token json.Token
		var nested uint64
		var err error

		for {
			token, err = decoder.Token()
			if err != nil {
				return fmt.Errorf("decode token: %w", err)
			}

			if nested == 1 && strings.EqualFold(fmt.Sprintf("%s", token), key) {
				break
			}

			if strToken := fmt.Sprintf("%s", token); strToken == "{" {
				nested++
			} else if strToken == "}" {
				nested--
			}
		}

		if _, err = decoder.Token(); err != nil {
			return fmt.Errorf("read opening token: %w", err)
		}
	}

	var obj T
	for decoder.More() {
		if err := decoder.Decode(&obj); err != nil {
			return fmt.Errorf("decode stream: %w", err)
		}
		output <- obj
	}

	if len(key) > 0 {
		if _, err := decoder.Token(); err != nil {
			return fmt.Errorf("read closing token: %w", err)
		}
	}

	return nil
}
