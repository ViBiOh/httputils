package query

import (
	"net/http"
	"testing"
)

func Test_GetBool(t *testing.T) {
	emptyRequest, _ := http.NewRequest(http.MethodGet, "http://localhost", nil)
	validRequest, _ := http.NewRequest(http.MethodGet, "http://localhost?valid", nil)
	validValueRequest, _ := http.NewRequest(http.MethodGet, "http://localhost?test=1&valid=false", nil)
	validInvalidRequest, _ := http.NewRequest(http.MethodGet, "http://localhost?test=1&valid=invalidBool", nil)

	var cases = []struct {
		intention string
		request   *http.Request
		name      string
		want      bool
	}{
		{
			"should work with empty param",
			emptyRequest,
			"valid",
			false,
		},
		{
			"should work with valid param",
			validRequest,
			"valid",
			true,
		},
		{
			"should work with valid value",
			validValueRequest,
			"valid",
			false,
		},
		{
			"should work with valid value not equal to a boolean",
			validInvalidRequest,
			"valid",
			false,
		},
	}

	for _, testCase := range cases {
		if result := GetBool(testCase.request, testCase.name); result != testCase.want {
			t.Errorf("%s\nGetBool(%+v, `%s`) = %+v, want %+v", testCase.intention, testCase.request, testCase.name, result, testCase.want)
		}
	}
}
