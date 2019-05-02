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

		if r.URL.Path == "/ok" {
			w.WriteHeader(http.StatusOK)
		} else if r.URL.Path == "/ko" {
			w.WriteHeader(http.StatusInternalServerError)
		} else if r.URL.Path == "/user-agent" {
			if r.Header.Get("User-Agent") != "Alcotest" {
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				w.WriteHeader(http.StatusServiceUnavailable)
			}
		} else {
			http.Error(w, "invalid", http.StatusNotFound)
		}
	}))
}

func TestGetStatusCode(t *testing.T) {
	testServer := createTestServer()
	defer testServer.Close()

	var cases = []struct {
		intention string
		url       string
		userAgent string
		want      int
		wantErr   error
	}{
		{
			"should handle invalid request",
			":",
			"",
			0,
			errors.New("parse :: missing protocol scheme"),
		},
		{
			"should handle malformed URL",
			"",
			"",
			0,
			errors.New("Get : unsupported protocol scheme \"\""),
		},
		{
			"should return valid status from server",
			fmt.Sprintf("%s/ok", testServer.URL),
			"",
			http.StatusOK,
			nil,
		},
		{
			"should return wrong status from server",
			fmt.Sprintf("%s/ko", testServer.URL),
			"",
			http.StatusInternalServerError,
			nil,
		},
		{
			"should set given User-Agent",
			fmt.Sprintf("%s/user-agent", testServer.URL),
			"Alcotest",
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

func TestDo(t *testing.T) {
	testServer := createTestServer()
	defer testServer.Close()

	var cases = []struct {
		intention string
		url       string
		userAgent string
		want      error
	}{
		{
			"should handle error during call",
			"http://",
			"TestDo",
			errors.New("http: no Host in request URL"),
		},
		{
			"should handle bad status code",
			fmt.Sprintf("%s/ko", testServer.URL),
			"TestDo",
			errors.New("alcotest failed: HTTP/500"),
		},
		{
			"should handle valid code",
			fmt.Sprintf("%s/ok", testServer.URL),
			"TestDo",
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
