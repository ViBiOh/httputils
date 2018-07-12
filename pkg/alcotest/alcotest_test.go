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
		for key := range r.Header {
			w.Header().Add(key, r.Header.Get(key))
		}

		if r.URL.Path == `/ok` {
			w.WriteHeader(http.StatusOK)
		} else if r.URL.Path == `/ko` {
			w.WriteHeader(http.StatusInternalServerError)
		} else if r.URL.Path == `/user-agent` {
			if r.Header.Get(`User-Agent`) != `Alcotest` {
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				w.WriteHeader(http.StatusServiceUnavailable)
			}
		} else {
			http.Error(w, `invalid`, http.StatusNotFound)
		}
	}))
}

func Test_Flags(t *testing.T) {
	var cases = []struct {
		intention string
		want      string
		wantType  string
	}{
		{
			`should add string url param to flags`,
			`url`,
			`*string`,
		},
		{
			`should add string userAgent param to flags`,
			`userAgent`,
			`*string`,
		},
	}

	for _, testCase := range cases {
		result := Flags(testCase.intention)[testCase.want]

		if result == nil {
			t.Errorf("%s\nFlags() = %+v, want `%s`", testCase.intention, result, testCase.want)
		}

		if fmt.Sprintf(`%T`, result) != testCase.wantType {
			t.Errorf("%s\nFlags() = `%T`, want `%s`", testCase.intention, result, testCase.wantType)
		}
	}
}

func Test_GetStatusCode(t *testing.T) {
	testServer := createTestServer()
	defer testServer.Close()

	var cases = []struct {
		intention   string
		url         string
		userAgent   string
		want        int
		wantErr     error
	}{
		{
			`should handle invalid request`,
			`:`,
			``,
			0,
			errors.New(`Error while creating request: parse :: missing protocol scheme`),
		},
		{
			`should handle malformed URL`,
			``,
			``,
			0,
			errors.New(`Get : unsupported protocol scheme ""`),
		},
		{
			`should return valid status from server`,
			fmt.Sprintf(`%s/ok`, testServer.URL),
			``,
			http.StatusOK,
			nil,
		},
		{
			`should return wrong status from server`,
			fmt.Sprintf(`%s/ko`, testServer.URL),
			``,
			http.StatusInternalServerError,
			nil,
		},
		{
			`should set given User-Agent`,
			fmt.Sprintf(`%s/user-agent`, testServer.URL),
			`Alcotest`,
			http.StatusServiceUnavailable,
			nil,
		},
	}

	var failed bool

	for _, testCase := range cases {
		result, err := GetStatusCode(testCase.url, testCase.userAgent)

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
			t.Errorf("%s\nGetStatusCode(`%s`, `%s`) = (%d, %+v), want (%d, %+v)", testCase.intention, testCase.url, testCase.userAgent, result, err, testCase.want, testCase.wantErr)
		}
	}
}

func Test_Do(t *testing.T) {
	testServer := createTestServer()
	defer testServer.Close()

	var cases = []struct {
		intention string
		url       string
		userAgent string
		want      error
	}{
		{
			`should handle error during call`,
			`http://`,
			`Test_Do`,
			errors.New(`Unable to blow in balloon: `),
		},
		{
			`should handle bad status code`,
			fmt.Sprintf(`%s/ko`, testServer.URL),
			`Test_Do`,
			errors.New(`Alcotest failed: HTTP/500`),
		},
		{
			`should handle valid code`,
			fmt.Sprintf(`%s/ok`, testServer.URL),
			`Test_Do`,
			nil,
		},
	}

	var failed bool

	for _, testCase := range cases {
		failed = false

		result := Do(testCase.url, testCase.userAgent)

		if result == nil && testCase.want != nil {
			failed = true
		} else if result != nil && testCase.want == nil {
			failed = true
		} else if result != nil && !strings.Contains(result.Error(), testCase.want.Error()) {
			failed = true
		}

		if failed {
			t.Errorf("%s\nDo(`%s`, `%s`) = %+v, want %+v", testCase.intention, testCase.url, testCase.userAgent, result, testCase.want)
		}
	}
}
