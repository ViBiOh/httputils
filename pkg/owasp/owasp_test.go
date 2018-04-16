package owasp

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var (
	hsts         = false
	csp          = `default-src 'self'; script-src 'self' 'unsafe-inline'`
	frameOptions = `allow-from https://vibioh.fr`
)

func Test_Flags(t *testing.T) {
	var cases = []struct {
		intention string
		want      int
	}{
		{
			`should add 3 params to flags`,
			3,
		},
	}

	for _, testCase := range cases {
		if result := Flags(`owasp_Test_Flags`); len(result) != testCase.want {
			t.Errorf("%s\nFlags() = %+v, want %+v", testCase.intention, result, testCase.want)
		}
	}
}

func Test_ServeHTTP(t *testing.T) {
	var cases = []struct {
		path        string
		config      map[string]interface{}
		handlerFunc func(w http.ResponseWriter, r *http.Request)
		want        map[string]string
	}{
		{
			`/`,
			Flags(`owasp_Test_ServeHTTP_default`),
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			map[string]string{
				`Content-Security-Policy`:           `default-src 'self'; base-uri 'self'`,
				`Referrer-Policy`:                   `strict-origin-when-cross-origin`,
				`X-Frame-Options`:                   `deny`,
				`X-Content-Type-Options`:            `nosniff`,
				`X-Xss-Protection`:                  `1; mode=block`,
				`X-Permitted-Cross-Domain-Policies`: `none`,
				`Strict-Transport-Security`:         `max-age=5184000`,
				`Cache-Control`:                     `no-cache`,
			},
		},
		{
			`/`,
			Flags(`owasp_Test_ServeHTTP_redirect`),
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusMovedPermanently)
			},
			map[string]string{
				`Content-Security-Policy`:           `default-src 'self'; base-uri 'self'`,
				`Referrer-Policy`:                   `strict-origin-when-cross-origin`,
				`X-Frame-Options`:                   `deny`,
				`X-Content-Type-Options`:            `nosniff`,
				`X-Xss-Protection`:                  `1; mode=block`,
				`X-Permitted-Cross-Domain-Policies`: `none`,
				`Strict-Transport-Security`:         `max-age=5184000`,
				`Cache-Control`:                     `no-cache`,
			},
		},
		{
			`/`,
			Flags(`owasp_Test_ServeHTTP_bad`),
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
			},
			map[string]string{
				`Content-Security-Policy`:           `default-src 'self'; base-uri 'self'`,
				`Referrer-Policy`:                   `strict-origin-when-cross-origin`,
				`X-Frame-Options`:                   `deny`,
				`X-Content-Type-Options`:            `nosniff`,
				`X-Xss-Protection`:                  `1; mode=block`,
				`X-Permitted-Cross-Domain-Policies`: `none`,
				`Strict-Transport-Security`:         `max-age=5184000`,
			},
		},
		{
			`/testCase.html`,
			map[string]interface{}{
				`csp`:          &csp,
				`hsts`:         &hsts,
				`frameOptions`: &frameOptions,
			},
			func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set(`Cache-Control`, `max-age=no-cache`)
				w.WriteHeader(http.StatusOK)
			},
			map[string]string{
				`Content-Security-Policy`:           `default-src 'self'; script-src 'self' 'unsafe-inline'`,
				`Referrer-Policy`:                   `strict-origin-when-cross-origin`,
				`X-Frame-Options`:                   `allow-from https://vibioh.fr`,
				`X-Content-Type-Options`:            `nosniff`,
				`X-Xss-Protection`:                  `1; mode=block`,
				`X-Permitted-Cross-Domain-Policies`: `none`,
				`Cache-Control`:                     `max-age=no-cache`,
			},
		},
	}

	for _, testCase := range cases {
		request := httptest.NewRecorder()
		Handler(testCase.config, http.HandlerFunc(testCase.handlerFunc)).ServeHTTP(request, httptest.NewRequest(http.MethodGet, fmt.Sprintf(`http://localhost%s`, testCase.path), nil))

		for key, value := range testCase.want {
			if result, ok := request.Result().Header[key]; !ok || (ok && strings.Join(result, ``) != value) {
				t.Errorf(`ServeHTTP() = [%v] = %v, want %v`, key, strings.Join(result, ``), value)
			}
		}
	}
}

func Benchmark_ServeHTTP(b *testing.B) {
	handler := Handler(
		map[string]interface{}{
			`csp`:          &csp,
			`hsts`:         &hsts,
			`frameOptions`: &frameOptions,
		}, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusMovedPermanently)
		}))

	request := httptest.NewRequest(http.MethodGet, `/`, nil)
	writer := httptest.NewRecorder()

	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(writer, request)
	}
}
