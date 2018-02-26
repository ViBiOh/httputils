package owasp

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func Test_Flags(t *testing.T) {
	var cases = []struct {
		intention string
		prefix    string
		want      map[string]interface{}
	}{
		{
			`default prefix`,
			``,
			map[string]interface{}{
				`csp`:          nil,
				`hsts`:         nil,
				`frameOptions`: nil,
			},
		},
		{
			`given prefix`,
			`test`,
			map[string]interface{}{
				`csp`:          nil,
				`hsts`:         nil,
				`frameOptions`: nil,
			},
		},
	}

	for _, testCase := range cases {
		if result := Flags(testCase.prefix); len(result) != len(testCase.want) {
			t.Errorf("%v\nFlags(%v) = %v, want %v", testCase.intention, testCase.prefix, result, testCase.want)
		}
	}
}

func Test_ServeHTTP(t *testing.T) {
	hsts := false
	csp := `default-src 'self'; script-src 'self' 'unsafe-inline'`
	frameOptions := `allow-from https://vibioh.fr`

	var cases = []struct {
		path        string
		config      map[string]interface{}
		handlerFunc func(w http.ResponseWriter, r *http.Request)
		want        map[string]string
	}{
		{
			`/`,
			nil,
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
			nil,
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
			nil,
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
		Handler(testCase.config, http.HandlerFunc(testCase.handlerFunc)).ServeHTTP(request, httptest.NewRequest(http.MethodGet, `http://localhost`+testCase.path, nil))

		for key, value := range testCase.want {
			if result, ok := request.Result().Header[key]; !ok || (ok && strings.Join(result, ``) != value) {
				t.Errorf(`ServeHTTP() = [%v] = %v, want %v`, key, strings.Join(result, ``), value)
			}
		}
	}
}

func BenchmarkServeHTTP(b *testing.B) {
	handler := Handler(nil, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusMovedPermanently)
	}))

	request := httptest.NewRequest(http.MethodGet, `/`, nil)
	writer := httptest.NewRecorder()

	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(writer, request)
	}
}
