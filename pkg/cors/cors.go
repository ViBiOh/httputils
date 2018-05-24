package cors

import (
	"flag"
	"fmt"
	"net/http"
	"strconv"

	"github.com/ViBiOh/httputils/pkg/tools"
)

// App stores informations
type App struct {
	origin      string
	headers     string
	methods     string
	exposes     string
	credentials string
}

// NewApp creates new App from Flags' config
func NewApp(config map[string]interface{}) *App {
	return &App{
		origin:      *(config[`origin`].(*string)),
		headers:     *(config[`headers`].(*string)),
		methods:     *(config[`methods`].(*string)),
		exposes:     *(config[`exposes`].(*string)),
		credentials: strconv.FormatBool(*(config[`credentials`].(*bool))),
	}
}

// Flags adds flags for given prefix
func Flags(prefix string) map[string]interface{} {
	return map[string]interface{}{
		`origin`:      flag.String(tools.ToCamel(fmt.Sprintf(`%sOrigin`, prefix)), `*`, `[cors] Access-Control-Allow-Origin`),
		`headers`:     flag.String(tools.ToCamel(fmt.Sprintf(`%sHeaders`, prefix)), `Content-Type`, `[cors] Access-Control-Allow-Headers`),
		`methods`:     flag.String(tools.ToCamel(fmt.Sprintf(`%sMethods`, prefix)), http.MethodGet, `[cors] Access-Control-Allow-Methods`),
		`exposes`:     flag.String(tools.ToCamel(fmt.Sprintf(`%sExpose`, prefix)), ``, `[cors] Access-Control-Expose-Headers`),
		`credentials`: flag.Bool(tools.ToCamel(fmt.Sprintf(`%sCredentials`, prefix)), false, `[cors] Access-Control-Allow-Credentials`),
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
