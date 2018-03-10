package alcotest

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func Test_Flags(t *testing.T) {
	var cases = []struct {
		intention string
		want      int
	}{
		{
			`should add one param to flags`,
			1,
		},
	}

	for _, testCase := range cases {
		if result := Flags(``); len(result) != testCase.want {
			t.Errorf("%s\nFlags() = %+v, want %+v", testCase.intention, result, testCase.want)
		}
	}
}

func Test_Do(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == `/ok` {
			w.WriteHeader(http.StatusOK)
		} else if r.URL.Path == `/ko` {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			http.Error(w, `invalid`, http.StatusNotFound)
		}
	}))
	defer testServer.Close()

	var cases = []struct {
		intention string
		url       string
		want      error
	}{
		{
			`should handle error while calling`,
			`http://`,
			errors.New(`Unable to blow in ballon`),
		},
		{
			`should handle bad status code`,
			`/ko`,
			errors.New(`Alcotest failed: HTTP/500`),
		},
		{
			`should handle valid code`,
			`/ok`,
			nil,
		},
	}

	var failed bool

	for _, testCase := range cases {
		failed = false

		result := Do(fmt.Sprintf(`%s%s`, testServer.URL, testCase.url))

		if result == nil && testCase.want != nil {
			failed = true
		} else if result != nil && testCase.want == nil {
			failed = true
		} else if result != nil && !strings.Contains(result.Error(), testCase.want.Error()) {
			failed = true
		}

		if failed {
			t.Errorf("%s\nDo(%+v) = %+v, want %+v", testCase.intention, testCase.url, result, testCase.want)
		}
	}
}
