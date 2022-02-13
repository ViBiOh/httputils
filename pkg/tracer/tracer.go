package tracer

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"strings"

	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
)

// App of package
type App struct {
	provider *trace.TracerProvider
}

// Config of package
type Config struct {
	url *string
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) Config {
	return Config{
		url: flags.New(prefix, "tracing", "URL").Default("http://localhost:14268/api/traces", overrides).Label("Jaeger endpoint URL").ToString(fs),
	}
}

func newExporter(url string) (trace.SpanExporter, error) {
	exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))
	if err != nil {
		return nil, fmt.Errorf("unable to create jaeger exporter: %s", err)
	}

	return exporter, nil
}

func newResource() (*resource.Resource, error) {
	newResource, err := resource.New(context.Background(), resource.WithFromEnv())
	if err != nil {
		return nil, fmt.Errorf("unable to create resource: %s", err)
	}

	r, err := resource.Merge(
		resource.Default(),
		newResource,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to merge resource with default: %s", err)
	}

	return r, nil
}

// New creates new App from Config
func New(config Config) (App, error) {
	url := strings.TrimSpace(*config.url)

	if len(url) == 0 {
		return App{}, nil
	}

	exporter, err := newExporter(url)
	if err != nil {
		return App{}, err
	}

	resource, err := newResource()
	if err != nil {
		return App{}, err
	}

	provider := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource),
		trace.WithSampler(trace.AlwaysSample()),
	)

	return App{
		provider: provider,
	}, nil
}

// Middleware for net/http package allowing tracer with open telemetry
func (a App) Middleware(next http.Handler) http.Handler {
	if next == nil || a.provider == nil {
		return next
	}

	return otelhttp.NewHandler(next, "http", otelhttp.WithTracerProvider(a.provider))
}

// Close shutdowns tracer provider gracefully
func (a App) Close() {
	if err := a.provider.Shutdown(context.Background()); err != nil {
		logger.Error("unable to shutdown trace provider: %s", err)
	}
}
