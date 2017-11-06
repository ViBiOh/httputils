package httputils

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
	var cases = []struct {
		intention string
		query     string
		want      bool
	}{
		{
			`should work with empty param`,
			``,
			false,
		},
		{
			`should work with pretty param`,
			`&&`,
			false,
		},
		{
			`should work with pretty param`,
			`pretty`,
			true,
		},
		{
			`should work with pretty param`,
			`test=1&pretty`,
			true,
		},
	}

	for _, testCase := range cases {
		if result := IsPretty(testCase.query); result != testCase.want {
			t.Errorf("%v\nIsPretty(%v) = %v, want %v", testCase.intention, testCase.query, result, testCase.want)
		}
	}
}

func TestResponseJSON(t *testing.T) {
	var cases = []struct {
		obj        interface{}
		pretty     bool
		want       string
		wantStatus int
		wantHeader map[string]string
	}{
		{
			nil,
			false,
			`null`,
			http.StatusOK,
			map[string]string{`Content-Type`: `application/json`, `Cache-Control`: `no-cache`},
		},
		{
			testStruct{id: `Test`},
			false,
			`{"Active":false,"Amount":0}`,
			http.StatusOK,
			map[string]string{`Content-Type`: `application/json`, `Cache-Control`: `no-cache`},
		},
		{
			testStruct{id: `Test`, Active: true, Amount: 12.34},
			true,
			`{
  "Active": true,
  "Amount": 12.34
}`,
			http.StatusOK,
			map[string]string{`Content-Type`: `application/json`, `Cache-Control`: `no-cache`},
		},
		{
			testFn,
			false,
			`Error while marshalling JSON response: json: unsupported type: func() string
`,
			500,
			nil,
		},
	}

	for _, testCase := range cases {
		writer := httptest.NewRecorder()
		ResponseJSON(writer, http.StatusOK, testCase.obj, testCase.pretty)

		if result := writer.Result().StatusCode; result != testCase.wantStatus {
			t.Errorf(`ResponseJSON(%v) = %v, want %v`, testCase.obj, result, testCase.wantStatus)
		}

		if result, _ := ReadBody(writer.Result().Body); string(result) != testCase.want {
			t.Errorf(`ResponseJSON(%v) = %v, want %v`, testCase.obj, string(result), testCase.want)
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
			map[string]string{`Content-Type`: `application/json`, `Cache-Control`: `no-cache`},
		},
		{
			[]testStruct{{id: `Test`}, {id: `Test`, Active: true, Amount: 12.34}},
			`{"results":[{"Active":false,"Amount":0},{"Active":true,"Amount":12.34}]}`,
			http.StatusOK,
			map[string]string{`Content-Type`: `application/json`, `Cache-Control`: `no-cache`},
		},
	}

	for _, testCase := range cases {
		writer := httptest.NewRecorder()
		ResponseArrayJSON(writer, http.StatusOK, testCase.obj, false)

		if result := writer.Result().StatusCode; result != testCase.wantStatus {
			t.Errorf(`ResponseJSON(%v) = %v, want %v`, testCase.obj, result, testCase.wantStatus)
		}

		if result, _ := ReadBody(writer.Result().Body); string(result) != testCase.want {
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
			map[string]string{`Content-Type`: `application/json`, `Cache-Control`: `no-cache`},
		},
		{
			`should calcul page count when pageSize match total`,
			1,
			10,
			40,
			[]testStruct{{id: `Test`}, {id: `Test`, Active: true, Amount: 12.34}},
			`{"results":[{"Active":false,"Amount":0},{"Active":true,"Amount":12.34}],"page":1,"pageSize":10,"pageCount":4,"total":40}`,
			http.StatusOK,
			map[string]string{`Content-Type`: `application/json`, `Cache-Control`: `no-cache`},
		},
		{
			`should calcul page count when pageSize don't match total`,
			1,
			10,
			45,
			[]testStruct{{id: `Test`}, {id: `Test`, Active: true, Amount: 12.34}},
			`{"results":[{"Active":false,"Amount":0},{"Active":true,"Amount":12.34}],"page":1,"pageSize":10,"pageCount":5,"total":45}`,
			http.StatusOK,
			map[string]string{`Content-Type`: `application/json`, `Cache-Control`: `no-cache`},
		},
	}

	for _, testCase := range cases {
		writer := httptest.NewRecorder()
		ResponsePaginatedJSON(writer, http.StatusOK, testCase.page, testCase.pageSize, testCase.total, testCase.obj, false)

		if result := writer.Result().StatusCode; result != testCase.wantStatus {
			t.Errorf(`%s\ResponsePaginatedJSON(%v, %v) = %v, want %v`, testCase.intention, testCase.total, testCase.obj, result, testCase.wantStatus)
		}

		if result, _ := ReadBody(writer.Result().Body); string(result) != testCase.want {
			t.Errorf(`%s\ResponsePaginatedJSON(%v, %v) = %v, want %v`, testCase.intention, testCase.total, testCase.obj, string(result), testCase.want)
		}

		for key, value := range testCase.wantHeader {
			if result, ok := writer.Result().Header[key]; !ok || strings.Join(result, ``) != value {
				t.Errorf(`%s\ResponsePaginatedJSON(%v, %v).Header[%s] = %v, want %v`, testCase.intention, testCase.total, testCase.obj, key, strings.Join(result, ``), value)
			}
		}
	}
}
