package renderer

import (
	"flag"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/ViBiOh/httputils/v4/pkg/flags"
	"github.com/ViBiOh/httputils/v4/pkg/httperror"
	"github.com/ViBiOh/httputils/v4/pkg/templates"
)

const (
	faviconPath = "/favicon"
	svgPath     = "/svg"
)

var (
	staticRootPaths = []string{"/robots.txt", "/sitemap.xml"}
	staticDir       = "static"
)

// App of package
type App interface {
	Handler(TemplateFunc) http.Handler
	Error(http.ResponseWriter, error)
}

// Config of package
type Config struct {
	templates *string
	statics   *string
	publicURL *string
	title     *string
}

type app struct {
	tpl        *template.Template
	content    map[string]interface{}
	staticsDir string
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) Config {
	return Config{
		templates: flags.New(prefix, "").Name("Templates").Default(flags.Default("Templates", "./templates/", overrides)).Label("HTML Templates folder").ToString(fs),
		statics:   flags.New(prefix, "").Name("Static").Default(flags.Default("Static", "./static/", overrides)).Label("Static folder, content served directly").ToString(fs),
		publicURL: flags.New(prefix, "").Name("PublicURL").Default(flags.Default("PublicURL", "http://localhost", overrides)).Label("Public URL").ToString(fs),
		title:     flags.New(prefix, "").Name("Title").Default(flags.Default("Title", "App", overrides)).Label("Application title").ToString(fs),
	}
}

// New creates new App from Config
func New(config Config, funcMap template.FuncMap, filesystem fs.FS) (App, error) {
	tpl := template.New("app").Funcs(funcMap)

	if filesystem != nil {
		tpl = template.Must(tpl.ParseFS(filesystem, "*.html"))
	} else {
		filesTemplates, err := templates.GetTemplates(strings.TrimSpace(*config.templates), ".html")
		if err != nil {
			return nil, fmt.Errorf("unable to get templates: %s", err)
		}

		tpl = template.Must(tpl.ParseFiles(filesTemplates...))
	}

	return app{
		tpl:        tpl,
		staticsDir: strings.TrimSpace(*config.statics),
		content: map[string]interface{}{
			"PublicURL": strings.TrimSpace(*config.publicURL),
			"Title":     strings.TrimSpace(*config.title),
			"Version":   os.Getenv("VERSION"),
		},
	}, nil
}

func isStaticRootPaths(requestPath string) bool {
	for _, rootPath := range staticRootPaths {
		if strings.EqualFold(rootPath, requestPath) {
			return true
		}
	}

	return false
}

func (a app) feedContent(content map[string]interface{}) map[string]interface{} {
	if content == nil {
		content = make(map[string]interface{})
	}

	for key, value := range a.content {
		content[key] = value
	}

	return content
}

func (a app) Handler(templateFunc TemplateFunc) http.Handler {
	svgHandler := http.StripPrefix(svgPath, a.svg())

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, faviconPath) || isStaticRootPaths(r.URL.Path) {
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

		a.html(w, r, templateFunc)
	})
}
