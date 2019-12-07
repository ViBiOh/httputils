package query

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestGetBool(t *testing.T) {
	var cases = []struct {
		intention string
		request   *http.Request
		name      string
		want      bool
	}{
		{
			"error",
			&http.Request{
				URL: &url.URL{
					RawQuery: "/%1",
				},
			},
			"",
			false,
		},
		{
			"should work with empty param",
			httptest.NewRequest(http.MethodGet, "http://localhost", nil),
			"valid",
			false,
		},
		{
			"should work with valid param",
			httptest.NewRequest(http.MethodGet, "http://localhost?valid", nil),
			"valid",
			true,
		},
		{
			"should work with valid value",
			httptest.NewRequest(http.MethodGet, "http://localhost?test=1&valid=false", nil),
			"valid",
			false,
		},
		{
			"should work with valid value not equal to a boolean",
			httptest.NewRequest(http.MethodGet, "http://localhost?test=1&valid=invalidBool", nil),
			"valid",
			false,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			if result := GetBool(testCase.request, testCase.name); result != testCase.want {
				t.Errorf("GetBool(%#v, `%s`) = %#v, want %#v", testCase.request, testCase.name, result, testCase.want)
			}
		})
	}
}
