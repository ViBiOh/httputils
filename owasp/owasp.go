package owasp

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/ViBiOh/httputils/tools"
)

const cacheControlHeader = `Cache-Control`
const defaultCsp = `default-src 'self'; base-uri 'self'`
const defaultHsts = true
const defaultFrameOptions = `deny`

// Flags add flags for given prefix
func Flags(prefix string) map[string]interface{} {
	return map[string]interface{}{
		`csp`:          flag.String(tools.ToCamel(fmt.Sprintf(`%s%s`, prefix, `Csp`)), defaultCsp, `[owasp] Content-Security-Policy`),
		`hsts`:         flag.Bool(tools.ToCamel(fmt.Sprintf(`%s%s`, prefix, `Hsts`)), defaultHsts, `[owasp] Indicate Strict Transport Security`),
		`frameOptions`: flag.String(tools.ToCamel(fmt.Sprintf(`%s%s`, prefix, `FrameOptions`)), defaultFrameOptions, `[owasp] X-Frame-Options`),
	}
}

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

// Handler for net/http package allowing owasp header
func Handler(config map[string]interface{}, next http.Handler) http.Handler {
	var (
		csp          = defaultCsp
		hsts         = defaultHsts
		frameOptions = defaultFrameOptions
	)

	var given interface{}
	var ok bool

	if given, ok = config[`csp`]; ok {
		csp = *(given.(*string))
	}
	if given, ok = config[`hsts`]; ok {
		hsts = *(given.(*bool))
	}
	if given, ok = config[`frameOptions`]; ok {
		frameOptions = *(given.(*string))
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(`Content-Security-Policy`, csp)
		w.Header().Set(`Referrer-Policy`, `strict-origin-when-cross-origin`)
		w.Header().Set(`X-Frame-Options`, frameOptions)
		w.Header().Set(`X-Content-Type-Options`, `nosniff`)
		w.Header().Set(`X-Xss-Protection`, `1; mode=block`)
		w.Header().Set(`X-Permitted-Cross-Domain-Policies`, `none`)

		if hsts {
			w.Header().Set(`Strict-Transport-Security`, `max-age=5184000`)
		}

		next.ServeHTTP(&middleware{w, r.URL.Path == `/`}, r)
	})
}
