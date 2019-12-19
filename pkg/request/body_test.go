package request

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

var (
	_ io.Reader     = errReader(0)
	_ io.ReadCloser = errCloser(0)
	_ io.ReadCloser = errReaderCloser(0)
)

type errReader int

func (errReader) Read(p []byte) (int, error) {
	return 0, errors.New("read error")
}

type errCloser int

func (errCloser) Read(p []byte) (int, error) {
	return 0, io.EOF
}

func (errCloser) Close() error {
	return errors.New("close error")
}

type errReaderCloser int

func (errReaderCloser) Read(p []byte) (int, error) {
	return 0, errors.New("read error")
}

func (errReaderCloser) Close() error {
	return errors.New("close error")
}

func TestReadContent(t *testing.T) {
	var cases = []struct {
		intention string
		reader    io.ReadCloser
		want      []byte
		wantErr   error
	}{
		{
			"nil input",
			nil,
			nil,
			nil,
		},
		{
			"basic read",
			ioutil.NopCloser(bytes.NewReader([]byte("Content"))),
			[]byte("Content"),
			nil,
		},
		{
			"read with error",
			ioutil.NopCloser(errReader(0)),
			[]byte{},
			errors.New("read error"),
		},
		{
			"close with error",
			errCloser(0),
			[]byte{},
			errors.New("close error"),
		},
		{
			"read and close error, close error",
			errReaderCloser(0),
			[]byte{},
			errors.New("close error"),
		},
		{
			"read and close error, read error",
			errReaderCloser(0),
			[]byte{},
			errors.New("read error"),
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			result, err := readContent(testCase.reader)

			failed := false

			if testCase.wantErr != nil && errors.Is(err, testCase.wantErr) {
				failed = true
			} else if !reflect.DeepEqual(result, testCase.want) {
				failed = true
			}

			if failed {
				t.Errorf("readContent() = (%s, %s), want (%s, %s)", result, err, testCase.want, testCase.wantErr)
			}
		})
	}
}

func TestReadBodyRequest(t *testing.T) {
	var cases = []struct {
		intention string
		input     *http.Request
		want      []byte
		wantErr   error
	}{
		{
			"nil",
			nil,
			nil,
			nil,
		},
		{
			"simple",
			httptest.NewRequest(http.MethodGet, "/", bytes.NewReader([]byte("Content"))),
			[]byte("Content"),
			nil,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			result, err := ReadBodyRequest(testCase.input)

			failed := false

			if testCase.wantErr != nil && errors.Is(err, testCase.wantErr) {
				failed = true
			} else if !reflect.DeepEqual(result, testCase.want) {
				failed = true
			}

			if failed {
				t.Errorf("ReadBodyRequest() = (%s, %s), want (%s, %s)", result, err, testCase.want, testCase.wantErr)
			}
		})
	}
}

func TestReadBodyResponse(t *testing.T) {
	var cases = []struct {
		intention string
		input     []byte
		want      []byte
		wantErr   error
	}{
		{
			"nil",
			nil,
			[]byte{},
			nil,
		},
		{
			"simple",
			[]byte("Content"),
			[]byte("Content"),
			nil,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			writer := httptest.NewRecorder()
			writer.Write(testCase.input)
			result, err := ReadBodyResponse(writer.Result())

			failed := false

			if testCase.wantErr != nil && errors.Is(err, testCase.wantErr) {
				failed = true
			} else if !reflect.DeepEqual(result, testCase.want) {
				failed = true
			}

			if failed {
				t.Errorf("ReadBodyResponse() = (%s, %s), want (%s, %s)", result, err, testCase.want, testCase.wantErr)
			}
		})
	}
}
