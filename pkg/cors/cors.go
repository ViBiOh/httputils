package cors

import (
	"flag"
	"net/http"
	"strconv"
	"strings"

	"github.com/ViBiOh/httputils/v4/pkg/flags"
	"github.com/ViBiOh/httputils/v4/pkg/model"
)

var (
	_ model.Middleware = App{}.Middleware
)

// App of package
type App struct {
	origin      string
	headers     string
	methods     string
	exposes     string
	credentials string
}

// Config of package
type Config struct {
	origin      *string
	headers     *string
	methods     *string
	exposes     *string
	credentials *bool
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) Config {
	return Config{
		origin:      flags.New(prefix, "cors").Name("Origin").Default(flags.Default("Origin", "*", overrides)).Label("Access-Control-Allow-Origin").ToString(fs),
		headers:     flags.New(prefix, "cors").Name("Headers").Default(flags.Default("Headers", "Content-Type", overrides)).Label("Access-Control-Allow-Headers").ToString(fs),
		methods:     flags.New(prefix, "cors").Name("Methods").Default(flags.Default("Methods", http.MethodGet, overrides)).Label("Access-Control-Allow-Methods").ToString(fs),
		exposes:     flags.New(prefix, "cors").Name("Expose").Default(flags.Default("Expose", "", overrides)).Label("Access-Control-Expose-Headers").ToString(fs),
		credentials: flags.New(prefix, "cors").Name("Credentials").Default(flags.Default("Credentials", false, overrides)).Label("Access-Control-Allow-Credentials").ToBool(fs),
	}
}

// New creates new App from Config
func New(config Config) App {
	return App{
		origin:      strings.TrimSpace(*config.origin),
		headers:     strings.TrimSpace(*config.headers),
		methods:     strings.TrimSpace(*config.methods),
		exposes:     strings.TrimSpace(*config.exposes),
		credentials: strconv.FormatBool(*config.credentials),
	}
}

// Middleware for net/http package allowing cors header
func (a App) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(a.origin) != 0 {
			w.Header().Add("Access-Control-Allow-Origin", a.origin)
		}
		if len(a.headers) != 0 {
			w.Header().Add("Access-Control-Allow-Headers", a.headers)
		}
		if len(a.methods) != 0 {
			w.Header().Add("Access-Control-Allow-Methods", a.methods)
		}
		if len(a.exposes) != 0 {
			w.Header().Add("Access-Control-Expose-Headers", a.exposes)
		}
		if len(a.credentials) != 0 {
			w.Header().Add("Access-Control-Allow-Credentials", a.credentials)
		}

		if next != nil {
			next.ServeHTTP(w, r)
		}
	})
}
