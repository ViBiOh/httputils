package cors

import (
	"flag"
	"net/http"
	"strconv"
	"strings"

	"github.com/ViBiOh/httputils/v3/pkg/flags"
	"github.com/ViBiOh/httputils/v3/pkg/model"
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
	origin      *string
	headers     *string
	methods     *string
	exposes     *string
	credentials *bool
}

type app struct {
	origin      string
	headers     string
	methods     string
	exposes     string
	credentials string
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string) Config {
	return Config{
		origin:      flags.New(prefix, "cors").Name("Origin").Default("*").Label("Access-Control-Allow-Origin").ToString(fs),
		headers:     flags.New(prefix, "cors").Name("Headers").Default("Content-Type").Label("Access-Control-Allow-Headers").ToString(fs),
		methods:     flags.New(prefix, "cors").Name("Methods").Default(http.MethodGet).Label("Access-Control-Allow-Methods").ToString(fs),
		exposes:     flags.New(prefix, "cors").Name("Expose").Default("").Label("Access-Control-Expose-Headers").ToString(fs),
		credentials: flags.New(prefix, "cors").Name("Credentials").Default(false).Label("Access-Control-Allow-Credentials").ToBool(fs),
	}
}

// New creates new App from Config
func New(config Config) App {
	return &app{
		origin:      strings.TrimSpace(*config.origin),
		headers:     strings.TrimSpace(*config.headers),
		methods:     strings.TrimSpace(*config.methods),
		exposes:     strings.TrimSpace(*config.exposes),
		credentials: strconv.FormatBool(*config.credentials),
	}
}

// Middleware for net/http package allowing cors header
func (a app) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", a.origin)
		w.Header().Set("Access-Control-Allow-Headers", a.headers)
		w.Header().Set("Access-Control-Allow-Methods", a.methods)
		w.Header().Set("Access-Control-Expose-Headers", a.exposes)
		w.Header().Set("Access-Control-Allow-Credentials", a.credentials)

		if next != nil {
			next.ServeHTTP(w, r)
		}
	})
}
