package request

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/ViBiOh/httputils/v4/pkg/model"
	"github.com/stretchr/testify/assert"
)

var (
	_ io.Reader     = errReader(0)
	_ io.ReadCloser = errCloser(0)
	_ io.ReadCloser = errReaderCloser(0)
)

type errReader int

func (errReader) Read(_ []byte) (int, error) {
	return 0, errors.New("read error")
}

type errCloser int

func (errCloser) Read(_ []byte) (int, error) {
	return 0, io.EOF
}

func (errCloser) Close() error {
	return errors.New("close error")
}

type errReaderCloser int

func (errReaderCloser) Read(_ []byte) (int, error) {
	return 0, errors.New("read error")
}

func (errReaderCloser) Close() error {
	return errors.New("close error")
}

func TestReadContent(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		reader  io.ReadCloser
		want    []byte
		wantErr error
	}{
		"nil input": {
			nil,
			nil,
			nil,
		},
		"basic read": {
			io.NopCloser(bytes.NewReader([]byte("Content"))),
			[]byte("Content"),
			nil,
		},
		"read with error": {
			io.NopCloser(errReader(0)),
			[]byte{},
			errors.New("read error"),
		},
		"close with error": {
			errCloser(0),
			[]byte{},
			errors.New("close error"),
		},
		"read and close error, close error": {
			errReaderCloser(0),
			[]byte{},
			errors.New("close error"),
		},
		"read and close error, read error": {
			errReaderCloser(0),
			[]byte{},
			errors.New("read error"),
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			result, err := readContent(testCase.reader)

			failed := false

			if testCase.wantErr != nil && errors.Is(err, testCase.wantErr) {
				failed = true
			} else if !reflect.DeepEqual(result, testCase.want) {
				failed = true
			}

			if failed {
				t.Errorf("readContent() = (`%s`, `%s`), want (`%s`, `%s`)", result, err, testCase.want, testCase.wantErr)
			}
		})
	}
}

func TestReadBodyRequest(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		input   *http.Request
		want    []byte
		wantErr error
	}{
		"nil": {
			nil,
			nil,
			nil,
		},
		"simple": {
			httptest.NewRequest(http.MethodGet, "/", bytes.NewReader([]byte("Content"))),
			[]byte("Content"),
			nil,
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			result, err := ReadBodyRequest(testCase.input, 1024)

			failed := false

			if testCase.wantErr != nil && errors.Is(err, testCase.wantErr) {
				failed = true
			} else if !reflect.DeepEqual(result, testCase.want) {
				failed = true
			}

			if failed {
				t.Errorf("ReadBodyRequest() = (`%s`, `%s`), want (`%s`, `%s`)", result, err, testCase.want, testCase.wantErr)
			}
		})
	}
}

func TestReadBodyResponse(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		input   []byte
		want    []byte
		wantErr error
	}{
		"nil": {
			nil,
			[]byte{},
			nil,
		},
		"simple": {
			[]byte("Content"),
			[]byte("Content"),
			nil,
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			writer := httptest.NewRecorder()
			if _, err := writer.Write(testCase.input); err != nil {
				t.Errorf("write: %s", err)
			}
			result, err := ReadBodyResponse(writer.Result())

			failed := false

			if testCase.wantErr != nil && errors.Is(err, testCase.wantErr) {
				failed = true
			} else if !reflect.DeepEqual(result, testCase.want) {
				failed = true
			}

			if failed {
				t.Errorf("ReadBodyResponse() = (`%s`, `%s`), want (`%s`, `%s`)", result, err, testCase.want, testCase.wantErr)
			}
		})
	}
}

func TestReadBodyRequestLimit(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		body    string
		maxSize int64
		want    string
		wantErr error
	}{
		"under the limit": {
			"Content",
			1024,
			"Content",
			nil,
		},
		"exactly the limit": {
			"Content",
			7,
			"Content",
			nil,
		},
		"one byte over the limit": {
			"Content",
			6,
			"",
			model.ErrInvalid,
		},
		"far over the limit": {
			strings.Repeat("x", 1<<20),
			16,
			"",
			model.ErrInvalid,
		},
		"empty body": {
			"",
			0,
			"",
			nil,
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(testCase.body))

			got, err := ReadBodyRequest(req, testCase.maxSize)

			assert.Equal(t, testCase.want, string(got))

			if testCase.wantErr != nil {
				assert.ErrorIs(t, err, testCase.wantErr)

				var tooLarge *http.MaxBytesError
				assert.ErrorAs(t, err, &tooLarge)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
