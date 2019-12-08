package httpjson

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ViBiOh/httputils/v3/pkg/request"
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
				t.Errorf("IsPretty() = %#v, want %#v", result, testCase.want)
			}
		})
	}
}

func TestResponseJSON(t *testing.T) {
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
			"json: unsupported type: func() string: cannot marshall json\n",
			http.StatusOK, // might not occur in real life
			map[string]string{"Content-Type": "application/json; charset=utf-8", "Cache-Control": "no-cache"},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			writer := httptest.NewRecorder()
			ResponseJSON(writer, http.StatusOK, testCase.obj, testCase.pretty)

			if result, _ := request.ReadBodyResponse(writer.Result()); string(result) != testCase.want {
				t.Errorf("ResponseJSON() = %s, want %s", string(result), testCase.want)
			}

			if result := writer.Result().StatusCode; result != testCase.wantStatus {
				t.Errorf("ResponseJSON() = %d, want %d", result, testCase.wantStatus)
			}

			for key, value := range testCase.wantHeader {
				if result, ok := writer.Result().Header[key]; !ok || strings.Join(result, "") != value {
					t.Errorf("ResponseJSON().Header[%s] = %s, want %s", key, strings.Join(result, ""), value)
				}
			}
		})
	}
}

func BenchmarkResponseJSON(b *testing.B) {
	var testCase = struct {
		obj interface{}
	}{
		testStruct{id: "Test", Active: true, Amount: 12.34},
	}

	writer := httptest.NewRecorder()

	for i := 0; i < b.N; i++ {
		ResponseJSON(writer, http.StatusOK, testCase.obj, false)
	}
}

func TestResponseArrayJSON(t *testing.T) {
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
			ResponseArrayJSON(writer, http.StatusOK, testCase.obj, false)

			if result, _ := request.ReadBodyResponse(writer.Result()); string(result) != testCase.want {
				t.Errorf("TestResponseArrayJSON() = %s, want %s", string(result), testCase.want)
			}

			for key, value := range testCase.wantHeader {
				if result, ok := writer.Result().Header[key]; !ok || strings.Join(result, "") != value {
					t.Errorf("TestResponseArrayJSON().Header[%s] = %s, want %s", key, strings.Join(result, ""), value)
				}
			}
		})
	}
}

func TestResponsePaginatedJSON(t *testing.T) {
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
			ResponsePaginatedJSON(writer, http.StatusOK, testCase.page, testCase.pageSize, testCase.total, testCase.obj, false)

			if result, _ := request.ReadBodyResponse(writer.Result()); string(result) != testCase.want {
				t.Errorf("ResponsePaginatedJSON() = %s, want %s", string(result), testCase.want)
			}

			for key, value := range testCase.wantHeader {
				if result, ok := writer.Result().Header[key]; !ok || strings.Join(result, "") != value {
					t.Errorf("ResponsePaginatedJSON().Header[%s] = %s, want %s", key, strings.Join(result, ""), value)
				}
			}
		})
	}
}
