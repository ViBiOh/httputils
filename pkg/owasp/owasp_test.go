package owasp

import (
	"flag"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestFlags(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		want string
	}{
		"simple": {
			"Usage of simple:\n  -csp string\n    \t[owasp] Content-Security-Policy ${SIMPLE_CSP} (default \"default-src 'self'; base-uri 'self'\")\n  -frameOptions string\n    \t[owasp] X-Frame-Options ${SIMPLE_FRAME_OPTIONS} (default \"deny\")\n  -hsts\n    \t[owasp] Indicate Strict Transport Security ${SIMPLE_HSTS} (default true)\n",
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			fs := flag.NewFlagSet(intention, flag.ContinueOnError)
			Flags(fs, "")

			var writer strings.Builder
			fs.SetOutput(&writer)
			fs.Usage()

			result := writer.String()

			if result != testCase.want {
				t.Errorf("Flags() = `%s`, want `%s`", result, testCase.want)
			}
		})
	}
}

func TestNew(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		want Service
	}{
		"simple": {
			Service{
				csp:          "default-src 'self'; base-uri 'self'",
				hsts:         true,
				frameOptions: "deny",
			},
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			fs := flag.NewFlagSet(intention, flag.ContinueOnError)

			if result := New(Flags(fs, "")); !reflect.DeepEqual(result, testCase.want) {
				t.Errorf("New() = %#v, want %#v", result, testCase.want)
			}
		})
	}
}

func TestMiddleware(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		service    Service
		next       http.Handler
		request    *http.Request
		wantStatus int
		wantHeader http.Header
	}{
		"simple": {
			Service{
				csp:          "default-src 'self'; base-uri 'self'",
				hsts:         false,
				frameOptions: "deny",
			},
			nil,
			httptest.NewRequest(http.MethodGet, "/", nil),
			http.StatusOK,
			http.Header{
				"Content-Security-Policy":           []string{"default-src 'self'; base-uri 'self'"},
				"Referrer-Policy":                   []string{"strict-origin-when-cross-origin"},
				"X-Frame-Options":                   []string{"deny"},
				"X-Content-Type-Options":            []string{"nosniff"},
				"X-Xss-Protection":                  []string{"1; mode=block"},
				"X-Permitted-Cross-Domain-Policies": []string{"none"},
			},
		},
		"no value": {
			Service{
				csp:          "",
				hsts:         false,
				frameOptions: "",
			},
			nil,
			httptest.NewRequest(http.MethodGet, "/", nil),
			http.StatusOK,
			http.Header{
				"Referrer-Policy":                   []string{"strict-origin-when-cross-origin"},
				"X-Content-Type-Options":            []string{"nosniff"},
				"X-Xss-Protection":                  []string{"1; mode=block"},
				"X-Permitted-Cross-Domain-Policies": []string{"none"},
			},
		},
		"hsts": {
			Service{
				csp:          "default-src 'self'; base-uri 'self'",
				hsts:         true,
				frameOptions: "deny",
			},
			nil,
			httptest.NewRequest(http.MethodGet, "/", nil),
			http.StatusOK,
			http.Header{
				"Content-Security-Policy":           []string{"default-src 'self'; base-uri 'self'"},
				"Referrer-Policy":                   []string{"strict-origin-when-cross-origin"},
				"X-Frame-Options":                   []string{"deny"},
				"X-Content-Type-Options":            []string{"nosniff"},
				"X-Xss-Protection":                  []string{"1; mode=block"},
				"X-Permitted-Cross-Domain-Policies": []string{"none"},
				"Strict-Transport-Security":         []string{"max-age=63072000; includeSubDomains; preload"},
			},
		},
		"next": {
			Service{
				csp:          "default-src 'self'; base-uri 'self'",
				hsts:         false,
				frameOptions: "deny",
			},
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			}),
			httptest.NewRequest(http.MethodGet, "/", nil),
			http.StatusNotFound,
			http.Header{
				"Content-Security-Policy":           []string{"default-src 'self'; base-uri 'self'"},
				"Referrer-Policy":                   []string{"strict-origin-when-cross-origin"},
				"X-Frame-Options":                   []string{"deny"},
				"X-Content-Type-Options":            []string{"nosniff"},
				"X-Xss-Protection":                  []string{"1; mode=block"},
				"X-Permitted-Cross-Domain-Policies": []string{"none"},
			},
		},
		"nonce": {
			Service{
				csp:          "default-src 'self'; base-uri 'self'; script-src 'self' 'httputils-nonce'",
				hsts:         false,
				frameOptions: "deny",
			},
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusTeapot)
			}),
			httptest.NewRequest(http.MethodGet, "/", nil),
			http.StatusTeapot,
			http.Header{
				"Content-Security-Policy":           []string{"default-src 'self'; base-uri 'self'; script-src 'self' 'nonce-"},
				"Referrer-Policy":                   []string{"strict-origin-when-cross-origin"},
				"X-Frame-Options":                   []string{"deny"},
				"X-Content-Type-Options":            []string{"nosniff"},
				"X-Xss-Protection":                  []string{"1; mode=block"},
				"X-Permitted-Cross-Domain-Policies": []string{"none"},
			},
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			writer := httptest.NewRecorder()
			testCase.service.Middleware(testCase.next).ServeHTTP(writer, testCase.request)

			if got := writer.Code; got != testCase.wantStatus {
				t.Errorf("Middleware = %d, want %d", got, testCase.wantStatus)
			}

			for key := range testCase.wantHeader {
				want := testCase.wantHeader.Get(key)
				if got := writer.Header().Get(key); !strings.HasPrefix(got, want) {
					t.Errorf("`%s` Header = `%s`, want `%s`", key, got, want)
				}
			}
		})
	}
}

func BenchmarkMiddleware(b *testing.B) {
	service := Service{
		csp:          "default-src 'self'; base-uri 'self'",
		hsts:         true,
		frameOptions: "deny",
	}

	middleware := service.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	writer := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		middleware.ServeHTTP(writer, request)
	}
}

func BenchmarkMiddlewareNonce(b *testing.B) {
	service := Service{
		csp:          "default-src 'self'; base-uri 'self'; script-src 'self' 'httputils-nonce'",
		hsts:         true,
		frameOptions: "deny",
	}

	middleware := service.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	writer := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		middleware.ServeHTTP(writer, request)
	}
}
