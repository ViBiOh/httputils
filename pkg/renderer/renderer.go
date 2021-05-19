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
)

const (
	faviconPath = "/favicon"
	svgPath     = "/svg"
)

var (
	staticRootPaths = []string{"/robots.txt", "/sitemap.xml"}
)

// App of package
type App interface {
	Handler(TemplateFunc) http.Handler
	Redirect(http.ResponseWriter, *http.Request, string, Message)
	Error(http.ResponseWriter, error)
}

// Config of package
type Config struct {
	publicURL  *string
	pathPrefix *string
	title      *string
}

type app struct {
	tpl        *template.Template
	content    map[string]interface{}
	staticFS   fs.FS
	pathPrefix string
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) Config {
	return Config{
		publicURL:  flags.New(prefix, "").Name("PublicURL").Default(flags.Default("PublicURL", "http://localhost", overrides)).Label("Public URL").ToString(fs),
		pathPrefix: flags.New(prefix, "").Name("PathPrefix").Default(flags.Default("PathPrefix", "", overrides)).Label("Root Path Prefix").ToString(fs),
		title:      flags.New(prefix, "").Name("Title").Default(flags.Default("Title", "App", overrides)).Label("Application title").ToString(fs),
	}
}

// New creates new App from Config
func New(config Config, filesystem fs.FS, funcMap template.FuncMap) (App, error) {
	staticFS, err := fs.Sub(filesystem, "static")
	if err != nil {
		return nil, fmt.Errorf("unable to get static/ filesystem: %s", err)
	}

	if funcMap == nil {
		funcMap = template.FuncMap{}
	}

	pathPrefix := strings.TrimSuffix(strings.TrimSpace(*config.pathPrefix), "/")
	funcMap["buildURL"] = func(url string) string {
		prefixedURL := path.Join(pathPrefix, url)
		if strings.HasSuffix(url, "/") {
			return fmt.Sprintf("%s/", prefixedURL)
		}

		return prefixedURL
	}

	tpl, err := template.New("app").Funcs(funcMap).ParseFS(filesystem, "templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("unable to parse templates/*.html templates: %s", err)
	}

	publicURL := strings.TrimSuffix(strings.TrimSpace(*config.publicURL), "/")

	return app{
		tpl:        tpl,
		staticFS:   staticFS,
		pathPrefix: pathPrefix,
		content: map[string]interface{}{
			"PublicURL": publicURL,
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
	fileHandler := http.FileServer(http.FS(a.staticFS))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, faviconPath) || isStaticRootPaths(r.URL.Path) {
			fileHandler.ServeHTTP(w, r)
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

	if len(a.pathPrefix) == 0 {
		return handler
	}

	return http.StripPrefix(a.pathPrefix, handler)
}
