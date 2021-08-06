package owasp

import (
	"flag"
	"net/http"
	"strings"

	"github.com/ViBiOh/httputils/v4/pkg/flags"
	"github.com/ViBiOh/httputils/v4/pkg/model"
)

var (
	_ model.Middleware = app{}.Middleware
)

// App of package
type App interface {
	Middleware(http.Handler) http.Handler
}

// Config of package
type Config struct {
	csp          *string
	hsts         *bool
	frameOptions *string
}

type app struct {
	csp          string
	frameOptions string
	hsts         bool
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) Config {
	return Config{
		csp:          flags.New(prefix, "owasp").Name("Csp").Default(flags.Default("Csp", "default-src 'self'; base-uri 'self'", overrides)).Label("Content-Security-Policy").ToString(fs),
		hsts:         flags.New(prefix, "owasp").Name("Hsts").Default(flags.Default("Hsts", true, overrides)).Label("Indicate Strict Transport Security").ToBool(fs),
		frameOptions: flags.New(prefix, "owasp").Name("FrameOptions").Default(flags.Default("FrameOptions", "deny", overrides)).Label("X-Frame-Options").ToString(fs),
	}
}

// New creates new App from Config
func New(config Config) App {
	return app{
		csp:          strings.TrimSpace(*config.csp),
		hsts:         *config.hsts,
		frameOptions: strings.TrimSpace(*config.frameOptions),
	}
}

// Middleware for net/http package allowing owasp header
func (a app) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Add("X-Content-Type-Options", "nosniff")
		w.Header().Add("X-Xss-Protection", "1; mode=block")
		w.Header().Add("X-Permitted-Cross-Domain-Policies", "none")

		if len(a.csp) != 0 {
			w.Header().Add("Content-Security-Policy", a.csp)
		}
		if len(a.frameOptions) != 0 {
			w.Header().Add("X-Frame-Options", a.frameOptions)
		}
		if a.hsts {
			w.Header().Add("Strict-Transport-Security", "max-age=10886400")
		}

		if next != nil {
			next.ServeHTTP(w, r)
		}
	})
}
