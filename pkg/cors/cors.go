package cors

import (
	"flag"
	"net/http"
	"strconv"

	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/model"
)

var _ model.Middleware = App{}.Middleware

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
		origin:      flags.String(fs, prefix, "cors", "Origin", "Access-Control-Allow-Origin", "*", overrides),
		headers:     flags.String(fs, prefix, "cors", "Headers", "Access-Control-Allow-Headers", "Content-Type", overrides),
		methods:     flags.String(fs, prefix, "cors", "Methods", "Access-Control-Allow-Methods", http.MethodGet, overrides),
		exposes:     flags.String(fs, prefix, "cors", "Expose", "Access-Control-Expose-Headers", "", overrides),
		credentials: flags.Bool(fs, prefix, "cors", "Credentials", "Access-Control-Allow-Credentials", false, overrides),
	}
}

// New creates new App from Config
func New(config Config) App {
	return App{
		origin:      *config.origin,
		headers:     *config.headers,
		methods:     *config.methods,
		exposes:     *config.exposes,
		credentials: strconv.FormatBool(*config.credentials),
	}
}

// Middleware for net/http package allowing cors header
func (a App) Middleware(next http.Handler) http.Handler {
	headers := http.Header{}

	if len(a.origin) != 0 {
		headers.Add("Access-Control-Allow-Origin", a.origin)
	}
	if len(a.headers) != 0 {
		headers.Add("Access-Control-Allow-Headers", a.headers)
	}
	if len(a.methods) != 0 {
		headers.Add("Access-Control-Allow-Methods", a.methods)
	}
	if len(a.exposes) != 0 {
		headers.Add("Access-Control-Expose-Headers", a.exposes)
	}
	if len(a.credentials) != 0 {
		headers.Add("Access-Control-Allow-Credentials", a.credentials)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for key, values := range headers {
			w.Header()[key] = values
		}

		if next != nil {
			next.ServeHTTP(w, r)
		}
	})
}
