package query

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIsRoot(t *testing.T) {
	var cases = []struct {
		intention string
		input     *http.Request
		want      bool
	}{
		{
			"empty",
			httptest.NewRequest(http.MethodGet, "http://localhost", nil),
			true,
		},
		{
			"trailing",
			httptest.NewRequest(http.MethodGet, "http://localhost/", nil),
			true,
		},
		{
			"complex",
			httptest.NewRequest(http.MethodGet, "http://localhost/id", nil),
			false,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			if result := IsRoot(testCase.input); result != testCase.want {
				t.Errorf("IsRoot() = %#v, want %#v", result, testCase.want)
			}
		})
	}
}

func TestGetID(t *testing.T) {
	var cases = []struct {
		intention string
		input     *http.Request
		want      string
	}{
		{
			"empty",
			httptest.NewRequest(http.MethodGet, "/", nil),
			"",
		},
		{
			"path",
			httptest.NewRequest(http.MethodGet, "/abc-1234", nil),
			"abc-1234",
		},
		{
			"complex",
			httptest.NewRequest(http.MethodGet, "/def-5678/links/", nil),
			"def-5678",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			if result := GetID(testCase.input); result != testCase.want {
				t.Errorf("getID() = %#v, want %#v", result, testCase.want)
			}
		})
	}
}
