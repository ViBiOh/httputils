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
