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
	cases := []struct {
		intention string
		want      string
	}{
		{
			"simple",
			"Usage of simple:\n  -csp string\n    \t[owasp] Content-Security-Policy {SIMPLE_CSP} (default \"default-src 'self'; base-uri 'self'\")\n  -frameOptions string\n    \t[owasp] X-Frame-Options {SIMPLE_FRAME_OPTIONS} (default \"deny\")\n  -hsts\n    \t[owasp] Indicate Strict Transport Security {SIMPLE_HSTS} (default true)\n",
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			fs := flag.NewFlagSet(tc.intention, flag.ContinueOnError)
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

func TestNew(t *testing.T) {
	cases := []struct {
		intention string
		want      App
	}{
		{
			"simple",
			App{
				csp:          "default-src 'self'; base-uri 'self'",
				hsts:         true,
				frameOptions: "deny",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			fs := flag.NewFlagSet(tc.intention, flag.ContinueOnError)

			if result := New(Flags(fs, "")); !reflect.DeepEqual(result, tc.want) {
				t.Errorf("New() = %#v, want %#v", result, tc.want)
			}
		})
	}
}

func TestMiddleware(t *testing.T) {
	cases := []struct {
		intention  string
		app        App
		next       http.Handler
		request    *http.Request
		wantStatus int
		wantHeader http.Header
	}{
		{
			"simple",
			App{
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
			"no value",
			App{
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
		{
			"hsts",
			App{
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
			"next",
			App{
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
		{
			"nonce",
			App{
				csp:          "default-src 'self'; base-uri 'self'; script-src 'self' 'nonce'",
				hsts:         false,
				frameOptions: "deny",
			},
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			}),
			httptest.NewRequest(http.MethodGet, "/", nil),
			http.StatusNotFound,
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

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			writer := httptest.NewRecorder()
			tc.app.Middleware(tc.next).ServeHTTP(writer, tc.request)

			if got := writer.Code; got != tc.wantStatus {
				t.Errorf("Middleware = %d, want %d", got, tc.wantStatus)
			}

			for key := range tc.wantHeader {
				want := tc.wantHeader.Get(key)
				if got := writer.Header().Get(key); !strings.HasPrefix(got, want) {
					t.Errorf("`%s` Header = `%s`, want `%s`", key, got, want)
				}
			}
		})
	}
}

func BenchmarkMiddleware(b *testing.B) {
	app := App{
		csp:          "default-src 'self'; base-uri 'self'",
		hsts:         true,
		frameOptions: "deny",
	}

	middleware := app.Middleware(nil)
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	writer := httptest.NewRecorder()

	for i := 0; i < b.N; i++ {
		middleware.ServeHTTP(writer, request)
	}
}

func BenchmarkMiddlewareNonce(b *testing.B) {
	app := App{
		csp:          "default-src 'self'; base-uri 'self'; script-src 'self' 'nonce'",
		hsts:         true,
		frameOptions: "deny",
	}

	middleware := app.Middleware(nil)
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	writer := httptest.NewRecorder()

	for i := 0; i < b.N; i++ {
		middleware.ServeHTTP(writer, request)
	}
}
