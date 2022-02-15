package tracer

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/httputils/v4/pkg/model"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	tr "go.opentelemetry.io/otel/trace"
)

// App of package
type App struct {
	provider *trace.TracerProvider
}

// Config of package
type Config struct {
	url  *string
	rate *string
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) Config {
	return Config{
		url:  flags.New(prefix, "tracing", "URL").Default("", overrides).Label("Jaeger endpoint URL (e.g. http://jaeger:14268/api/traces)").ToString(fs),
		rate: flags.New(prefix, "tracing", "Rate").Default("always", overrides).Label("Jaeger sample rate, 'always', 'never' or a float value").ToString(fs),
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

	var sampler trace.Sampler
	switch rate := strings.TrimSpace(*config.rate); rate {
	case "always":
		sampler = trace.AlwaysSample()
	case "never":
		sampler = trace.AlwaysSample()
	default:
		rateRatio, err := strconv.ParseFloat(rate, 64)
		if err != nil {
			return App{}, fmt.Errorf("unable to parse sample rate `%s`: %s", rate, err)
		}
		sampler = trace.TraceIDRatioBased(rateRatio)
	}

	provider := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource),
		trace.WithSampler(sampler),
	)

	return App{
		provider: provider,
	}, nil
}

// GetProvider returns current provider
func (a App) GetProvider() tr.TracerProvider {
	return a.provider
}

// GetTracer return a new tracer
func (a App) GetTracer(name string) tr.Tracer {
	if a.provider == nil {
		return nil
	}

	return a.provider.Tracer(name)
}

// AddTracerToClient add tracer to a given http client
func AddTracerToClient(httpClient *http.Client, tracerProvider tr.TracerProvider) *http.Client {
	if model.IsNil(tracerProvider) {
		return httpClient
	}

	httpClient.Transport = otelhttp.NewTransport(httpClient.Transport, otelhttp.WithTracerProvider(tracerProvider), otelhttp.WithPropagators(propagation.Baggage{}))
	return httpClient
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
	if a.provider == nil {
		return
	}

	if err := a.provider.Shutdown(context.Background()); err != nil {
		logger.Error("unable to shutdown trace provider: %s", err)
	}
}
