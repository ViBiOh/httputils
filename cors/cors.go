package owasp

import (
	"flag"
	"net/http"
)

type middleware struct {
	http.ResponseWriter
}

var (
	origin  = flag.String(`corsOrigin`, `*`, `Access-Control-Allow-Origin`)
	headers = flag.String(`corsHeaders`, `Content-Type`, `Access-Control-Allow-Headers`)
	methods = flag.String(`corsMethods`, `GET`, `Access-Control-Allow-Methods`)
)

func (m *middleware) WriteHeader(status int) {
	m.Header().Add(`Access-Control-Allow-Origin`, *origin)
	m.Header().Add(`Access-Control-Allow-Headers`, *headers)
	m.Header().Add(`Access-Control-Allow-Methods`, *methods)

	m.ResponseWriter.WriteHeader(status)
}

// Handler for net/http package allowing cors header
type Handler struct {
	H http.Handler
}

func (handler Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler.H.ServeHTTP(&middleware{w}, r)
}
