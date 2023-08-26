package owasp

import (
	"crypto/rand"
	"encoding/base64"
	"flag"
	"net/http"
	"strings"

	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/model"
)

const (
	nonceKey  = "httputils-nonce"
	cspHeader = "Content-Security-Policy"
)

var _ model.Middleware = Service{}.Middleware

type Service struct {
	csp          string
	frameOptions string
	hsts         bool
}

type Config struct {
	CSP          string
	FrameOptions string
	HSTS         bool
}

func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) Config {
	var config Config

	flags.New("Csp", cspHeader).Prefix(prefix).DocPrefix("owasp").StringVar(fs, &config.CSP, "default-src 'self'; base-uri 'self'", overrides)
	flags.New("Hsts", "Indicate Strict Transport Security").Prefix(prefix).DocPrefix("owasp").BoolVar(fs, &config.HSTS, true, overrides)
	flags.New("FrameOptions", "X-Frame-Options").Prefix(prefix).DocPrefix("owasp").StringVar(fs, &config.FrameOptions, "deny", overrides)

	return config
}

func New(config Config) Service {
	return Service{
		csp:          config.CSP,
		hsts:         config.HSTS,
		frameOptions: config.FrameOptions,
	}
}

func (s Service) Middleware(next http.Handler) http.Handler {
	headers := http.Header{}

	headers.Add("Referrer-Policy", "strict-origin-when-cross-origin")
	headers.Add("X-Content-Type-Options", "nosniff")
	headers.Add("X-Xss-Protection", "1; mode=block")
	headers.Add("X-Permitted-Cross-Domain-Policies", "none")

	if len(s.frameOptions) != 0 {
		headers.Add("X-Frame-Options", s.frameOptions)
	}
	if s.hsts {
		headers.Add("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
	}

	nonce := false
	if len(s.csp) != 0 {
		if strings.Contains(s.csp, nonceKey) {
			nonce = true
		} else {
			headers.Add(cspHeader, s.csp)
		}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for key, values := range headers {
			w.Header()[key] = values
		}

		if next != nil {
			writer := w
			if nonce {
				writer = newDelegator(w, s.csp)
			}

			next.ServeHTTP(writer, r)
		}
	})
}

func WriteNonce(w http.ResponseWriter, nonce string) {
	if nonceWriter, ok := w.(responseWriter); ok {
		nonceWriter.WriteNonce(nonce)
	}
}

func Nonce() string {
	raw := make([]byte, 16)

	if _, err := rand.Read(raw); err != nil {
		return "r4nd0m"
	}

	return base64.StdEncoding.EncodeToString(raw)
}
