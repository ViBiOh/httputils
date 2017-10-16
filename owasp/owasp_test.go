package owasp

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestServeHTTP(t *testing.T) {
	var cases = []struct {
		path        string
		handlerFunc func(w http.ResponseWriter, r *http.Request)
		want        map[string]string
	}{
		{
			`/`,
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			map[string]string{
				`Content-Security-Policy`:           `default-src 'self'`,
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
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusMovedPermanently)
			},
			map[string]string{
				`Content-Security-Policy`:           `default-src 'self'`,
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
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
			},
			map[string]string{
				`Content-Security-Policy`:           `default-src 'self'`,
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
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			map[string]string{
				`Content-Security-Policy`:           `default-src 'self'`,
				`Referrer-Policy`:                   `strict-origin-when-cross-origin`,
				`X-Frame-Options`:                   `deny`,
				`X-Content-Type-Options`:            `nosniff`,
				`X-Xss-Protection`:                  `1; mode=block`,
				`X-Permitted-Cross-Domain-Policies`: `none`,
				`Strict-Transport-Security`:         `max-age=5184000`,
				`Cache-Control`:                     `max-age=864000`,
			},
		},
	}

	for _, testCase := range cases {
		request := httptest.NewRecorder()
		Handler(http.HandlerFunc(testCase.handlerFunc)).ServeHTTP(request, httptest.NewRequest(http.MethodGet, `http://localhost`+testCase.path, nil))

		for key, value := range testCase.want {
			if result, ok := request.Result().Header[key]; !ok || (ok && strings.Join(result, ``) != value) {
				t.Errorf(`ServeHTTP() = [%v] = %v, want %v`, key, strings.Join(result, ``), value)
			}
		}
	}
}

func BenchmarkServeHTTP(b *testing.B) {
	handler := Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusMovedPermanently)
	}))

	request := httptest.NewRequest(http.MethodGet, `/`, nil)
	writer := httptest.NewRecorder()

	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(writer, request)
	}
}
