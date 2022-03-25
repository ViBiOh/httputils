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
		} else if r.URL.Path == "/health" {
			w.WriteHeader(http.StatusOK)
		} else if r.URL.Path == "/ko" {
			w.WriteHeader(http.StatusInternalServerError)
		} else if r.URL.Path == "/reset" {
			w.WriteHeader(http.StatusResetContent)
		} else if r.URL.Path == "/user-agent" {
			if r.Header.Get("User-Agent") == defaultUserAgent {
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
	cases := map[string]struct {
		want string
	}{
		"simple": {
			"Usage of simple:\n  -url string\n    \t[alcotest] URL to check {SIMPLE_URL}\n  -userAgent string\n    \t[alcotest] User-Agent for check {SIMPLE_USER_AGENT} (default \"Alcotest\")\n",
		},
	}

	for intention, tc := range cases {
		t.Run(intention, func(t *testing.T) {
			fs := flag.NewFlagSet(intention, flag.ContinueOnError)
			Flags(fs, "")

			var writer strings.Builder
			fs.SetOutput(&writer)
			fs.Usage()

			result := writer.String()

			if result != tc.want {
				t.Errorf("Flags() = `%s`, want `%s`", result, tc.want)
			}
		})
	}
}

func TestGetStatusCode(t *testing.T) {
	testServer := createTestServer()
	defer testServer.Close()

	cases := map[string]struct {
		url       string
		userAgent string
		want      int
		wantErr   error
	}{
		"should handle invalid request": {
			":",
			"",
			0,
			errors.New(`parse ":": missing protocol scheme`),
		},
		"should handle malformed URL": {
			"",
			"",
			0,
			errors.New(`Get "": unsupported protocol scheme ""`),
		},
		"should return valid status from server": {
			testServer.URL + "/ok",
			"",
			http.StatusOK,
			nil,
		},
		"should return wrong status from server": {
			testServer.URL + "/ko",
			"",
			http.StatusInternalServerError,
			errors.New("HTTP/500"),
		},
		"should set given User-Agent": {
			fmt.Sprintf("%s/user-agent", testServer.URL),
			"Test",
			http.StatusServiceUnavailable,
			errors.New("HTTP/503"),
		},
	}

	for intention, tc := range cases {
		t.Run(intention, func(t *testing.T) {
			result, err := GetStatusCode(tc.url, tc.userAgent)

			failed := false

			if err == nil && tc.wantErr != nil {
				failed = true
			} else if err != nil && tc.wantErr == nil {
				failed = true
			} else if err != nil && !strings.Contains(err.Error(), tc.wantErr.Error()) {
				failed = true
			} else if result != tc.want {
				failed = true
			}

			if failed {
				t.Errorf("GetStatusCode() = (%d, `%s`), want (%d, `%s`)", result, err, tc.want, tc.wantErr)
			}
		})
	}
}

func TestDo(t *testing.T) {
	testServer := createTestServer()
	defer testServer.Close()

	cases := map[string]struct {
		url       string
		userAgent string
		want      error
	}{
		"should handle error during call": {
			"http://",
			"TestDo",
			errors.New("http: no Host in request URL"),
		},
		"should handle bad status code": {
			fmt.Sprintf("%s/ko", testServer.URL),
			"TestDo",
			errors.New("HTTP/500"),
		},
		"should handle valid status code": {
			fmt.Sprintf("%s/reset", testServer.URL),
			"TestDo",
			errors.New("alcotest failed: HTTP/205"),
		},
		"should handle valid code": {
			fmt.Sprintf("%s/ok", testServer.URL),
			"TestDo",
			nil,
		},
	}

	for intention, tc := range cases {
		t.Run(intention, func(t *testing.T) {
			failed := false

			result := Do(tc.url, tc.userAgent)

			if result == nil && tc.want != nil {
				failed = true
			} else if result != nil && tc.want == nil {
				failed = true
			} else if result != nil && !strings.Contains(result.Error(), tc.want.Error()) {
				failed = true
			}

			if failed {
				t.Errorf("Do() = `%s`, want `%s`", result, tc.want)
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

	cases := map[string]struct {
		input Config
		want  int
	}{
		"nothing to do": {
			Config{
				url:       &emptyString,
				userAgent: &userAgent,
			},
			-1,
		},
		"valid": {
			Config{
				url:       &healthy,
				userAgent: &userAgent,
			},
			0,
		},
		"invalid": {
			Config{
				url:       &unhealthy,
				userAgent: &userAgent,
			},
			1,
		},
	}

	for intention, tc := range cases {
		t.Run(intention, func(t *testing.T) {
			result := -1
			exitFunc = func(code int) {
				result = code
			}

			DoAndExit(tc.input)

			if result != tc.want {
				t.Errorf("DoAndExit() = %d, want %d", result, tc.want)
			}
		})
	}
}

func BenchmarkDoAndExit(b *testing.B) {
	testServer := createTestServer()
	defer testServer.Close()

	defaultURL = testServer.URL + "/health"

	config := Config{
		url:       &defaultURL,
		userAgent: &defaultUserAgent,
	}

	req, _ = http.NewRequest(http.MethodGet, defaultURL, nil)
	defaultHeader.Set("User-Agent", defaultUserAgent)
	exitFunc = func(_ int) {}

	for i := 0; i < b.N; i++ {
		DoAndExit(config)
	}
}
