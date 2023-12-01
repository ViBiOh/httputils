package httpjson

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ViBiOh/httputils/v4/pkg/httperror"
)

var (
	ErrCannotMarshal = errors.New("cannot marshal json")
	headers          = http.Header{}
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

func RawWrite(w io.Writer, obj any) error {
	if err := json.NewEncoder(w).Encode(obj); err != nil {
		return fmt.Errorf("%s: %w", err, ErrCannotMarshal)
	}

	return nil
}

func Write(ctx context.Context, w http.ResponseWriter, status int, obj any) {
	for key, value := range headers {
		w.Header()[key] = value
	}
	w.WriteHeader(status)

	if err := RawWrite(w, obj); err != nil {
		httperror.InternalServerError(ctx, w, err)
	}
}

func WriteArray(ctx context.Context, w http.ResponseWriter, status int, array any) {
	Write(ctx, w, status, items{array})
}

func WritePagination(ctx context.Context, w http.ResponseWriter, status int, pageSize, total uint, last string, array any) {
	pageCount := total / pageSize
	if total%pageSize != 0 {
		pageCount++
	}

	Write(ctx, w, status, pagination{Items: array, PageSize: pageSize, PageCount: pageCount, Total: total, Last: last})
}

func Parse(req *http.Request, obj any) error {
	if err := json.NewDecoder(req.Body).Decode(obj); err != nil {
		return fmt.Errorf("parse JSON: %w", err)
	}

	return nil
}

func Read(resp *http.Response, obj any) error {
	var err error

	if err = json.NewDecoder(resp.Body).Decode(obj); err != nil {
		err = fmt.Errorf("read JSON: %w", err)
	}

	if closeErr := resp.Body.Close(); closeErr != nil {
		return errors.Join(err, closeErr)
	}

	return err
}

func Stream[T any](stream io.Reader, output chan<- T, key string, closeChan bool) error {
	if closeChan {
		defer close(output)
	}
	decoder := json.NewDecoder(stream)

	if len(key) > 0 {
		if err := moveDecoderToKey(decoder, key); err != nil {
			return err
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

func moveDecoderToKey(decoder *json.Decoder, key string) error {
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

	return nil
}
