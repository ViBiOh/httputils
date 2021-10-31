package cors

import (
	"flag"
	"net/http"
	"strconv"

	"github.com/ViBiOh/httputils/v4/pkg/flags"
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
		origin:      flags.New(prefix, "cors", "Origin").Default("*", overrides).Label("Access-Control-Allow-Origin").ToString(fs),
		headers:     flags.New(prefix, "cors", "Headers").Default("Content-Type", overrides).Label("Access-Control-Allow-Headers").ToString(fs),
		methods:     flags.New(prefix, "cors", "Methods").Default(http.MethodGet, overrides).Label("Access-Control-Allow-Methods").ToString(fs),
		exposes:     flags.New(prefix, "cors", "Expose").Default("", overrides).Label("Access-Control-Expose-Headers").ToString(fs),
		credentials: flags.New(prefix, "cors", "Credentials").Default(false, overrides).Label("Access-Control-Allow-Credentials").ToBool(fs),
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
