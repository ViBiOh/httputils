package httputils

import (
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
			200,
			map[string]string{`Content-Type`: `application/json`, `Cache-Control`: `no-cache`},
		},
		{
			testStruct{id: `Test`},
			`{"Active":false,"Amount":0}`,
			200,
			map[string]string{`Content-Type`: `application/json`, `Cache-Control`: `no-cache`},
		},
		{
			testStruct{id: `Test`, Active: true, Amount: 12.34},
			`{"Active":true,"Amount":12.34}`,
			200,
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
		ResponseJSON(writer, testCase.obj)

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
		ResponseJSON(writer, testCase.obj)
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
			200,
			map[string]string{`Content-Type`: `application/json`, `Cache-Control`: `no-cache`},
		},
		{
			[]testStruct{{id: `Test`}, {id: `Test`, Active: true, Amount: 12.34}},
			`{"results":[{"Active":false,"Amount":0},{"Active":true,"Amount":12.34}]}`,
			200,
			map[string]string{`Content-Type`: `application/json`, `Cache-Control`: `no-cache`},
		},
	}

	for _, testCase := range cases {
		writer := httptest.NewRecorder()
		ResponseArrayJSON(writer, testCase.obj)

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
