package alcotest

import (
	"errors"
	"flag"
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
		} else if r.URL.Path == "/reset" {
			w.WriteHeader(http.StatusResetContent)
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

func TestFlags(t *testing.T) {
	var cases = []struct {
		intention string
		want      string
	}{
		{
			"simple",
			"Usage of simple:\n  -url string\n    \t[alcotest] URL to check {SIMPLE_URL}\n  -userAgent string\n    \t[alcotest] User-Agent for check {SIMPLE_USER_AGENT} (default \"Alcotest\")\n",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			fs := flag.NewFlagSet(testCase.intention, flag.ContinueOnError)
			Flags(fs, "")

			var writer strings.Builder
			fs.SetOutput(&writer)
			fs.Usage()

			result := writer.String()

			if result != testCase.want {
				t.Errorf("Flags() = %s, want %s", result, testCase.want)
			}
		})
	}
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
			errors.New("HTTP/500"),
		},
		{
			"should set given User-Agent",
			fmt.Sprintf("%s/user-agent", testServer.URL),
			"Alcotest",
			http.StatusServiceUnavailable,
			errors.New("HTTP/503"),
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			result, err := GetStatusCode(testCase.url, testCase.userAgent)

			failed := false

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
				t.Errorf("GetStatusCode() = (%d, %s), want (%d, %s)", result, err, testCase.want, testCase.wantErr)
			}
		})
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
			errors.New("HTTP/500"),
		},
		{
			"should handle bad status code",
			fmt.Sprintf("%s/reset", testServer.URL),
			"TestDo",
			errors.New("alcotest failed: HTTP/205"),
		},
		{
			"should handle valid code",
			fmt.Sprintf("%s/ok", testServer.URL),
			"TestDo",
			nil,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			failed := false

			result := Do(testCase.url, testCase.userAgent)

			if result == nil && testCase.want != nil {
				failed = true
			} else if result != nil && testCase.want == nil {
				failed = true
			} else if result != nil && !strings.Contains(result.Error(), testCase.want.Error()) {
				failed = true
			}

			if failed {
				t.Errorf("Do() = %s, want %s", result, testCase.want)
			}
		})
	}
}

func TestDoAndExit(t *testing.T) {
	testServer := createTestServer()
	defer testServer.Close()

	emptyString := ""
	healthy := testServer.URL + "/ok"
	unhealthy := testServer.URL + "/ko"
	userAgent := "TestDoAndExit"

	var cases = []struct {
		intention string
		input     Config
		want      int
	}{
		{
			"nothing to do",
			Config{
				url:       &emptyString,
				userAgent: &userAgent,
			},
			-1,
		},
		{
			"valid",
			Config{
				url:       &healthy,
				userAgent: &userAgent,
			},
			0,
		},
		{
			"invalid",
			Config{
				url:       &unhealthy,
				userAgent: &userAgent,
			},
			1,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			result := -1
			exitFunc = func(code int) {
				result = code
			}

			DoAndExit(testCase.input)

			if result != testCase.want {
				t.Errorf("DoAndExit() = %d, want %d", result, testCase.want)
			}
		})
	}
}
