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

func TestIsPretty(t *testing.T) {
	var cases = []struct {
		intention string
		input     *http.Request
		want      bool
	}{
		{
			"empty",
			httptest.NewRequest(http.MethodGet, "/?pretty", nil),
			true,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			if result := IsPretty(testCase.input); result != testCase.want {
				t.Errorf("IsPretty() = %t, want %t", result, testCase.want)
			}
		})
	}
}

func TestWrite(t *testing.T) {
	var cases = []struct {
		intention  string
		obj        interface{}
		pretty     bool
		want       string
		wantStatus int
		wantHeader map[string]string
	}{

		{
			"nil",
			nil,
			false,
			"null\n",
			http.StatusOK,
			map[string]string{"Content-Type": "application/json; charset=utf-8", "Cache-Control": "no-cache"},
		},
		{
			"simple object",
			testStruct{id: "Test"},
			false,
			"{\"Active\":false,\"Amount\":0}\n",
			http.StatusOK,
			map[string]string{"Content-Type": "application/json; charset=utf-8", "Cache-Control": "no-cache"},
		},
		{
			"pretty",
			testStruct{id: "Test", Active: true, Amount: 12.34},
			true,
			"{\n  \"Active\": true,\n  \"Amount\": 12.34\n}\n",
			http.StatusOK,
			map[string]string{"Content-Type": "application/json; charset=utf-8", "Cache-Control": "no-cache"},
		},
		{
			"error",
			func() string {
				return "test"
			},
			false,
			"Oops! Something went wrong. Server's logs contain more details.\n",
			http.StatusOK, // might not occur in real life
			map[string]string{"Content-Type": "application/json; charset=utf-8", "Cache-Control": "no-cache"},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			writer := httptest.NewRecorder()
			Write(writer, http.StatusOK, testCase.obj, testCase.pretty)

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

func BenchmarkWrite(b *testing.B) {
	var testCase = struct {
		obj interface{}
	}{
		testStruct{id: "Test", Active: true, Amount: 12.34},
	}

	writer := httptest.NewRecorder()

	for i := 0; i < b.N; i++ {
		Write(writer, http.StatusOK, testCase.obj, false)
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
			"{\"results\":null}\n",
			map[string]string{"Content-Type": "application/json; charset=utf-8", "Cache-Control": "no-cache"},
		},
		{
			"simple",
			[]testStruct{{id: "First", Active: true, Amount: 12.34}, {id: "Second", Active: true, Amount: 12.34}},
			"{\"results\":[{\"Active\":true,\"Amount\":12.34},{\"Active\":true,\"Amount\":12.34}]}\n",
			map[string]string{"Content-Type": "application/json; charset=utf-8", "Cache-Control": "no-cache"},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			writer := httptest.NewRecorder()
			WriteArray(writer, http.StatusOK, testCase.obj, false)

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

func TestWritePagination(t *testing.T) {
	var cases = []struct {
		intention  string
		page       uint
		pageSize   uint
		total      uint
		obj        interface{}
		want       string
		wantHeader map[string]string
	}{
		{
			"simple",
			1,
			2,
			2,
			[]testStruct{{id: "Test"}, {id: "Test", Active: true, Amount: 12.34}},
			"{\"results\":[{\"Active\":false,\"Amount\":0},{\"Active\":true,\"Amount\":12.34}],\"page\":1,\"pageSize\":2,\"pageCount\":1,\"total\":2}\n",
			map[string]string{"Content-Type": "application/json; charset=utf-8", "Cache-Control": "no-cache"},
		},
		{
			"compute page count rounded",
			1,
			10,
			40,
			[]testStruct{{id: "Test"}, {id: "Test", Active: true, Amount: 12.34}},
			"{\"results\":[{\"Active\":false,\"Amount\":0},{\"Active\":true,\"Amount\":12.34}],\"page\":1,\"pageSize\":10,\"pageCount\":4,\"total\":40}\n",
			map[string]string{"Content-Type": "application/json; charset=utf-8", "Cache-Control": "no-cache"},
		},
		{
			"compute page count exceed",
			1,
			10,
			45,
			[]testStruct{{id: "Test"}, {id: "Test", Active: true, Amount: 12.34}},
			"{\"results\":[{\"Active\":false,\"Amount\":0},{\"Active\":true,\"Amount\":12.34}],\"page\":1,\"pageSize\":10,\"pageCount\":5,\"total\":45}\n",
			map[string]string{"Content-Type": "application/json; charset=utf-8", "Cache-Control": "no-cache"},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			writer := httptest.NewRecorder()
			WritePagination(writer, http.StatusOK, testCase.page, testCase.pageSize, testCase.total, testCase.obj, false)

			if result, _ := request.ReadBodyResponse(writer.Result()); string(result) != testCase.want {
				t.Errorf("WritePagination() = `%s`, want `%s`", string(result), testCase.want)
			}

			for key, value := range testCase.wantHeader {
				if result, ok := writer.Result().Header[key]; !ok || strings.Join(result, "") != value {
					t.Errorf("WritePagination().Header[%s] = `%s`, want `%s`", key, strings.Join(result, ""), value)
				}
			}
		})
	}
}

type errReader int

func (errReader) Read(_ []byte) (int, error) {
	return 0, errors.New("read error")
}

func TestRead(t *testing.T) {
	type args struct {
		resp   *http.Response
		obj    interface{}
		action string
	}

	var cases = []struct {
		intention string
		args      args
		want      interface{}
		wantErr   error
	}{
		{
			"read error",
			args{
				resp: &http.Response{
					Body: io.NopCloser(errReader(0)),
				},
				action: "read error",
			},
			nil,
			errors.New("unable to read body response of read error"),
		},
		{
			"parse error",
			args{
				resp: &http.Response{
					Body: io.NopCloser(bytes.NewReader([]byte("invalid json"))),
				},
				action: "read error",
			},
			nil,
			errors.New("unable to parse body of read error"),
		},
		{
			"valid",
			args{
				resp: &http.Response{
					Body: io.NopCloser(bytes.NewReader([]byte(`{"key": "value","valid":true}`))),
				},
				obj:    make(map[string]interface{}),
				action: "valid",
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
			gotErr := Read(tc.args.resp, &tc.args.obj, tc.args.action)

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
