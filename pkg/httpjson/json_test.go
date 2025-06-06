package httpjson

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ViBiOh/httputils/v4/pkg/request"
	"github.com/stretchr/testify/assert"
)

type testStruct struct {
	id     string
	Active bool
	Amount float64
}

func TestRawWrite(t *testing.T) {
	t.Parallel()

	type args struct {
		writer *bytes.Buffer
		obj    any
	}

	cases := map[string]struct {
		args    args
		want    string
		wantErr error
	}{
		"invalid": {
			args{
				writer: bytes.NewBufferString(""),
				obj:    func() {},
			},
			"",
			ErrCannotMarshal,
		},
		"simple": {
			args{
				writer: bytes.NewBufferString(""),
				obj: map[string]any{
					"key":   "value",
					"valid": true,
				},
			},
			"{\"key\":\"value\",\"valid\":true}\n",
			nil,
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			gotErr := RawWrite(testCase.args.writer, testCase.args.obj)

			assert.Equal(t, testCase.want, testCase.args.writer.String())
			checkErr(t, testCase.wantErr, gotErr)
		})
	}
}

func TestWrite(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		obj        any
		want       string
		wantStatus int
		wantHeader map[string]string
	}{
		"nil": {
			nil,
			"null\n",
			http.StatusOK,
			map[string]string{"Content-Type": "application/json; charset=utf-8", "Cache-Control": "no-cache"},
		},
		"simple object": {
			testStruct{id: "Test"},
			"{\"Active\":false,\"Amount\":0}\n",
			http.StatusOK,
			map[string]string{"Content-Type": "application/json; charset=utf-8", "Cache-Control": "no-cache"},
		},
		"error": {
			func() string {
				return "test"
			},
			"Oops! Something went wrong. Server's logs contain more details.\n",
			http.StatusOK, // might not occur in real life
			map[string]string{"Content-Type": "application/json; charset=utf-8", "Cache-Control": "no-cache"},
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			writer := httptest.NewRecorder()
			Write(context.Background(), writer, http.StatusOK, testCase.obj)

			if result, _ := request.ReadBodyResponse(writer.Result()); string(result) != testCase.want {
				t.Errorf("Write() = `%s`, want `%s`", string(result), testCase.want)
			}

			if result := writer.Result().StatusCode; result != testCase.wantStatus {
				t.Errorf("Write() = %d, want %d", result, testCase.wantStatus)
			}

			for key, value := range testCase.wantHeader {
				if result, ok := writer.Result().Header[key]; !ok || strings.Join(result, "") != value {
					t.Errorf("Write().Header[%s] = `%s`, want `%s`", key, strings.Join(result, ""), value)
				}
			}
		})
	}
}

func BenchmarkRawWrite(b *testing.B) {
	obj := testStruct{id: "Test", Active: true, Amount: 12.34}

	for b.Loop() {
		if err := RawWrite(io.Discard, &obj); err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkWrite(b *testing.B) {
	testCase := struct {
		obj any
	}{
		testStruct{id: "Test", Active: true, Amount: 12.34},
	}

	writer := httptest.NewRecorder()

	for b.Loop() {
		Write(context.Background(), writer, http.StatusOK, &testCase.obj)
	}
}

func TestWriteArray(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		obj        any
		want       string
		wantHeader map[string]string
	}{
		"nil": {
			nil,
			"{\"items\":null}\n",
			map[string]string{"Content-Type": "application/json; charset=utf-8", "Cache-Control": "no-cache"},
		},
		"simple": {
			[]testStruct{{id: "First", Active: true, Amount: 12.34}, {id: "Second", Active: true, Amount: 12.34}},
			"{\"items\":[{\"Active\":true,\"Amount\":12.34},{\"Active\":true,\"Amount\":12.34}]}\n",
			map[string]string{"Content-Type": "application/json; charset=utf-8", "Cache-Control": "no-cache"},
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			writer := httptest.NewRecorder()
			WriteArray(context.Background(), writer, http.StatusOK, testCase.obj)

			if result, _ := request.ReadBodyResponse(writer.Result()); string(result) != testCase.want {
				t.Errorf("TestWriteArray() = `%s`, want `%s`", string(result), testCase.want)
			}

			for key, value := range testCase.wantHeader {
				if result, ok := writer.Result().Header[key]; !ok || strings.Join(result, "") != value {
					t.Errorf("TestWriteArray().Header[%s] = `%s`, want `%s`", key, strings.Join(result, ""), value)
				}
			}
		})
	}
}

func TestParse(t *testing.T) {
	t.Parallel()

	type args struct {
		req *http.Request
	}

	cases := map[string]struct {
		args    args
		want    map[string]any
		wantErr error
	}{
		"valid": {
			args{
				req: httptest.NewRequest(http.MethodGet, "/", bytes.NewBufferString(`{"key": "value","valid":true}`)),
			},
			map[string]any{
				"key":   "value",
				"valid": true,
			},
			nil,
		},
		"invalid": {
			args{
				req: httptest.NewRequest(http.MethodGet, "/", bytes.NewBufferString(`{"key": "value","valid":true`)),
			},
			nil,
			errors.New("EOF"),
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			got, gotErr := Parse[map[string]any](testCase.args.req)

			assert.Equal(t, testCase.want, got)
			checkErr(t, testCase.wantErr, gotErr)
		})
	}
}

type errCloser struct {
	io.Reader
}

func (errCloser) Close() error {
	return errors.New("close error")
}

func TestRead(t *testing.T) {
	t.Parallel()

	type args struct {
		resp *http.Response
	}

	cases := map[string]struct {
		args    args
		want    map[string]any
		wantErr error
	}{
		"parse error": {
			args{
				resp: &http.Response{
					Body: io.NopCloser(bytes.NewReader([]byte("invalid json"))),
				},
			},
			nil,
			errors.New("invalid character"),
		},
		"close error": {
			args{
				resp: &http.Response{
					Body: errCloser{bytes.NewReader([]byte(`{"key": "value","valid":true}`))},
				},
			},
			map[string]any{
				"key":   "value",
				"valid": true,
			},
			errors.New("close error"),
		},
		"both error": {
			args{
				resp: &http.Response{
					Body: errCloser{bytes.NewReader([]byte(`invalid json`))},
				},
			},
			nil,
			errors.New("invalid character 'i' looking for beginning of value\nclose error"),
		},
		"valid": {
			args{
				resp: &http.Response{
					Body: io.NopCloser(bytes.NewReader([]byte(`{"key": "value","valid":true}`))),
				},
			},
			map[string]any{
				"key":   "value",
				"valid": true,
			},
			nil,
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			got, gotErr := Read[map[string]any](testCase.args.resp)

			assert.Equal(t, testCase.want, got)
			checkErr(t, testCase.wantErr, gotErr)
		})
	}
}

func TestStream(t *testing.T) {
	t.Parallel()

	type simpleStruct struct {
		Value string `json:"value"`
	}

	type args struct {
		stream io.Reader
		key    string
	}

	cases := map[string]struct {
		args    args
		want    []string
		wantErr error
	}{
		"invalid json": {
			args{
				stream: strings.NewReader("invalid json"),
				key:    "items",
			},
			nil,
			errors.New("decode token"),
		},
		"no opening token": {
			args{
				stream: strings.NewReader(`{"count": 10, "items"}`),
				key:    "items",
			},
			nil,
			errors.New("read opening token"),
		},
		"no closing token": {
			args{
				stream: strings.NewReader(`{"count": 10, "items": [{"value":"test"},{"value":"next"},{"value":"final"}`),
				key:    "items",
			},
			[]string{"test", "next", "final"},
			errors.New("read closing token"),
		},
		"success": {
			args{
				stream: strings.NewReader(`{"count": 10, "items": [{"value":"test"},{"value":"next"},{"value":"final"}]}`),
				key:    "items",
			},
			[]string{"test", "next", "final"},
			nil,
		},
		"nested": {
			args{
				stream: strings.NewReader(`{"count": 10, "nested": {"items": ["test"]}, "items": [{"value":"test"},{"value":"next"},{"value":"final"}]}`),
				key:    "items",
			},
			[]string{"test", "next", "final"},
			nil,
		},
		"streamed": {
			args{
				stream: strings.NewReader(`{"value":"test"}{"value":"next"}{"value":"final"}`),
				key:    "",
			},
			[]string{"test", "next", "final"},
			nil,
		},
		"direct array": {
			args{
				stream: strings.NewReader(`[{"value":"test"},{"value":"next"},{"value":"final"}]`),
				key:    ".",
			},
			[]string{"test", "next", "final"},
			nil,
		},
		"stream error": {
			args{
				stream: strings.NewReader(`{"value":"test"}{"value":"next"}{"value":"final}`),
				key:    "",
			},
			[]string{"test", "next"},
			errors.New("decode stream"),
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			output := make(chan simpleStruct, 4)
			done := make(chan struct{})
			var got []string

			go func() {
				defer close(done)
				for item := range output {
					got = append(got, item.Value)
				}
			}()

			gotErr := Stream(testCase.args.stream, output, testCase.args.key, true)

			<-done

			assert.Equal(t, testCase.want, got)
			checkErr(t, testCase.wantErr, gotErr)
		})
	}
}

func checkErr(t *testing.T, want, got error) {
	t.Helper()

	if want == nil {
		assert.NoError(t, got)
	} else {
		assert.ErrorContains(t, got, want.Error())
	}
}
