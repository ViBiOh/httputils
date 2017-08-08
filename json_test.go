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

func TestResponseJSON(t *testing.T) {
	var tests = []struct {
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
	}

	for _, test := range tests {
		writer := httptest.NewRecorder()
		ResponseJSON(writer, test.obj)

		if result := writer.Result().StatusCode; result != test.wantStatus {
			t.Errorf(`ResponseJSON(%v) = %v, want %v`, test.obj, result, test.wantStatus)
		}

		if result, _ := ReadBody(writer.Result().Body); string(result) != test.want {
			t.Errorf(`ResponseJSON(%v) = %v, want %v`, test.obj, string(result), test.want)
		}

		for key, value := range test.wantHeader {
			if result, ok := writer.Result().Header[key]; !ok || strings.Join(result, ``) != value {
				t.Errorf(`ResponseJSON(%v).Header[%s] = %v, want %v`, test.obj, key, strings.Join(result, ``), value)
			}
		}
	}
}

func TestResponseArrayJSON(t *testing.T) {
	var tests = []struct {
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
			[]testStruct{testStruct{id: `Test`}, testStruct{id: `Test`, Active: true, Amount: 12.34}},
			`{"results":[{"Active":false,"Amount":0},{"Active":true,"Amount":12.34}]}`,
			200,
			map[string]string{`Content-Type`: `application/json`, `Cache-Control`: `no-cache`},
		},
	}

	for _, test := range tests {
		writer := httptest.NewRecorder()
		ResponseArrayJSON(writer, test.obj)

		if result := writer.Result().StatusCode; result != test.wantStatus {
			t.Errorf(`ResponseJSON(%v) = %v, want %v`, test.obj, result, test.wantStatus)
		}

		if result, _ := ReadBody(writer.Result().Body); string(result) != test.want {
			t.Errorf(`ResponseJSON(%v) = %v, want %v`, test.obj, string(result), test.want)
		}

		for key, value := range test.wantHeader {
			if result, ok := writer.Result().Header[key]; !ok || strings.Join(result, ``) != value {
				t.Errorf(`ResponseJSON(%v).Header[%s] = %v, want %v`, test.obj, key, strings.Join(result, ``), value)
			}
		}
	}
}
