package query

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIsRoot(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		input *http.Request
		want  bool
	}{
		"empty": {
			httptest.NewRequest(http.MethodGet, "http://localhost", nil),
			true,
		},
		"trailing": {
			httptest.NewRequest(http.MethodGet, "http://localhost/", nil),
			true,
		},
		"complex": {
			httptest.NewRequest(http.MethodGet, "http://localhost/id", nil),
			false,
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			if result := IsRoot(testCase.input); result != testCase.want {
				t.Errorf("IsRoot() = %t, want %t", result, testCase.want)
			}
		})
	}
}
