package httpjson

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/ViBiOh/httputils/v4/pkg/request"
)

type testStruct struct {
	id     string
	Active bool
	Amount float64
}

func TestRawWrite(t *testing.T) {
	type args struct {
		writer *bytes.Buffer
		obj    interface{}
	}

	var cases = []struct {
		intention string
		args      args
		want      string
		wantErr   error
	}{
		{
			"invalid",
			args{
				writer: bytes.NewBufferString(""),
				obj:    func() {},
			},
			"",
			ErrCannotMarshal,
		},
		{
			"simple",
			args{
				writer: bytes.NewBufferString(""),
				obj: map[string]interface{}{
					"key":   "value",
					"valid": true,
				},
			},
			"{\"key\":\"value\",\"valid\":true}\n",
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			gotErr := RawWrite(tc.args.writer, tc.args.obj)

			failed := false

			if tc.wantErr == nil && gotErr != nil {
				failed = true
			} else if tc.wantErr != nil && gotErr == nil {
				failed = true
			} else if tc.wantErr != nil && !strings.Contains(gotErr.Error(), tc.wantErr.Error()) {
				failed = true
			} else if tc.args.writer.String() != tc.want {
				failed = true
			}

			if failed {
				t.Errorf("RawWrite() = (`%s`, `%s`), want (`%s`, `%s`)", tc.args.writer.String(), gotErr, tc.want, tc.wantErr)
			}
		})
	}
}

func TestWrite(t *testing.T) {
	var cases = []struct {
		intention  string
		obj        interface{}
		want       string
		wantStatus int
		wantHeader map[string]string
	}{

		{
			"nil",
			nil,
			"null\n",
			http.StatusOK,
			map[string]string{"Content-Type": "application/json; charset=utf-8", "Cache-Control": "no-cache"},
		},
		{
			"simple object",
			testStruct{id: "Test"},
			"{\"Active\":false,\"Amount\":0}\n",
			http.StatusOK,
			map[string]string{"Content-Type": "application/json; charset=utf-8", "Cache-Control": "no-cache"},
		},
		{
			"error",
			func() string {
				return "test"
			},
			"Oops! Something went wrong. Server's logs contain more details.\n",
			http.StatusOK, // might not occur in real life
			map[string]string{"Content-Type": "application/json; charset=utf-8", "Cache-Control": "no-cache"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			writer := httptest.NewRecorder()
			Write(writer, http.StatusOK, tc.obj)

			if result, _ := request.ReadBodyResponse(writer.Result()); string(result) != tc.want {
				t.Errorf("Write() = `%s`, want `%s`", string(result), tc.want)
			}

			if result := writer.Result().StatusCode; result != tc.wantStatus {
				t.Errorf("Write() = %d, want %d", result, tc.wantStatus)
			}

			for key, value := range tc.wantHeader {
				if result, ok := writer.Result().Header[key]; !ok || strings.Join(result, "") != value {
					t.Errorf("Write().Header[%s] = `%s`, want `%s`", key, strings.Join(result, ""), value)
				}
			}
		})
	}
}

func BenchmarkRawWrite(b *testing.B) {
	obj := testStruct{id: "Test", Active: true, Amount: 12.34}

	for i := 0; i < b.N; i++ {
		if err := RawWrite(io.Discard, &obj); err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkWrite(b *testing.B) {
	var tc = struct {
		obj interface{}
	}{
		testStruct{id: "Test", Active: true, Amount: 12.34},
	}

	writer := httptest.NewRecorder()

	for i := 0; i < b.N; i++ {
		Write(writer, http.StatusOK, &tc.obj)
	}
}

func TestWriteArray(t *testing.T) {
	var cases = []struct {
		intention  string
		obj        interface{}
		want       string
		wantHeader map[string]string
	}{
		{
			"nil",
			nil,
			"{\"items\":null}\n",
			map[string]string{"Content-Type": "application/json; charset=utf-8", "Cache-Control": "no-cache"},
		},
		{
			"simple",
			[]testStruct{{id: "First", Active: true, Amount: 12.34}, {id: "Second", Active: true, Amount: 12.34}},
			"{\"items\":[{\"Active\":true,\"Amount\":12.34},{\"Active\":true,\"Amount\":12.34}]}\n",
			map[string]string{"Content-Type": "application/json; charset=utf-8", "Cache-Control": "no-cache"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			writer := httptest.NewRecorder()
			WriteArray(writer, http.StatusOK, tc.obj)

			if result, _ := request.ReadBodyResponse(writer.Result()); string(result) != tc.want {
				t.Errorf("TestWriteArray() = `%s`, want `%s`", string(result), tc.want)
			}

			for key, value := range tc.wantHeader {
				if result, ok := writer.Result().Header[key]; !ok || strings.Join(result, "") != value {
					t.Errorf("TestWriteArray().Header[%s] = `%s`, want `%s`", key, strings.Join(result, ""), value)
				}
			}
		})
	}
}

func TestWritePagination(t *testing.T) {
	var cases = []struct {
		intention  string
		pageSize   uint
		total      uint
		last       string
		obj        interface{}
		want       string
		wantHeader map[string]string
	}{
		{
			"simple",
			2,
			2,
			"8000",
			[]testStruct{{id: "Test"}, {id: "Test", Active: true, Amount: 12.34}},
			"{\"items\":[{\"Active\":false,\"Amount\":0},{\"Active\":true,\"Amount\":12.34}],\"last\":\"8000\",\"pageSize\":2,\"pageCount\":1,\"total\":2}\n",
			map[string]string{"Content-Type": "application/json; charset=utf-8", "Cache-Control": "no-cache"},
		},
		{
			"compute page count rounded",
			10,
			40,
			"8000",
			[]testStruct{{id: "Test"}, {id: "Test", Active: true, Amount: 12.34}},
			"{\"items\":[{\"Active\":false,\"Amount\":0},{\"Active\":true,\"Amount\":12.34}],\"last\":\"8000\",\"pageSize\":10,\"pageCount\":4,\"total\":40}\n",
			map[string]string{"Content-Type": "application/json; charset=utf-8", "Cache-Control": "no-cache"},
		},
		{
			"compute page count exceed",
			10,
			45,
			"8000",
			[]testStruct{{id: "Test"}, {id: "Test", Active: true, Amount: 12.34}},
			"{\"items\":[{\"Active\":false,\"Amount\":0},{\"Active\":true,\"Amount\":12.34}],\"last\":\"8000\",\"pageSize\":10,\"pageCount\":5,\"total\":45}\n",
			map[string]string{"Content-Type": "application/json; charset=utf-8", "Cache-Control": "no-cache"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			writer := httptest.NewRecorder()
			WritePagination(writer, http.StatusOK, tc.pageSize, tc.total, tc.last, tc.obj)

			if result, _ := request.ReadBodyResponse(writer.Result()); string(result) != tc.want {
				t.Errorf("WritePagination() = `%s`, want `%s`", string(result), tc.want)
			}

			for key, value := range tc.wantHeader {
				if result, ok := writer.Result().Header[key]; !ok || strings.Join(result, "") != value {
					t.Errorf("WritePagination().Header[%s] = `%s`, want `%s`", key, strings.Join(result, ""), value)
				}
			}
		})
	}
}

func TestParse(t *testing.T) {
	type args struct {
		req *http.Request
		obj interface{}
	}

	var cases = []struct {
		intention string
		args      args
		want      interface{}
		wantErr   error
	}{
		{
			"valid",
			args{
				req: httptest.NewRequest(http.MethodGet, "/", bytes.NewBuffer([]byte(`{"key": "value","valid":true}`))),
				obj: make(map[string]interface{}),
			},
			map[string]interface{}{
				"key":   "value",
				"valid": true,
			},
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			gotErr := Parse(tc.args.req, &tc.args.obj)

			failed := false

			if tc.wantErr == nil && gotErr != nil {
				failed = true
			} else if tc.wantErr != nil && gotErr == nil {
				failed = true
			} else if tc.wantErr != nil && !strings.Contains(gotErr.Error(), tc.wantErr.Error()) {
				failed = true
			} else if !reflect.DeepEqual(tc.args.obj, tc.want) {
				failed = true
			}

			if failed {
				t.Errorf("Parse() = (%+v, `%s`), want (%+v, `%s`)", tc.args.obj, gotErr, tc.want, tc.wantErr)
			}
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
	type args struct {
		resp *http.Response
		obj  interface{}
	}

	var cases = []struct {
		intention string
		args      args
		want      interface{}
		wantErr   error
	}{
		{
			"parse error",
			args{
				resp: &http.Response{
					Body: io.NopCloser(bytes.NewReader([]byte("invalid json"))),
				},
			},
			nil,
			errors.New("unable to parse JSON body"),
		},
		{
			"close error",
			args{
				resp: &http.Response{
					Body: errCloser{bytes.NewReader([]byte(`{"key": "value","valid":true}`))},
				},
				obj: make(map[string]interface{}),
			},
			map[string]interface{}{
				"key":   "value",
				"valid": true,
			},
			errors.New("close error"),
		},
		{
			"both error",
			args{
				resp: &http.Response{
					Body: errCloser{bytes.NewReader([]byte(`invalid json`))},
				},
			},
			nil,
			errors.New("unable to parse JSON body: invalid character 'i' looking for beginning of value: close error"),
		},
		{
			"valid",
			args{
				resp: &http.Response{
					Body: io.NopCloser(bytes.NewReader([]byte(`{"key": "value","valid":true}`))),
				},
				obj: make(map[string]interface{}),
			},
			map[string]interface{}{
				"key":   "value",
				"valid": true,
			},
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			gotErr := Read(tc.args.resp, &tc.args.obj)

			failed := false

			if tc.wantErr == nil && gotErr != nil {
				failed = true
			} else if tc.wantErr != nil && gotErr == nil {
				failed = true
			} else if tc.wantErr != nil && !strings.Contains(gotErr.Error(), tc.wantErr.Error()) {
				failed = true
			} else if !reflect.DeepEqual(tc.args.obj, tc.want) {
				failed = true
			}

			if failed {
				t.Errorf("Read() = (%+v, `%s`), want (%+v, `%s`)", tc.args.obj, gotErr, tc.want, tc.wantErr)
			}
		})
	}
}

func TestStream(t *testing.T) {
	newObj := func() interface{} {
		return new(string)
	}

	type args struct {
		stream io.Reader
		key    string
	}

	var cases = []struct {
		intention string
		args      args
		want      []string
		wantErr   error
	}{
		{
			"invalid json",
			args{
				stream: strings.NewReader("invalid json"),
				key:    "items",
			},
			nil,
			errors.New("unable to read token"),
		},
		{
			"no opening token",
			args{
				stream: strings.NewReader(`{"count": 10, "items"}`),
				key:    "items",
			},
			nil,
			errors.New("unable to read opening token"),
		},
		{
			"no closing token",
			args{
				stream: strings.NewReader(`{"count": 10, "items": ["test", "next", "final"}`),
				key:    "items",
			},
			[]string{"test", "next", "final"},
			errors.New("unable to read closing token"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			output := make(chan interface{}, 4)
			done := make(chan struct{})
			var got []string

			go func() {
				defer close(done)
				for item := range output {
					got = append(got, *(item.(*string)))
				}
			}()

			gotErr := Stream(tc.args.stream, newObj, output, tc.args.key)

			<-done

			failed := false

			if tc.wantErr == nil && gotErr != nil {
				failed = true
			} else if tc.wantErr != nil && gotErr == nil {
				failed = true
			} else if tc.wantErr != nil && !strings.Contains(gotErr.Error(), tc.wantErr.Error()) {
				failed = true
			} else if !reflect.DeepEqual(got, tc.want) {
				failed = true
			}

			if failed {
				t.Errorf("Stream() = (%+v, `%s`), want (%+v, `%s`)", got, gotErr, tc.want, tc.wantErr)
			}
		})
	}
}
