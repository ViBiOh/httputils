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
	return Config{
		origin:      fs.String(tools.ToCamel(fmt.Sprintf(`%sOrigin`, prefix)), `*`, `[cors] Access-Control-Allow-Origin`),
		headers:     fs.String(tools.ToCamel(fmt.Sprintf(`%sHeaders`, prefix)), `Content-Type`, `[cors] Access-Control-Allow-Headers`),
		methods:     fs.String(tools.ToCamel(fmt.Sprintf(`%sMethods`, prefix)), http.MethodGet, `[cors] Access-Control-Allow-Methods`),
		exposes:     fs.String(tools.ToCamel(fmt.Sprintf(`%sExpose`, prefix)), ``, `[cors] Access-Control-Expose-Headers`),
		credentials: fs.Bool(tools.ToCamel(fmt.Sprintf(`%sCredentials`, prefix)), false, `[cors] Access-Control-Allow-Credentials`),
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
		w.Header().Set(`Access-Control-Allow-Origin`, a.origin)
		w.Header().Set(`Access-Control-Allow-Headers`, a.headers)
		w.Header().Set(`Access-Control-Allow-Methods`, a.methods)
		w.Header().Set(`Access-Control-Expose-Headers`, a.exposes)
		w.Header().Set(`Access-Control-Allow-Credentials`, a.credentials)

		next.ServeHTTP(w, r)
	})
}
