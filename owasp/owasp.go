package owasp

import (
	"flag"
	"net/http"
)

type middleware struct {
	http.ResponseWriter
	path string
}

var (
	csp  = flag.String(`csp`, `default-src 'self'`, `Content-Security-Policy`)
	hsts = flag.Bool(`hsts`, true, `Indicate Strict Transport Security`)
)

func (m *middleware) WriteHeader(status int) {
	if status < http.StatusBadRequest {
		m.Header().Add(`Content-Security-Policy`, *csp)
		m.Header().Add(`Referrer-Policy`, `strict-origin-when-cross-origin`)
		m.Header().Add(`X-Frame-Options`, `deny`)
		m.Header().Add(`X-Content-Type-Options`, `nosniff`)
		m.Header().Add(`X-XSS-Protection`, `1; mode=block`)
		m.Header().Add(`X-Permitted-Cross-Domain-Policies`, `none`)
	}

	if *hsts {
		m.Header().Add(`Strict-Transport-Security`, `max-age=5184000`)
	}

	if status == http.StatusOK || status == http.StatusMovedPermanently {
		if m.path == `/` {
			m.Header().Add(`Cache-Control`, `no-cache`)
		} else {
			m.Header().Add(`Cache-Control`, `max-age=864000`)
		}
	}

	m.ResponseWriter.WriteHeader(status)
}

// Handler for net/http package allowing owasp header
type Handler struct {
	H http.Handler
}

func (handler Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler.H.ServeHTTP(&middleware{w, r.URL.Path}, r)
}
