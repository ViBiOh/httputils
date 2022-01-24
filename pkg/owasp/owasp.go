package owasp

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"flag"
	"net/http"
	"strings"

	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/model"
)

type key int

const (
	ctxNonceKey key = iota

	nonceKey = "httputils-nonce"

	cspHeader = "Content-Security-Policy"
)

var _ model.Middleware = App{}.Middleware

// App of package
type App struct {
	csp          string
	frameOptions string
	hsts         bool
}

// Config of package
type Config struct {
	csp          *string
	hsts         *bool
	frameOptions *string
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) Config {
	return Config{
		csp:          flags.New(prefix, "owasp", "Csp").Default("default-src 'self'; base-uri 'self'", overrides).Label(cspHeader).ToString(fs),
		hsts:         flags.New(prefix, "owasp", "Hsts").Default(true, overrides).Label("Indicate Strict Transport Security").ToBool(fs),
		frameOptions: flags.New(prefix, "owasp", "FrameOptions").Default("deny", overrides).Label("X-Frame-Options").ToString(fs),
	}
}

// New creates new App from Config
func New(config Config) App {
	return App{
		csp:          *config.csp,
		hsts:         *config.hsts,
		frameOptions: *config.frameOptions,
	}
}

// Middleware for net/http package allowing owasp header
func (a App) Middleware(next http.Handler) http.Handler {
	headers := http.Header{}

	headers.Add("Referrer-Policy", "strict-origin-when-cross-origin")
	headers.Add("X-Content-Type-Options", "nosniff")
	headers.Add("X-Xss-Protection", "1; mode=block")
	headers.Add("X-Permitted-Cross-Domain-Policies", "none")

	nonce := false
	if len(a.csp) != 0 {
		if strings.Contains(a.csp, nonceKey) {
			nonce = true
		} else {
			headers.Add(cspHeader, a.csp)
		}
	}
	if len(a.frameOptions) != 0 {
		headers.Add("X-Frame-Options", a.frameOptions)
	}
	if a.hsts {
		headers.Add("Strict-Transport-Security", "max-age=10886400")
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for key, values := range headers {
			w.Header()[key] = values
		}

		if nonce {
			nonceValue := generateNonce()
			w.Header().Add(cspHeader, strings.ReplaceAll(a.csp, nonceKey, "nonce-"+nonceValue))
			r = r.WithContext(nonceInCtx(r.Context(), nonceValue))
		}

		if next != nil {
			next.ServeHTTP(w, r)
		}
	})
}

func nonceInCtx(ctx context.Context, nonce string) context.Context {
	return context.WithValue(ctx, ctxNonceKey, nonce)
}

// NonceFromCtx retrieves nonce from context
func NonceFromCtx(ctx context.Context) string {
	rawUser := ctx.Value(ctxNonceKey)
	if rawUser == nil {
		return ""
	}

	if nonce, ok := rawUser.(string); ok {
		return nonce
	}

	return ""
}

func generateNonce() string {
	raw := make([]byte, 16)
	_, err := rand.Read(raw)
	if err != nil {
		return "r4nd0m"
	}

	return base64.StdEncoding.EncodeToString(raw)
}
