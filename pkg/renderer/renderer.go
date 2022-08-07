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
	"time"

	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/httperror"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/httputils/v4/pkg/tracer"
	"go.opentelemetry.io/otel/trace"
)

const (
	svgPath = "/svg"
)

var (
	staticFolders       = []string{"/images", "/scripts", "/styles"}
	staticRootPaths     = []string{"/robots.txt", "/sitemap.xml", "/favicon.ico"}
	staticCacheDuration = fmt.Sprintf("public, max-age=%.0f", (time.Hour * 24).Seconds())
)

// App of package
type App struct {
	tracer           trace.Tracer
	tpl              *template.Template
	content          map[string]any
	staticFileSystem http.FileSystem
	staticHandler    http.Handler
	pathPrefix       string
	publicURL        string
	minify           bool
}

// Config of package
type Config struct {
	publicURL  *string
	pathPrefix *string
	title      *string
	minify     *bool
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) Config {
	return Config{
		publicURL:  flags.String(fs, prefix, "", "PublicURL", "Public URL", "http://localhost:1080", overrides),
		pathPrefix: flags.String(fs, prefix, "", "PathPrefix", "Root Path Prefix", "", overrides),
		title:      flags.String(fs, prefix, "", "Title", "Application title", "App", overrides),
		minify:     flags.Bool(fs, prefix, "", "Minify", "Minify HTML", true, overrides),
	}
}

// New creates new App from Config
func New(config Config, filesystem fs.FS, funcMap template.FuncMap, tracer trace.Tracer) (App, error) {
	staticFS, err := fs.Sub(filesystem, "static")
	if err != nil {
		return App{}, fmt.Errorf("get static/ filesystem: %s", err)
	}

	pathPrefix := strings.TrimSuffix(*config.pathPrefix, "/")
	publicURL := strings.TrimSuffix(*config.publicURL, "/")

	staticFileSystem := http.FS(staticFS)
	staticHandler := http.FileServer(staticFileSystem)

	instance := App{
		tracer:           tracer,
		staticFileSystem: staticFileSystem,
		staticHandler:    staticHandler,
		pathPrefix:       pathPrefix,
		publicURL:        publicURL,
		minify:           *config.minify,
		content: map[string]any{
			"Title":   *config.title,
			"Version": os.Getenv("VERSION"),
		},
	}

	if funcMap == nil {
		funcMap = template.FuncMap{}
	}

	funcMap["url"] = instance.url
	funcMap["publicURL"] = instance.PublicURL

	tpl, err := template.New("app").Funcs(funcMap).ParseFS(filesystem, "templates/*.html")
	if err != nil {
		return App{}, fmt.Errorf("parse templates/*.html templates: %s", err)
	}

	instance.tpl = tpl

	if strings.HasPrefix(instance.publicURL, "http://localhost") {
		logger.Warn("PublicURL has a development/debug value: `%s`. You may need to configure it.", instance.publicURL)
	}

	return instance, nil
}

// PublicURL computes public URL of given path
func (a App) PublicURL(url string) string {
	return a.publicURL + a.url(url)
}

func (a App) url(url string) string {
	prefixedURL := path.Join(a.pathPrefix, url)
	if len(prefixedURL) > 1 && strings.HasSuffix(url, "/") {
		return prefixedURL + "/"
	}

	return prefixedURL
}

func isStaticPaths(requestPath string) bool {
	for _, rootPath := range staticRootPaths {
		if strings.EqualFold(rootPath, requestPath) {
			return true
		}
	}

	for _, folder := range staticFolders {
		if strings.HasPrefix(requestPath, folder) {
			return true
		}
	}

	return false
}

func (a App) feedContent(content map[string]any) map[string]any {
	if content == nil {
		content = make(map[string]any)
	}

	for key, value := range a.content {
		if _, ok := content[key]; !ok {
			content[key] = value
		}
	}

	return content
}

// Handler for request. Should be use with net/http
func (a App) Handler(templateFunc TemplateFunc) http.Handler {
	svgHandler := http.StripPrefix(svgPath, a.svg())

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, end := tracer.StartSpan(r.Context(), a.tracer, "renderer")
		defer end()

		r = r.WithContext(ctx)

		if a.handleStatic(w, r) {
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

		a.render(w, r, templateFunc)
	})

	if len(a.pathPrefix) == 0 {
		return handler
	}

	return http.StripPrefix(a.pathPrefix, handler)
}

func (a App) handleStatic(w http.ResponseWriter, r *http.Request) bool {
	if !isStaticPaths(r.URL.Path) {
		return false
	}

	file, err := a.staticFileSystem.Open(r.URL.Path)
	if err != nil {
		return false
	}

	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			logger.Warn("close static file: %s", err)
		}
	}()

	w.Header().Add("Cache-Control", staticCacheDuration)
	a.staticHandler.ServeHTTP(w, r)
	return true
}
