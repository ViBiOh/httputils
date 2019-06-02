package owasp

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestHandler(t *testing.T) {
	var cases = []struct {
		intention  string
		app        App
		next       http.Handler
		request    *http.Request
		want       int
		wantHeader http.Header
	}{
		{
			"default param",
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
			"add hsts",
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
			"add cache control for index if not set",
			App{
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
			App{
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
			App{
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
			App{
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

			testCase.app.Handler(testCase.next).ServeHTTP(writer, testCase.request)

			if writer.Code != testCase.want {
				t.Errorf("Handler(%#v) = %d, want %d", testCase.next, writer.Code, testCase.want)
			}

			if !reflect.DeepEqual(writer.Header(), testCase.wantHeader) {
				t.Errorf("Handler(%#v) = %#v, want %#v", testCase.next, writer.Header(), testCase.wantHeader)
			}
		})
	}
}
