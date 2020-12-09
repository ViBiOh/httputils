package renderer

import (
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/ViBiOh/httputils/v3/pkg/flags"
	"github.com/ViBiOh/httputils/v3/pkg/httperror"
	"github.com/ViBiOh/httputils/v3/pkg/renderer/model"
	"github.com/ViBiOh/httputils/v3/pkg/templates"
)

const (
	faviconPath = "/favicon"
	svgPath     = "/svg"
)

var (
	rootPaths = []string{"/robots.txt", "/sitemap.xml"}
	staticDir = "static"
)

// App of package
type App interface {
	Handler(model.TemplateFunc) http.Handler
}

// Config of package
type Config struct {
	templates *string
	statics   *string
}

type app struct {
	tpl        *template.Template
	version    string
	staticsDir string
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string) Config {
	return Config{
		templates: flags.New(prefix, "").Name("Templates").Default("./templates/").Label("HTML Templates folder").ToString(fs),
		statics:   flags.New(prefix, "").Name("Static").Default("./static/").Label("Static folder, content served directly").ToString(fs),
	}
}

// New creates new App from Config
func New(config Config, funcMap template.FuncMap) (App, error) {
	filesTemplates, err := templates.GetTemplates(strings.TrimSpace(*config.templates), ".html")
	if err != nil {
		return nil, fmt.Errorf("unable to get templates: %s", err)
	}

	return app{
		tpl:     template.Must(template.New("app").Funcs(funcMap).ParseFiles(filesTemplates...)),
		version: os.Getenv("VERSION"),
	}, nil
}

func isRootPaths(requestPath string) bool {
	for _, rootPath := range rootPaths {
		if strings.EqualFold(rootPath, requestPath) {
			return true
		}
	}

	return false
}

func (a app) Handler(templateFunc model.TemplateFunc) http.Handler {
	svgHandler := http.StripPrefix(svgPath, a.svg())

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, faviconPath) || isRootPaths(r.URL.Path) {
			http.ServeFile(w, r, path.Join(a.staticsDir, r.URL.Path))
			return
		}

		if a.tpl == nil {
			httperror.NotFound(w)
			return
		}

		if strings.HasPrefix(r.URL.Path, svgPath) {
			svgHandler.ServeHTTP(w, r)
			return
		}

		templateName, status, content, err := templateFunc(r)
		if err != nil {
			a.error(w, err)
			return
		}

		content["Version"] = a.version

		message := model.ParseMessage(r)
		if len(message.Content) > 0 {
			content["Message"] = message
		}

		if err := templates.ResponseHTMLTemplate(a.tpl.Lookup(templateName), w, content, status); err != nil {
			httperror.InternalServerError(w, err)
		}
	})
}
