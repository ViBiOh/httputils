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

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			if result := IsRoot(tc.input); result != tc.want {
				t.Errorf("IsRoot() = %t, want %t", result, tc.want)
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

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			if result := GetID(tc.input); result != tc.want {
				t.Errorf("getID() = `%s`, want `%s`", result, tc.want)
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

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			result, err := GetUintID(httptest.NewRequest(http.MethodGet, tc.input, nil))

			failed := false

			if tc.wantErr != nil && !errors.Is(err, tc.wantErr) {
				failed = true
			} else if result != tc.want {
				failed = true
			}

			if failed {
				t.Errorf("GetUintID() = (%d, %v), want (%d, %v)", result, err, tc.want, tc.wantErr)
			}
		})
	}
}
