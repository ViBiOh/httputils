package owasp

import (
	"flag"
	"fmt"
	"net/http"
	"strings"

	"github.com/ViBiOh/httputils/pkg/tools"
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
	docPrefix := prefix
	if prefix == "" {
		docPrefix = "owasp"
	}

	return Config{
		csp:          fs.String(tools.ToCamel(fmt.Sprintf("%sCsp", prefix)), "default-src 'self'; base-uri 'self'", fmt.Sprintf("[%s] Content-Security-Policy", docPrefix)),
		hsts:         fs.Bool(tools.ToCamel(fmt.Sprintf("%sHsts", prefix)), true, fmt.Sprintf("[%s] Indicate Strict Transport Security", docPrefix)),
		frameOptions: fs.String(tools.ToCamel(fmt.Sprintf("%sFrameOptions", prefix)), "deny", fmt.Sprintf("[%s] X-Frame-Options", docPrefix)),
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
