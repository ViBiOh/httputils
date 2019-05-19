package cors

import (
	"flag"
	"fmt"
	"net/http"
	"strconv"

	"github.com/ViBiOh/httputils/pkg/tools"
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
	docPrefix := prefix
	if prefix == "" {
		docPrefix = "cors"
	}

	return Config{
		origin:      fs.String(tools.ToCamel(fmt.Sprintf("%sOrigin", prefix)), "*", fmt.Sprintf("[%s] Access-Control-Allow-Origin", docPrefix)),
		headers:     fs.String(tools.ToCamel(fmt.Sprintf("%sHeaders", prefix)), "Content-Type", fmt.Sprintf("[%s] Access-Control-Allow-Headers", docPrefix)),
		methods:     fs.String(tools.ToCamel(fmt.Sprintf("%sMethods", prefix)), http.MethodGet, fmt.Sprintf("[%s] Access-Control-Allow-Methods", docPrefix)),
		exposes:     fs.String(tools.ToCamel(fmt.Sprintf("%sExpose", prefix)), "", fmt.Sprintf("[%s] Access-Control-Expose-Headers", docPrefix)),
		credentials: fs.Bool(tools.ToCamel(fmt.Sprintf("%sCredentials", prefix)), false, fmt.Sprintf("[%s] Access-Control-Allow-Credentials", docPrefix)),
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
