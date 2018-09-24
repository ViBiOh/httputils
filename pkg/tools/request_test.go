package tools

import (
	"net/http"
	"testing"
)

func Test_IsRoot(t *testing.T) {
	empty, _ := http.NewRequest(http.MethodGet, ``, nil)
	slash, _ := http.NewRequest(http.MethodGet, `/`, nil)
	content, _ := http.NewRequest(http.MethodGet, `/id`, nil)

	var cases = []struct {
		intention string
		input     *http.Request
		want      bool
	}{
		{
			`should work with given params`,
			empty,
			true,
		},
		{
			`should work with given params`,
			slash,
			true,
		},
		{
			`should work with given params`,
			content,
			false,
		},
	}

	for _, testCase := range cases {
		if result := IsRoot(testCase.input); result != testCase.want {
			t.Errorf("%s\nIsRoot(`%+v`) = %+v, want %+v", testCase.intention, testCase.input, result, testCase.want)
		}
	}
}

func Test_GetID(t *testing.T) {
	emptyRequest, _ := http.NewRequest(http.MethodGet, `/`, nil)
	simpleRequest, _ := http.NewRequest(http.MethodGet, `/abc-1234`, nil)
	complexRequest, _ := http.NewRequest(http.MethodGet, `/def-5678/links/`, nil)

	var cases = []struct {
		intention string
		request   *http.Request
		want      string
	}{
		{
			`should work with empty URL`,
			emptyRequest,
			``,
		},
		{
			`should work with simple ID URL`,
			simpleRequest,
			`abc-1234`,
		},
		{
			`should work with complex ID URL`,
			complexRequest,
			`def-5678`,
		},
	}

	for _, testCase := range cases {
		if result := GetID(testCase.request); result != testCase.want {
			t.Errorf("%s\ngetID(`%+v`) = %+v, want %+v", testCase.intention, testCase.request, result, testCase.want)
		}
	}
}
