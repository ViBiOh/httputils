package owasp

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/ViBiOh/httputils/pkg/tools"
)

const cacheControlHeader = `Cache-Control`

type middleware struct {
	http.ResponseWriter
	index bool
}

func (m *middleware) WriteHeader(status int) {
	if status == http.StatusOK || status == http.StatusMovedPermanently {
		if m.Header().Get(cacheControlHeader) == `` {
			if m.index {
				m.Header().Set(cacheControlHeader, `no-cache`)
			} else {
				m.Header().Set(cacheControlHeader, `max-age=864000`)
			}
		}
	}

	m.ResponseWriter.WriteHeader(status)
}

func (m *middleware) Push(target string, opts *http.PushOptions) error {
	if pusher, ok := m.ResponseWriter.(http.Pusher); ok {
		return pusher.Push(target, opts)
	}
	return http.ErrNotSupported
}

func (m *middleware) Flush() {
	if f, ok := m.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// App stores informations
type App struct {
	csp          string
	hsts         bool
	frameOptions string
}

// NewApp creates new App from Flags' config
func NewApp(config map[string]interface{}) *App {
	return &App{
		csp:          *(config[`csp`].(*string)),
		hsts:         *(config[`hsts`].(*bool)),
		frameOptions: *(config[`frameOptions`].(*string)),
	}
}

// Flags adds flags for given prefix
func Flags(prefix string) map[string]interface{} {
	return map[string]interface{}{
		`csp`:          flag.String(tools.ToCamel(fmt.Sprintf(`%sCsp`, prefix)), `default-src 'self'; base-uri 'self'`, `[owasp] Content-Security-Policy`),
		`hsts`:         flag.Bool(tools.ToCamel(fmt.Sprintf(`%sHsts`, prefix)), true, `[owasp] Indicate Strict Transport Security`),
		`frameOptions`: flag.String(tools.ToCamel(fmt.Sprintf(`%sFrameOptions`, prefix)), `deny`, `[owasp] X-Frame-Options`),
	}
}

// Handler for net/http package allowing owasp header
func (a App) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(`Content-Security-Policy`, a.csp)
		w.Header().Set(`Referrer-Policy`, `strict-origin-when-cross-origin`)
		w.Header().Set(`X-Frame-Options`, a.frameOptions)
		w.Header().Set(`X-Content-Type-Options`, `nosniff`)
		w.Header().Set(`X-Xss-Protection`, `1; mode=block`)
		w.Header().Set(`X-Permitted-Cross-Domain-Policies`, `none`)

		if a.hsts {
			w.Header().Set(`Strict-Transport-Security`, `max-age=10886400`)
		}

		next.ServeHTTP(&middleware{w, r.URL.Path == `/`}, r)
	})
}
