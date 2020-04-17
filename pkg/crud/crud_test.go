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
			"Usage of simple:\n  -defaultPage uint\n    \t[crud] Default page {SIMPLE_DEFAULT_PAGE} (default 1)\n  -defaultPageSize uint\n    \t[crud] Default page size {SIMPLE_DEFAULT_PAGE_SIZE} (default 20)\n  -maxPageSize uint\n    \t[crud] Max page size {SIMPLE_MAX_PAGE_SIZE} (default 100)\n  -name string\n    \t[crud] Resource's name {SIMPLE_NAME}\n  -path string\n    \t[crud] HTTP Path {SIMPLE_PATH}\n",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			fs := flag.NewFlagSet(testCase.intention, flag.ContinueOnError)
			GetConfiguredFlags("", "")(fs, "")

			var writer strings.Builder
			fs.SetOutput(&writer)
			fs.Usage()

			result := writer.String()

			if result != testCase.want {
				t.Errorf("Flags() = `%s`, want `%s`", result, testCase.want)
			}
		})
	}
}

func TestNew(t *testing.T) {
	one := uint(1)
	name := "test"

	var cases = []struct {
		intention string
		config    Config
		service   Service
		want      App
		wantErr   error
	}{
		{
			"missing service",
			Config{},
			nil,
			nil,
			ErrServiceIsRequired,
		},
		{
			"missing service",
			Config{
				defaultPage:     &one,
				defaultPageSize: &one,
				maxPageSize:     &one,
				name:            &name,
				path:            &name,
			},
			testService{},
			&app{
				defaultPage:     1,
				defaultPageSize: 1,
				maxPageSize:     1,
				name:            "test",
				path:            "test",
				service:         testService{},
			},
			nil,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			result, err := New(testCase.config, testCase.service)

			failed := false

			if testCase.wantErr != nil && !errors.Is(err, testCase.wantErr) {
				failed = true
			} else if !reflect.DeepEqual(result, testCase.want) {
				failed = true
			}

			if failed {
				t.Errorf("New() = (%v, `%s`), want (%v, `%s`)", result, err, testCase.want, testCase.wantErr)
			}
		})
	}
}

func TestHandler(t *testing.T) {
	var cases = []struct {
		intention  string
		request    *http.Request
		want       string
		wantStatus int
	}{
		{
			"not allowed",
			httptest.NewRequest(http.MethodHead, "/", nil),
			"",
			http.StatusMethodNotAllowed,
		},
		{
			"list",
			httptest.NewRequest(http.MethodGet, "/", nil),
			"{\"results\":[{\"id\":1,\"name\":\"First\"},{\"id\":2,\"name\":\"First\"}],\"page\":0,\"pageSize\":2,\"pageCount\":5,\"total\":10}\n",
			http.StatusOK,
		},
		{
			"get invalid uint",
			httptest.NewRequest(http.MethodGet, "/-40", nil),
			"invalid unsigned integer value for ID\n",
			http.StatusBadRequest,
		},
		{
			"get",
			httptest.NewRequest(http.MethodGet, "/8000", nil),
			"{\"id\":8000,\"name\":\"Test\"}\n",
			http.StatusOK,
		},
		{
			"create",
			httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name": "success"}`)),
			"{\"id\":1,\"name\":\"success\"}\n",
			http.StatusCreated,
		},
		{
			"create with id",
			httptest.NewRequest(http.MethodPost, "/8000", nil),
			"",
			http.StatusMethodNotAllowed,
		},
		{
			"update root",
			httptest.NewRequest(http.MethodPut, "/", nil),
			"",
			http.StatusMethodNotAllowed,
		},
		{
			"update invalid uint",
			httptest.NewRequest(http.MethodPut, "/-40", nil),
			"invalid unsigned integer value for ID\n",
			http.StatusBadRequest,
		},
		{
			"update",
			httptest.NewRequest(http.MethodPut, "/8000", strings.NewReader(`{"id":8000,"name": "success"}`)),
			"{\"id\":8000,\"name\":\"success\"}\n",
			http.StatusOK,
		},
		{
			"delete root",
			httptest.NewRequest(http.MethodDelete, "/", nil),
			"",
			http.StatusMethodNotAllowed,
		},
		{
			"delete invalid uint",
			httptest.NewRequest(http.MethodDelete, "/-40", nil),
			"invalid unsigned integer value for ID\n",
			http.StatusBadRequest,
		},
		{
			"delete",
			httptest.NewRequest(http.MethodDelete, "/7000", nil),
			"",
			http.StatusNoContent,
		},
		{
			"options",
			httptest.NewRequest(http.MethodOptions, "/", nil),
			"",
			http.StatusNoContent,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			instance := app{
				service:         testService{},
				defaultPageSize: 2,
			}

			writer := httptest.NewRecorder()
			instance.Handler().ServeHTTP(writer, testCase.request)

			if result := writer.Code; result != testCase.wantStatus {
				t.Errorf("Handler() = %d, want %d", result, testCase.wantStatus)
			}

			if result, _ := request.ReadBodyResponse(writer.Result()); string(result) != testCase.want {
				t.Errorf("Handler() = `%s`, want `%s`", string(result), testCase.want)
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
			"unauthorized",
			ErrUnauthorized,
			true,
			http.StatusUnauthorized,
			"authentication required\n",
		},
		{
			"forbidden",
			ErrForbidden,
			true,
			http.StatusForbidden,
			"⛔️\n",
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
				t.Errorf("HandleError() = `%s`, want `%s`", result, testCase.wantContent)
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
		want      interface{}
		wantErr   error
	}{
		{
			"nil",
			nil,
			nil,
			errors.New("nil request"),
		},
		{
			"read error",
			&http.Request{
				Body:   ioutil.NopCloser(errReader(0)),
				Header: http.Header{},
			},
			nil,
			errors.New("body read error"),
		},
		{
			"valid",
			httptest.NewRequest(http.MethodGet, "/", strings.NewReader(`{"id": 1, "name": "test"}`)),
			&testItem{
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
			} else if !reflect.DeepEqual(result, testCase.want) {
				failed = true
			}

			if failed {
				t.Errorf("ReadPayload() = (%v, `%s`), want (%v, `%s`)", result, err, testCase.want, testCase.wantErr)
			}
		})
	}
}

func TestList(t *testing.T) {
	var cases = []struct {
		intention  string
		request    *http.Request
		want       string
		wantStatus int
	}{
		{
			"invalid pagination",
			httptest.NewRequest(http.MethodGet, "/?page=-1", nil),
			"page is invalid strconv.ParseUint: parsing \"-1\": invalid syntax: invalid value\n",
			http.StatusBadRequest,
		},
		{
			"error",
			httptest.NewRequest(http.MethodGet, "/?page=2", nil),
			"internal server error\n",
			http.StatusInternalServerError,
		},
		{
			"too far",
			httptest.NewRequest(http.MethodGet, "/?page=3", nil),
			"",
			http.StatusRequestedRangeNotSatisfiable,
		},
		{
			"valid",
			httptest.NewRequest(http.MethodGet, "/", nil),
			"{\"results\":[{\"id\":1,\"name\":\"First\"},{\"id\":2,\"name\":\"First\"}],\"page\":0,\"pageSize\":2,\"pageCount\":5,\"total\":10}\n",
			http.StatusOK,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			instance := app{
				service:         testService{},
				defaultPageSize: 2,
			}

			recorder := httptest.NewRecorder()
			instance.list(recorder, testCase.request)

			if result, _ := request.ReadBodyResponse(recorder.Result()); string(result) != testCase.want {
				t.Errorf("list() = `%s`, want `%s`", result, testCase.want)
			}

			if result := recorder.Result().StatusCode; result != testCase.wantStatus {
				t.Errorf("list() = %d, want %d", result, testCase.wantStatus)
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
				t.Errorf("get() = `%s`, want `%s`", result, testCase.want)
			}

			if result := recorder.Result().StatusCode; result != testCase.wantStatus {
				t.Errorf("get() = %d, want %d", result, testCase.wantStatus)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	var cases = []struct {
		intention  string
		request    *http.Request
		want       string
		wantStatus int
	}{
		{
			"unmarshal error",
			httptest.NewRequest(http.MethodPost, "/", nil),
			"unmarshal error: unexpected end of JSON input\n",
			http.StatusBadRequest,
		},
		{
			"invalid payload",
			httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"invalid"}`)),
			"{\"results\":[{\"field\":\"name\",\"label\":\"invalid name\"}]}\n",
			http.StatusBadRequest,
		},
		{
			"create error",
			httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"error"}`)),
			"internal server error\n",
			http.StatusInternalServerError,
		},
		{
			"create success",
			httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"success"}`)),
			"{\"id\":1,\"name\":\"success\"}\n",
			http.StatusCreated,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			instance := app{
				service: testService{},
			}

			writer := httptest.NewRecorder()
			instance.create(writer, testCase.request)

			if result, _ := request.ReadBodyResponse(writer.Result()); string(result) != testCase.want {
				t.Errorf("get() = `%s`, want `%s`", result, testCase.want)
			}

			if result := writer.Result().StatusCode; result != testCase.wantStatus {
				t.Errorf("get() = %d, want %d", result, testCase.wantStatus)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	var cases = []struct {
		intention  string
		request    *http.Request
		id         uint64
		want       string
		wantStatus int
	}{
		{
			"unmarshal error",
			httptest.NewRequest(http.MethodPut, "/", nil),
			0,
			"unmarshal error: unexpected end of JSON input\n",
			http.StatusBadRequest,
		},
		{
			"invalid payload",
			httptest.NewRequest(http.MethodPut, "/", strings.NewReader(`{"name":"invalid"}`)),
			0,
			"{\"results\":[{\"field\":\"name\",\"label\":\"invalid name\"}]}\n",
			http.StatusBadRequest,
		},
		{
			"update error",
			httptest.NewRequest(http.MethodPut, "/", strings.NewReader(`{"name":"error"}`)),
			0,
			"internal server error\n",
			http.StatusInternalServerError,
		},
		{
			"not found",
			httptest.NewRequest(http.MethodPut, "/", strings.NewReader(`{"name":"success"}`)),
			2000,
			"¯\\_(ツ)_/¯\n",
			http.StatusNotFound,
		},
		{
			"update success",
			httptest.NewRequest(http.MethodPut, "/", strings.NewReader(`{"id":8000,"name":"success"}`)),
			8000,
			"{\"id\":8000,\"name\":\"success\"}\n",
			http.StatusOK,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			instance := app{
				service: testService{},
			}

			writer := httptest.NewRecorder()
			instance.update(writer, testCase.request, testCase.id)

			if result, _ := request.ReadBodyResponse(writer.Result()); string(result) != testCase.want {
				t.Errorf("update() = `%s`, want `%s`", result, testCase.want)
			}

			if result := writer.Result().StatusCode; result != testCase.wantStatus {
				t.Errorf("update() = %d, want %d", result, testCase.wantStatus)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	var cases = []struct {
		intention  string
		request    *http.Request
		id         uint64
		want       string
		wantStatus int
	}{
		{
			"not found",
			httptest.NewRequest(http.MethodDelete, "/", nil),
			2000,
			"¯\\_(ツ)_/¯\n",
			http.StatusNotFound,
		},
		{
			"error",
			httptest.NewRequest(http.MethodDelete, "/", nil),
			6000,
			"{\"results\":[{\"field\":\"name\",\"label\":\"invalid name\"},{\"field\":\"value\",\"label\":\"invalid value\"}]}\n",
			http.StatusBadRequest,
		},
		{
			"delete error",
			httptest.NewRequest(http.MethodDelete, "/", nil),
			8000,
			"internal server error\n",
			http.StatusInternalServerError,
		},
		{
			"delete success",
			httptest.NewRequest(http.MethodDelete, "/", nil),
			7000,
			"",
			http.StatusNoContent,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			instance := app{
				service: testService{},
			}

			writer := httptest.NewRecorder()
			instance.delete(writer, testCase.request, testCase.id)

			if result, _ := request.ReadBodyResponse(writer.Result()); string(result) != testCase.want {
				t.Errorf("delete() = `%s`, want `%s`", result, testCase.want)
			}

			if result := writer.Result().StatusCode; result != testCase.wantStatus {
				t.Errorf("delete() = %d, want %d", result, testCase.wantStatus)
			}
		})
	}
}
