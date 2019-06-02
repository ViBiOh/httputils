package tools

import (
	"net/http"
	"testing"
)

func TestIsRoot(t *testing.T) {
	empty, _ := http.NewRequest(http.MethodGet, "", nil)
	slash, _ := http.NewRequest(http.MethodGet, "/", nil)
	content, _ := http.NewRequest(http.MethodGet, "/id", nil)

	var cases = []struct {
		intention string
		input     *http.Request
		want      bool
	}{
		{
			"should work with given params",
			empty,
			true,
		},
		{
			"should work with given params",
			slash,
			true,
		},
		{
			"should work with given params",
			content,
			false,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			if result := IsRoot(testCase.input); result != testCase.want {
				t.Errorf("IsRoot(`%#v`) = %#v, want %#v", testCase.input, result, testCase.want)
			}
		})
	}
}

func TestGetID(t *testing.T) {
	emptyRequest, _ := http.NewRequest(http.MethodGet, "/", nil)
	simpleRequest, _ := http.NewRequest(http.MethodGet, "/abc-1234", nil)
	complexRequest, _ := http.NewRequest(http.MethodGet, "/def-5678/links/", nil)

	var cases = []struct {
		intention string
		request   *http.Request
		want      string
	}{
		{
			"should work with empty URL",
			emptyRequest,
			"",
		},
		{
			"should work with simple ID URL",
			simpleRequest,
			"abc-1234",
		},
		{
			"should work with complex ID URL",
			complexRequest,
			"def-5678",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			if result := GetID(testCase.request); result != testCase.want {
				t.Errorf("getID(`%#v`) = %#v, want %#v", testCase.request, result, testCase.want)
			}
		})
	}
}
