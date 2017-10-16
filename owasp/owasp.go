package owasp

import (
	"flag"
	"net/http"
)

var (
	csp  = flag.String(`csp`, `default-src 'self'`, `Content-Security-Policy`)
	hsts = flag.Bool(`hsts`, true, `Indicate Strict Transport Security`)
)

type middleware struct {
	http.ResponseWriter
	index bool
}

func (m *middleware) WriteHeader(status int) {
	if status == http.StatusOK || status == http.StatusMovedPermanently {
		if m.index {
			m.Header().Add(`Cache-Control`, `no-cache`)
		} else {
			m.Header().Add(`Cache-Control`, `max-age=864000`)
		}
	}

	m.ResponseWriter.WriteHeader(status)
}

// Handler for net/http package allowing owasp header
func Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add(`Content-Security-Policy`, *csp)
		w.Header().Add(`Referrer-Policy`, `strict-origin-when-cross-origin`)
		w.Header().Add(`X-Frame-Options`, `deny`)
		w.Header().Add(`X-Content-Type-Options`, `nosniff`)
		w.Header().Add(`X-Xss-Protection`, `1; mode=block`)
		w.Header().Add(`X-Permitted-Cross-Domain-Policies`, `none`)

		if *hsts {
			w.Header().Add(`Strict-Transport-Security`, `max-age=5184000`)
		}

		next.ServeHTTP(&middleware{w, r.URL.Path == `/`}, r)
	})
}
