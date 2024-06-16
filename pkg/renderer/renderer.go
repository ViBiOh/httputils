package renderer

import (
	"context"
	"flag"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/httperror"
	"github.com/ViBiOh/httputils/v4/pkg/telemetry"
	"github.com/ViBiOh/httputils/v4/pkg/templates"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

var (
	staticFolders       = []string{"/images", "/scripts", "/styles"}
	staticRootPaths     = []string{"/robots.txt", "/sitemap.xml", "/favicon.ico"}
	staticCacheDuration = fmt.Sprintf("public, max-age=%.0f", (time.Hour * 24 * 180).Seconds())
)

type Service struct {
	tracer           trace.Tracer
	tpl              *template.Template
	content          map[string]any
	staticFileSystem http.FileSystem
	staticHandler    http.Handler
	generatedMeter   metric.Int64Counter
	pathPrefix       string
	publicURL        string
	minify           bool
}

type Config struct {
	PublicURL  string
	PathPrefix string
	Title      string
	Minify     bool
}

func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) *Config {
	var config Config

	flags.New("PublicURL", "Public URL").Prefix(prefix).DocPrefix("").StringVar(fs, &config.PublicURL, "http://localhost:1080", overrides)
	flags.New("PathPrefix", "Root Path Prefix").Prefix(prefix).DocPrefix("").StringVar(fs, &config.PathPrefix, "", overrides)
	flags.New("Title", "Application title").Prefix(prefix).DocPrefix("").StringVar(fs, &config.Title, "App", overrides)
	flags.New("Minify", "Minify HTML").Prefix(prefix).DocPrefix("").BoolVar(fs, &config.Minify, true, overrides)

	return &config
}

func New(ctx context.Context, config *Config, filesystem fs.FS, funcMap template.FuncMap, meterProvider metric.MeterProvider, tracerProvider trace.TracerProvider) (*Service, error) {
	staticFS, err := fs.Sub(filesystem, "static")
	if err != nil {
		return nil, fmt.Errorf("get static/ filesystem: %w", err)
	}

	pathPrefix := strings.TrimSuffix(config.PathPrefix, "/")
	publicURL := strings.TrimSuffix(config.PublicURL, "/")

	staticFileSystem := http.FS(staticFS)
	staticHandler := http.FileServer(staticFileSystem)

	instance := Service{
		staticFileSystem: staticFileSystem,
		staticHandler:    staticHandler,
		pathPrefix:       pathPrefix,
		publicURL:        publicURL,
		minify:           config.Minify,
		content: map[string]any{
			"Title":   config.Title,
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
		return nil, fmt.Errorf("parse templates/*.html templates: %w", err)
	}

	instance.tpl = tpl

	if strings.HasPrefix(instance.publicURL, "http://localhost") {
		slog.LogAttrs(ctx, slog.LevelWarn, "PublicURL has a development/debug value: You may need to configure it.", slog.String("url", instance.publicURL))
	}

	if meterProvider != nil {
		meter := meterProvider.Meter("github.com/ViBiOh/httputils/v4/pkg/renderer")

		instance.generatedMeter, err = meter.Int64Counter("templates.generated")
		if err != nil {
			return nil, fmt.Errorf("create generated meter: %w", err)
		}
	}

	if tracerProvider != nil {
		instance.tracer = tracerProvider.Tracer("renderer")
	}

	return &instance, nil
}

func (s *Service) Register(mux *http.ServeMux, templateFunc TemplateFunc) {
	mux.Handle(path.Join(s.pathPrefix, "/svg/{path...}"), s.HandleSVG())

	for _, folder := range staticFolders {
		mux.Handle(path.Join(s.pathPrefix, folder, "{path...}"), s.HandleStatic(folder))
	}

	for _, item := range staticRootPaths {
		mux.Handle(path.Join(s.pathPrefix, item), s.HandleStatic(item))
	}

	mux.Handle(path.Join(s.pathPrefix, "/"), s.Handler(templateFunc))
}

func (s *Service) PublicURL(url string) string {
	return s.publicURL + s.url(url)
}

func (s *Service) url(url string) string {
	prefixedURL := path.Join(s.pathPrefix, url)
	if len(prefixedURL) > 1 && strings.HasSuffix(url, "/") {
		return prefixedURL + "/"
	}

	return prefixedURL
}

func (s *Service) feedContent(content map[string]any) map[string]any {
	if content == nil {
		content = make(map[string]any)
	}

	for key, value := range s.content {
		if _, ok := content[key]; !ok {
			content[key] = value
		}
	}

	return content
}

func (s *Service) Handler(templateFunc TemplateFunc) http.Handler {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, end := telemetry.StartSpan(r.Context(), s.tracer, "renderer", trace.WithSpanKind(trace.SpanKindInternal))
		defer end(nil)

		r = r.WithContext(ctx)

		if s.tpl == nil {
			httperror.NotFound(ctx, w)

			return
		}

		s.render(w, r, templateFunc)
	})

	if len(s.pathPrefix) == 0 {
		return handler
	}

	return http.StripPrefix(s.pathPrefix, handler)
}

func (s *Service) HandleStatic(prefix string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		telemetry.SetRouteTag(ctx, "/static")

		item := r.PathValue("path")
		if len(prefix) != 0 {
			item = path.Join(prefix, item)
		}

		file, err := s.staticFileSystem.Open(item)
		if err != nil {
			httperror.NotFound(ctx, w)

			return
		}

		defer func() {
			if closeErr := file.Close(); closeErr != nil {
				slog.LogAttrs(r.Context(), slog.LevelWarn, "close static file", slog.Any("error", err))
			}
		}()

		w.Header().Add("Cache-Control", staticCacheDuration)
		s.staticHandler.ServeHTTP(w, r)
	})
}

func (s Service) HandleSVG() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		telemetry.SetRouteTag(ctx, "/svg")

		tpl := s.tpl.Lookup("svg-" + strings.Trim(r.PathValue("path"), "/"))
		if tpl == nil {
			httperror.NotFound(ctx, w)

			return
		}

		w.Header().Add("Cache-Control", staticCacheDuration)
		w.Header().Add("Content-Type", "image/svg+xml")

		if err := templates.WriteTemplate(ctx, s.tracer, tpl, w, r.URL.Query().Get("fill"), "text/xml"); err != nil {
			httperror.InternalServerError(ctx, w, err)
		}
	})
}
