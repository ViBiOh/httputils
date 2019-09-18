package cors

import (
	"flag"
	"net/http"
	"strconv"

	"github.com/ViBiOh/httputils/v2/pkg/tools"
)

// Config of package
type Config struct {
	origin      *string
	headers     *string
	methods     *string
	exposes     *string
	credentials *bool
}

// App of package
type App struct {
	origin      string
	headers     string
	methods     string
	exposes     string
	credentials string
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string) Config {
	return Config{
		origin:      tools.NewFlag(prefix, "cors").Name("Origin").Default("*").Label("Access-Control-Allow-Origin").ToString(fs),
		headers:     tools.NewFlag(prefix, "cors").Name("Headers").Default("Content-Type").Label("Access-Control-Allow-Headers").ToString(fs),
		methods:     tools.NewFlag(prefix, "cors").Name("Methods").Default(http.MethodGet).Label("Access-Control-Allow-Methods").ToString(fs),
		exposes:     tools.NewFlag(prefix, "cors").Name("Expose").Default("").Label("Access-Control-Expose-Headers").ToString(fs),
		credentials: tools.NewFlag(prefix, "cors").Name("Credentials").Default(false).Label("Access-Control-Allow-Credentials").ToBool(fs),
	}
}

// New creates new App from Config
func New(config Config) *App {
	return &App{
		origin:      *config.origin,
		headers:     *config.headers,
		methods:     *config.methods,
		exposes:     *config.exposes,
		credentials: strconv.FormatBool(*config.credentials),
	}
}

// Handler for net/http package allowing cors header
func (a App) Handler(next http.Handler) http.Handler {
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
