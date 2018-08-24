package httpjson

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ViBiOh/httputils/pkg/request"
)

type testStruct struct {
	id     string
	Active bool
	Amount float64
}

func testFn() string {
	return `toto`
}

func Test_IsPretty(t *testing.T) {
	emptyRequest, _ := http.NewRequest(http.MethodGet, `http://localhost`, nil)
	prettyRequest, _ := http.NewRequest(http.MethodGet, `http://localhost?pretty`, nil)
	prettyValueRequest, _ := http.NewRequest(http.MethodGet, `http://localhost?test=1&pretty=false`, nil)
	prettyInvalidRequest, _ := http.NewRequest(http.MethodGet, `http://localhost?test=1&pretty=invalidBool`, nil)

	var cases = []struct {
		intention string
		request     *http.Request
		want      bool
	}{
		{
			`should work with empty param`,
			emptyRequest,
			false,
		},
		{
			`should work with pretty param`,
			prettyRequest,
			true,
		},
		{
			`should work with pretty value`,
			prettyValueRequest,
			false,
		},
		{
			`should work with pretty value not equal to a boolean`,
			prettyInvalidRequest,
			true,
		},
	}

	for _, testCase := range cases {
		if result := IsPretty(testCase.request); result != testCase.want {
			t.Errorf("%v\nIsPretty(%v) = %v, want %v", testCase.intention, testCase.request, result, testCase.want)
		}
	}
}

func Test_ResponseJSON(t *testing.T) {
	var cases = []struct {
		intention  string
		obj        interface{}
		pretty     bool
		want       string
		wantHeader map[string]string
		wantErr    error
	}{

		{
			`should work with nil obj`,
			nil,
			false,
			`null`,
			map[string]string{`Content-Type`: `application/json; charset=utf-8`, `Cache-Control`: `no-cache`},
			nil,
		},
		{
			`should work with given obj`,
			testStruct{id: `Test`},
			false,
			`{"Active":false,"Amount":0}`,
			map[string]string{`Content-Type`: `application/json; charset=utf-8`, `Cache-Control`: `no-cache`},
			nil,
		},
		{
			`should work with pretty print`,
			testStruct{id: `Test`, Active: true, Amount: 12.34},
			true,
			`{
  "Active": true,
  "Amount": 12.34
}`,
			map[string]string{`Content-Type`: `application/json; charset=utf-8`, `Cache-Control`: `no-cache`},
			nil,
		},
		{
			`should work with error print`,
			testFn,
			false,
			``,
			nil,
			errors.New(`Error while marshalling JSON response: json: unsupported type: func() string`),
		},
	}

	var failed bool

	for _, testCase := range cases {
		writer := httptest.NewRecorder()
		err := ResponseJSON(writer, http.StatusOK, testCase.obj, testCase.pretty)

		rawResult, _ := request.ReadBody(writer.Result().Body)
		result := string(rawResult)

		failed = false

		if err == nil && testCase.wantErr != nil {
			failed = true
		} else if err != nil && testCase.wantErr == nil {
			failed = true
		} else if err != nil && err.Error() != testCase.wantErr.Error() {
			failed = true
		} else if result != testCase.want {
			failed = true
		}

		if failed {
			t.Errorf("%s\nResponseJSON(%+v, %+v) = (%+v, %+v), want (%+v, %+v)", testCase.intention, testCase.obj, testCase.pretty, result, err, testCase.want, testCase.wantErr)
		}

		for key, value := range testCase.wantHeader {
			if result, ok := writer.Result().Header[key]; !ok || strings.Join(result, ``) != value {
				t.Errorf(`ResponseJSON(%v).Header[%s] = %v, want %v`, testCase.obj, key, strings.Join(result, ``), value)
			}
		}
	}
}

func BenchmarkResponseJSON(b *testing.B) {
	var testCase = struct {
		obj interface{}
	}{
		testStruct{id: `Test`},
	}

	writer := httptest.NewRecorder()

	for i := 0; i < b.N; i++ {
		ResponseJSON(writer, http.StatusOK, testCase.obj, false)
	}
}

func TestResponseArrayJSON(t *testing.T) {
	var cases = []struct {
		obj        interface{}
		want       string
		wantStatus int
		wantHeader map[string]string
	}{
		{
			nil,
			`{"results":null}`,
			http.StatusOK,
			map[string]string{`Content-Type`: `application/json; charset=utf-8`, `Cache-Control`: `no-cache`},
		},
		{
			[]testStruct{{id: `Test`}, {id: `Test`, Active: true, Amount: 12.34}},
			`{"results":[{"Active":false,"Amount":0},{"Active":true,"Amount":12.34}]}`,
			http.StatusOK,
			map[string]string{`Content-Type`: `application/json; charset=utf-8`, `Cache-Control`: `no-cache`},
		},
	}

	for _, testCase := range cases {
		writer := httptest.NewRecorder()
		ResponseArrayJSON(writer, http.StatusOK, testCase.obj, false)

		if result := writer.Result().StatusCode; result != testCase.wantStatus {
			t.Errorf(`ResponseJSON(%v) = %v, want %v`, testCase.obj, result, testCase.wantStatus)
		}

		if result, _ := request.ReadBody(writer.Result().Body); string(result) != testCase.want {
			t.Errorf(`ResponseJSON(%v) = %v, want %v`, testCase.obj, string(result), testCase.want)
		}

		for key, value := range testCase.wantHeader {
			if result, ok := writer.Result().Header[key]; !ok || strings.Join(result, ``) != value {
				t.Errorf(`ResponseJSON(%v).Header[%s] = %v, want %v`, testCase.obj, key, strings.Join(result, ``), value)
			}
		}
	}
}

func Test_ResponsePaginatedJSON(t *testing.T) {
	var cases = []struct {
		intention  string
		page       uint
		pageSize   uint
		total      uint
		obj        interface{}
		want       string
		wantStatus int
		wantHeader map[string]string
	}{
		{
			`should work with given params`,
			1,
			2,
			2,
			[]testStruct{{id: `Test`}, {id: `Test`, Active: true, Amount: 12.34}},
			`{"results":[{"Active":false,"Amount":0},{"Active":true,"Amount":12.34}],"page":1,"pageSize":2,"pageCount":1,"total":2}`,
			http.StatusOK,
			map[string]string{`Content-Type`: `application/json; charset=utf-8`, `Cache-Control`: `no-cache`},
		},
		{
			`should calcul page count when pageSize match total`,
			1,
			10,
			40,
			[]testStruct{{id: `Test`}, {id: `Test`, Active: true, Amount: 12.34}},
			`{"results":[{"Active":false,"Amount":0},{"Active":true,"Amount":12.34}],"page":1,"pageSize":10,"pageCount":4,"total":40}`,
			http.StatusOK,
			map[string]string{`Content-Type`: `application/json; charset=utf-8`, `Cache-Control`: `no-cache`},
		},
		{
			`should calcul page count when pageSize don't match total`,
			1,
			10,
			45,
			[]testStruct{{id: `Test`}, {id: `Test`, Active: true, Amount: 12.34}},
			`{"results":[{"Active":false,"Amount":0},{"Active":true,"Amount":12.34}],"page":1,"pageSize":10,"pageCount":5,"total":45}`,
			http.StatusOK,
			map[string]string{`Content-Type`: `application/json; charset=utf-8`, `Cache-Control`: `no-cache`},
		},
	}

	for _, testCase := range cases {
		writer := httptest.NewRecorder()
		ResponsePaginatedJSON(writer, http.StatusOK, testCase.page, testCase.pageSize, testCase.total, testCase.obj, false)

		if result := writer.Result().StatusCode; result != testCase.wantStatus {
			t.Errorf(`%s\ResponsePaginatedJSON(%v, %v) = %v, want %v`, testCase.intention, testCase.total, testCase.obj, result, testCase.wantStatus)
		}

		if result, _ := request.ReadBody(writer.Result().Body); string(result) != testCase.want {
			t.Errorf(`%s\ResponsePaginatedJSON(%v, %v) = %v, want %v`, testCase.intention, testCase.total, testCase.obj, string(result), testCase.want)
		}

		for key, value := range testCase.wantHeader {
			if result, ok := writer.Result().Header[key]; !ok || strings.Join(result, ``) != value {
				t.Errorf(`%s\ResponsePaginatedJSON(%v, %v).Header[%s] = %v, want %v`, testCase.intention, testCase.total, testCase.obj, key, strings.Join(result, ``), value)
			}
		}
	}
}
