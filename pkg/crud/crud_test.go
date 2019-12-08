package crud

import (
	"errors"
	"flag"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/ViBiOh/httputils/v3/pkg/request"
)

type errReader int

func (errReader) Read(p []byte) (int, error) {
	return 0, errors.New("read error")
}

func TestFlags(t *testing.T) {
	var cases = []struct {
		intention string
		want      string
	}{
		{
			"simple",
			"Usage of simple:\n  -defaultPage uint\n    \t[crud] Default page {SIMPLE_DEFAULT_PAGE} (default 1)\n  -defaultPageSize uint\n    \t[crud] Default page size {SIMPLE_DEFAULT_PAGE_SIZE} (default 20)\n  -maxPageSize uint\n    \t[crud] Max page size {SIMPLE_MAX_PAGE_SIZE} (default 100)\n",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			fs := flag.NewFlagSet(testCase.intention, flag.ContinueOnError)
			Flags(fs, "")

			var writer strings.Builder
			fs.SetOutput(&writer)
			fs.Usage()

			result := writer.String()

			if result != testCase.want {
				t.Errorf("Flags() = %s, want %s", result, testCase.want)
			}
		})
	}
}

func TestHandleError(t *testing.T) {
	var cases = []struct {
		intention   string
		error       error
		want        bool
		wantStatus  int
		wantContent string
	}{
		{
			"no error",
			nil,
			false,
			http.StatusOK,
			"",
		},
		{
			"invalid",
			ErrInvalid,
			true,
			http.StatusBadRequest,
			"invalid\n",
		},
		{
			"invalid",
			ErrNotFound,
			true,
			http.StatusNotFound,
			"¯\\_(ツ)_/¯\n",
		},
		{
			"invalid",
			errors.New("unable to handle request"),
			true,
			http.StatusInternalServerError,
			"internal server error\n",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			recorder := httptest.NewRecorder()

			if result := handleError(recorder, testCase.error); result != testCase.want {
				t.Errorf("HandleError() = %t, want %t", result, testCase.want)
			}

			if result := recorder.Result().StatusCode; result != testCase.wantStatus {
				t.Errorf("HandleError() = %d, want %d", result, testCase.wantStatus)
			}

			if result, _ := ioutil.ReadAll(recorder.Body); string(result) != testCase.wantContent {
				t.Errorf("HandleError() = %s, want %s", result, testCase.wantContent)
			}
		})
	}
}

func TestWriteErrors(t *testing.T) {
	var cases = []struct {
		intention  string
		input      []error
		want       string
		wantStatus int
	}{
		{
			"empty",
			nil,
			"invalid payload:\n",
			http.StatusBadRequest,
		},
		{
			"multiple errors",
			[]error{
				errors.New("invalid name"),
				errors.New("invalid email"),
			},
			"invalid payload:\n\tinvalid name\n\tinvalid email\n",
			http.StatusBadRequest,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			writer := httptest.NewRecorder()
			writeErrors(writer, testCase.input)

			if result := writer.Code; result != testCase.wantStatus {
				t.Errorf("writeErrors = %d, want %d", result, testCase.wantStatus)
			}

			if result, _ := request.ReadBodyResponse(writer.Result()); string(result) != testCase.want {
				t.Errorf("writeErrors = %s, want %s", string(result), testCase.want)
			}
		})
	}
}

func TestReadFilters(t *testing.T) {
	var cases = []struct {
		intention string
		input     *http.Request
		want      map[string][]string
	}{
		{
			"empty",
			httptest.NewRequest(http.MethodGet, "/", nil),
			make(map[string][]string, 0),
		},
		{
			"parse error",
			&http.Request{
				URL: &url.URL{
					RawQuery: "/%1",
				},
			},
			nil,
		},
		{
			"remove reserved params",
			httptest.NewRequest(http.MethodGet, "/?page=1&pageSize=10&name=Test", nil),
			map[string][]string{
				"name": {"Test"},
			},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			if result := readFilters(testCase.input); !reflect.DeepEqual(result, testCase.want) {
				t.Errorf("ReadFilters() = %#v, want %#v", result, testCase.want)
			}
		})
	}
}

func TestReadPayload(t *testing.T) {
	var cases = []struct {
		intention string
		input     *http.Request
		want      Item
		wantErr   error
	}{
		{
			"read error",
			&http.Request{
				Body: ioutil.NopCloser(errReader(0)),
			},
			nil,
			errors.New("body read error"),
		},
		{
			"nil",
			nil,
			testItem{},
			errors.New("unmarshall error"),
		},
		{
			"valid",
			httptest.NewRequest(http.MethodGet, "/", strings.NewReader(`{"id": 1, "name": "test"}`)),
			testItem{
				ID:   1,
				Name: "test",
			},
			nil,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			instance := app{
				service: testService{},
			}

			result, err := instance.readPayload(testCase.input)

			failed := false

			if testCase.wantErr != nil && err == nil {
				failed = true
			} else if testCase.wantErr == nil && err != nil {
				failed = true
			} else if testCase.wantErr != nil && !strings.HasPrefix(err.Error(), testCase.wantErr.Error()) {
				failed = true
			} else if result != testCase.want {
				failed = true
			}

			if failed {
				t.Errorf("ReadPayload() = (%v, %s), want (%v, %s)", result, err, testCase.want, testCase.wantErr)
			}
		})
	}
}

func TestGet(t *testing.T) {
	var cases = []struct {
		intention  string
		request    *http.Request
		id         uint64
		want       string
		wantStatus int
	}{
		{
			"null",
			httptest.NewRequest(http.MethodGet, "/", nil),
			0,
			"null\n",
			http.StatusOK,
		},
		{
			"not found",
			httptest.NewRequest(http.MethodGet, "/", nil),
			2000,
			"¯\\_(ツ)_/¯\n",
			http.StatusNotFound,
		},
		{
			"internal error",
			httptest.NewRequest(http.MethodGet, "/", nil),
			4000,
			"internal server error\n",
			http.StatusInternalServerError,
		},
		{
			"valid",
			httptest.NewRequest(http.MethodGet, "/", nil),
			8000,
			"{\"id\":8000,\"name\":\"Test\"}\n",
			http.StatusOK,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			instance := app{
				service: testService{},
			}

			recorder := httptest.NewRecorder()
			instance.get(recorder, testCase.request, testCase.id)

			if result, _ := request.ReadBodyResponse(recorder.Result()); string(result) != testCase.want {
				t.Errorf("get() = %s, want %s", result, testCase.want)
			}

			if result := recorder.Result().StatusCode; result != testCase.wantStatus {
				t.Errorf("get() = %d, want %d", result, testCase.wantStatus)
			}
		})
	}
}
