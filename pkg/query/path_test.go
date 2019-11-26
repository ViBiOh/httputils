package query

import (
	"errors"
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

func TestGetUintID(t *testing.T) {
	var cases = []struct {
		intention string
		input     string
		want      uint64
		wantErr   error
	}{
		{
			"empty",
			"/",
			0,
			ErrInvalidInteger,
		},
		{
			"valid",
			"/8000",
			8000,
			nil,
		},
		{
			"invalid uint",
			"/-8000",
			0,
			ErrInvalidInteger,
		},
		{
			"valid trailing slash",
			"/8000/do",
			8000,
			nil,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			result, err := GetUintID(httptest.NewRequest(http.MethodGet, testCase.input, nil))

			failed := false

			if testCase.wantErr != nil && !errors.Is(err, testCase.wantErr) {
				failed = true
			} else if result != testCase.want {
				failed = true
			}

			if failed {
				t.Errorf("GetUintID() = (%d, %#v), want (%d, %#v)", result, err, testCase.want, testCase.wantErr)
			}
		})
	}
}
