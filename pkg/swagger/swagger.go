package swagger

import (
	"flag"
	"fmt"
	"net/http"
	"strings"

	"github.com/ViBiOh/httputils/pkg/httperror"
	"github.com/ViBiOh/httputils/pkg/tools"
)

// Config of package
type Config struct {
	path *string
}

// App of package
type App struct {
	path string
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string) Config {
	return Config{
		path: fs.String(tools.ToCamel(fmt.Sprintf("%sPathPrefix", prefix)), fmt.Sprintf("/%s", prefix), fmt.Sprintf("[%s] HTTP Path prefix", prefix)),
	}
}

// New creates new App from Config
func New(config Config) *App {
	return &App{
		path: strings.TrimSpace(*config.path),
	}
}

// Handler for request. Should be use with net/http
func (a App) Handler() http.Handler {
	return http.StripPrefix(a.path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Content-Type", "text/html")
		w.Header().Set("X-UA-Compatible", "IE=edge,chrome=1")

		if _, err := w.Write(index); err != nil {
			httperror.InternalServerError(w, err)
		}
	}))
}
