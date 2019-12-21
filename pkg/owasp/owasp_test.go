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
	var cases = []struct {
		intention string
		want      string
	}{
		{
			"simple",
			"Usage of simple:\n  -csp string\n    \t[owasp] Content-Security-Policy {SIMPLE_CSP} (default \"default-src 'self'; base-uri 'self'\")\n  -frameOptions string\n    \t[owasp] X-Frame-Options {SIMPLE_FRAME_OPTIONS} (default \"deny\")\n  -hsts\n    \t[owasp] Indicate Strict Transport Security {SIMPLE_HSTS} (default true)\n",
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
				t.Errorf("Flags() =`%s`, want`%s`", result, testCase.want)
			}
		})
	}
}

func TestNew(t *testing.T) {
	var cases = []struct {
		intention string
		want      App
	}{
		{
			"simple",
			&app{
				csp:          "default-src 'self'; base-uri 'self'",
				hsts:         true,
				frameOptions: "deny",
			},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			fs := flag.NewFlagSet(testCase.intention, flag.ContinueOnError)

			if result := New(Flags(fs, "")); !reflect.DeepEqual(result, testCase.want) {
				t.Errorf("New() = %#v, want %#v", result, testCase.want)
			}
		})
	}
}

func TestMiddleware(t *testing.T) {
	var cases = []struct {
		intention  string
		app        App
		next       http.Handler
		request    *http.Request
		want       int
		wantHeader http.Header
	}{
		{
			"simple",
			app{
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
		{
			"hsts",
			app{
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
				"Strict-Transport-Security":         []string{"max-age=10886400"},
			},
		},
		{
			"add cache control for index if not set",
			app{
				csp:          "default-src 'self'; base-uri 'self'",
				hsts:         false,
				frameOptions: "deny",
			},
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}),
			httptest.NewRequest(http.MethodGet, "/", nil),
			http.StatusOK,
			http.Header{
				"Content-Security-Policy":           []string{"default-src 'self'; base-uri 'self'"},
				"Referrer-Policy":                   []string{"strict-origin-when-cross-origin"},
				"X-Frame-Options":                   []string{"deny"},
				"X-Content-Type-Options":            []string{"nosniff"},
				"X-Xss-Protection":                  []string{"1; mode=block"},
				"X-Permitted-Cross-Domain-Policies": []string{"none"},
				"Cache-Control":                     []string{"no-cache"},
			},
		},
		{
			"add cache control for no index if not set",
			app{
				csp:          "default-src 'self'; base-uri 'self'",
				hsts:         false,
				frameOptions: "deny",
			},
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}),
			httptest.NewRequest(http.MethodGet, "/404.html", nil),
			http.StatusOK,
			http.Header{
				"Content-Security-Policy":           []string{"default-src 'self'; base-uri 'self'"},
				"Referrer-Policy":                   []string{"strict-origin-when-cross-origin"},
				"X-Frame-Options":                   []string{"deny"},
				"X-Content-Type-Options":            []string{"nosniff"},
				"X-Xss-Protection":                  []string{"1; mode=block"},
				"X-Permitted-Cross-Domain-Policies": []string{"none"},
				"Cache-Control":                     []string{"max-age=864000"},
			},
		},
		{
			"do no touch cache control if set",
			app{
				csp:          "default-src 'self'; base-uri 'self'",
				hsts:         false,
				frameOptions: "deny",
			},
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set(cacheControlHeader, "max-age=3000")
				w.WriteHeader(http.StatusOK)
			}),
			httptest.NewRequest(http.MethodGet, "/404.html", nil),
			http.StatusOK,
			http.Header{
				"Content-Security-Policy":           []string{"default-src 'self'; base-uri 'self'"},
				"Referrer-Policy":                   []string{"strict-origin-when-cross-origin"},
				"X-Frame-Options":                   []string{"deny"},
				"X-Content-Type-Options":            []string{"nosniff"},
				"X-Xss-Protection":                  []string{"1; mode=block"},
				"X-Permitted-Cross-Domain-Policies": []string{"none"},
				"Cache-Control":                     []string{"max-age=3000"},
			},
		},
		{
			"write header if not done",
			app{
				csp:          "default-src 'self'; base-uri 'self'",
				hsts:         false,
				frameOptions: "deny",
			},
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(""))
			}),
			httptest.NewRequest(http.MethodGet, "/", nil),
			http.StatusOK,
			http.Header{
				"Content-Security-Policy":           []string{"default-src 'self'; base-uri 'self'"},
				"Referrer-Policy":                   []string{"strict-origin-when-cross-origin"},
				"X-Frame-Options":                   []string{"deny"},
				"X-Content-Type-Options":            []string{"nosniff"},
				"X-Xss-Protection":                  []string{"1; mode=block"},
				"X-Permitted-Cross-Domain-Policies": []string{"none"},
				"Cache-Control":                     []string{"no-cache"},
				"Content-Type":                      []string{"text/plain; charset=utf-8"},
			},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			writer := httptest.NewRecorder()

			testCase.app.Middleware(testCase.next).ServeHTTP(writer, testCase.request)

			if writer.Code != testCase.want {
				t.Errorf("Middleware() = %d, want %d", writer.Code, testCase.want)
			}

			if !reflect.DeepEqual(writer.Header(), testCase.wantHeader) {
				t.Errorf("Middleware() = %#v, want %#v", writer.Header(), testCase.wantHeader)
			}
		})
	}
}
