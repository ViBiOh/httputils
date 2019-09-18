package owasp

import (
	"flag"
	"net/http"
	"strings"

	"github.com/ViBiOh/httputils/v2/pkg/tools"
)

// Config of package
type Config struct {
	csp          *string
	hsts         *bool
	frameOptions *string
}

// App of package
type App struct {
	csp          string
	hsts         bool
	frameOptions string
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string) Config {
	return Config{
		csp:          tools.NewFlag(prefix, "owasp").Name("Csp").Default("default-src 'self'; base-uri 'self'").Label("Content-Security-Policy").ToString(fs),
		hsts:         tools.NewFlag(prefix, "owasp").Name("Hsts").Default(true).Label("Indicate Strict Transport Security").ToBool(fs),
		frameOptions: tools.NewFlag(prefix, "owasp").Name("FrameOptions").Default("deny").Label("X-Frame-Options").ToString(fs),
	}
}

// New creates new App from Config
func New(config Config) *App {
	return &App{
		csp:          strings.TrimSpace(*config.csp),
		hsts:         *config.hsts,
		frameOptions: strings.TrimSpace(*config.frameOptions),
	}
}

// Handler for net/http package allowing owasp header
func (a App) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", a.csp)
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("X-Frame-Options", a.frameOptions)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Xss-Protection", "1; mode=block")
		w.Header().Set("X-Permitted-Cross-Domain-Policies", "none")

		if a.hsts {
			w.Header().Set("Strict-Transport-Security", "max-age=10886400")
		}

		if next != nil {
			next.ServeHTTP(&middleware{w, tools.IsRoot(r), false}, r)
		}
	})
}
