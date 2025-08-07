package httpjson

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
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

func RawWrite(w io.Writer, obj any) error {
	if err := json.NewEncoder(w).Encode(obj); err != nil {
		return fmt.Errorf("%s: %w", err, ErrCannotMarshal)
	}

	return nil
}

func Write(ctx context.Context, w http.ResponseWriter, status int, obj any) {
	maps.Copy(w.Header(), headers)
	w.WriteHeader(status)

	if err := RawWrite(w, obj); err != nil {
		httperror.InternalServerError(ctx, w, err)
	}
}

func WriteArray(ctx context.Context, w http.ResponseWriter, status int, array any) {
	Write(ctx, w, status, items{array})
}

func Parse[T any](req *http.Request) (T, error) {
	var output T

	return output, json.NewDecoder(req.Body).Decode(&output)
}

func Read[T any](resp *http.Response) (output T, err error) {
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			err = errors.Join(err, closeErr)
		}
	}()

	return output, json.NewDecoder(resp.Body).Decode(&output)
}

// Stream decodes JSON from a `reader` and send unmarshalled `T` to the given `output` chan
// `closeChan` make Stream to close the `output` channel once read.
// There is no guarantee that the `reader` wil be fully read. For HTTP Body you might want to drain it.
//
// `key` supports three kinds of values:
// - ""  => the reader contains a stream of objects like `{"id":12}{"id":34}{"id":56}`, e.g. processing a websocket
// - "." => the reader contains an array of objects like `[{"id":12},{"id":34},{"id":56}]`
// - <key name>" => the reader contains an array of objects at the given root key `{"items":[{"id":12},{"id":34},{"id":56}]}`.
func Stream[T any](reader io.Reader, output chan<- T, key string, closeChan bool) error {
	if closeChan {
		defer close(output)
	}
	decoder := json.NewDecoder(reader)

	if len(key) > 0 {
		if err := moveToKey(decoder, key); err != nil {
			return fmt.Errorf("move to key: %w", err)
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

func moveToKey(decoder *json.Decoder, key string) error {
	if key != "." {
		if err := findStart(decoder, key); err != nil {
			return fmt.Errorf("find start: %w", err)
		}
	}

	if _, err := decoder.Token(); err != nil {
		return fmt.Errorf("read opening token: %w", err)
	}

	return nil
}

func findStart(decoder *json.Decoder, key string) error {
	var nested uint64

	for {
		token, err := decoder.Token()
		if err != nil {
			return fmt.Errorf("decode token: %w", err)
		}

		switch tokenType := token.(type) {
		case string:
			if nested == 1 && strings.EqualFold(tokenType, key) {
				return nil
			}

		case json.Delim:
			switch tokenType {
			case '{':
				nested++
			case '}':
				nested--
			}
		}
	}
}
