package alcotest

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func createTestServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == `/ok` {
			w.WriteHeader(http.StatusOK)
		} else if r.URL.Path == `/ko` {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			http.Error(w, `invalid`, http.StatusNotFound)
		}
	}))
}

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

func Test_GetStatusCode(t *testing.T) {
	testServer := createTestServer()
	defer testServer.Close()

	var cases = []struct {
		intention string
		url       string
		want      int
		wantErr   error
	}{
		{
			`should handle empty string`,
			``,
			0,
			errors.New(`Get : unsupported protocol scheme ""`),
		},
		{
			`should return status from server`,
			fmt.Sprintf(`%s/ok`, testServer.URL),
			http.StatusOK,
			nil,
		},
	}

	var failed bool

	for _, testCase := range cases {
		result, err := GetStatusCode(testCase.url)

		failed = false

		if err == nil && testCase.wantErr != nil {
			failed = true
		} else if err != nil && testCase.wantErr == nil {
			failed = true
		} else if err != nil && err.Error() != testCase.wantErr.Error() {
			failed = true
		} else if result != testCase.want {
			failed = true
		}

		if failed {
			t.Errorf("%s\nGetStatusCode(%+v) = (%+v, %+v), want (%+v, %+v)", testCase.intention, testCase.url, result, err, testCase.want, testCase.wantErr)
		}
	}
}

func Test_Do(t *testing.T) {
	testServer := createTestServer()
	defer testServer.Close()

	var cases = []struct {
		intention string
		url       string
		want      error
	}{
		{
			`should handle error during call`,
			`http://`,
			errors.New(`Unable to blow in ballon`),
		},
		{
			`should handle bad status code`,
			fmt.Sprintf(`%s/ko`, testServer.URL),
			errors.New(`Alcotest failed: HTTP/500`),
		},
		{
			`should handle valid code`,
			fmt.Sprintf(`%s/ok`, testServer.URL),
			nil,
		},
	}

	var failed bool

	for _, testCase := range cases {
		failed = false

		result := Do(testCase.url)

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
