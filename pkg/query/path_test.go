package query

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIsRoot(t *testing.T) {
	cases := []struct {
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
