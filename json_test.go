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

func TestResponseJSON(t *testing.T) {
	var cases = []struct {
		obj        interface{}
		want       string
		wantStatus int
		wantHeader map[string]string
	}{
		{
			nil,
			`null`,
			http.StatusOK,
			map[string]string{`Content-Type`: `application/json`, `Cache-Control`: `no-cache`},
		},
		{
			testStruct{id: `Test`},
			`{"Active":false,"Amount":0}`,
			http.StatusOK,
			map[string]string{`Content-Type`: `application/json`, `Cache-Control`: `no-cache`},
		},
		{
			testStruct{id: `Test`, Active: true, Amount: 12.34},
			`{"Active":true,"Amount":12.34}`,
			http.StatusOK,
			map[string]string{`Content-Type`: `application/json`, `Cache-Control`: `no-cache`},
		},
		{
			testFn,
			`Error while marshalling JSON response: json: unsupported type: func() string
`,
			500,
			nil,
		},
	}

	for _, testCase := range cases {
		writer := httptest.NewRecorder()
		ResponseJSON(writer, http.StatusOK, testCase.obj)

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
		ResponseJSON(writer, http.StatusOK, testCase.obj)
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
		ResponseArrayJSON(writer, http.StatusOK, testCase.obj)

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
func Test_ResponsPaginatedJSON(t *testing.T) {
	var cases = []struct {
		intention  string
		obj        interface{}
		total      int64
		want       string
		wantStatus int
		wantHeader map[string]string
	}{
		{
			`should work with given params`,
			[]testStruct{{id: `Test`}, {id: `Test`, Active: true, Amount: 12.34}},
			1,
			`{"results":[{"Active":false,"Amount":0},{"Active":true,"Amount":12.34}],"total":1}`,
			http.StatusOK,
			map[string]string{`Content-Type`: `application/json`, `Cache-Control`: `no-cache`},
		},
	}

	for _, testCase := range cases {
		writer := httptest.NewRecorder()
		ResponsPaginatedJSON(writer, http.StatusOK, testCase.total, testCase.obj)

		if result := writer.Result().StatusCode; result != testCase.wantStatus {
			t.Errorf(`%s\ResponsPaginatedJSON(%v, %v) = %v, want %v`, testCase.intention, testCase.total, testCase.obj, result, testCase.wantStatus)
		}

		if result, _ := ReadBody(writer.Result().Body); string(result) != testCase.want {
			t.Errorf(`%s\ResponsPaginatedJSON(%v, %v) = %v, want %v`, testCase.intention, testCase.total, testCase.obj, string(result), testCase.want)
		}

		for key, value := range testCase.wantHeader {
			if result, ok := writer.Result().Header[key]; !ok || strings.Join(result, ``) != value {
				t.Errorf(`%s\ResponsPaginatedJSON(%v, %v).Header[%s] = %v, want %v`, testCase.intention, testCase.total, testCase.obj, key, strings.Join(result, ``), value)
			}
		}
	}
}
