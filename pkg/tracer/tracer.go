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

var noopFunc = func(...tr.SpanEndOption) {
	// Nothing to do
}

type App struct {
	provider *trace.TracerProvider
}

type Config struct {
	url  *string
	rate *string
}

func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) Config {
	return Config{
		url:  flags.String(fs, prefix, "tracing", "URL", "Jaeger endpoint URL (e.g. http://jaeger:14268/api/traces)", "", overrides),
		rate: flags.String(fs, prefix, "tracing", "Rate", "Jaeger sample rate, 'always', 'never' or a float value", "always", overrides),
	}
}

func newExporter(url string) (trace.SpanExporter, error) {
	exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))
	if err != nil {
		return nil, fmt.Errorf("create jaeger exporter: %w", err)
	}

	return exporter, nil
}

func newResource() (*resource.Resource, error) {
	newResource, err := resource.New(context.Background(), resource.WithFromEnv())
	if err != nil {
		return nil, fmt.Errorf("create resource: %w", err)
	}

	r, err := resource.Merge(
		resource.Default(),
		newResource,
	)
	if err != nil {
		return nil, fmt.Errorf("merge resource with default: %w", err)
	}

	return r, nil
}

func New(config Config) (App, error) {
	url := strings.TrimSpace(*config.url)

	if len(url) == 0 {
		return App{}, nil
	}

	tracerExporter, err := newExporter(url)
	if err != nil {
		return App{}, err
	}

	tracerResource, err := newResource()
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
			return App{}, fmt.Errorf("parse sample rate `%s`: %w", rate, err)
		}
		sampler = trace.TraceIDRatioBased(rateRatio)
	}

	provider := trace.NewTracerProvider(
		trace.WithBatcher(tracerExporter),
		trace.WithResource(tracerResource),
		trace.WithSampler(sampler),
	)

	return App{
		provider: provider,
	}, nil
}

func (a App) GetProvider() tr.TracerProvider {
	return a.provider
}

func (a App) GetTracer(name string) tr.Tracer {
	if a.provider == nil {
		return nil
	}

	return a.provider.Tracer(name)
}

func (a App) Middleware(next http.Handler) http.Handler {
	if next == nil || a.provider == nil {
		return next
	}

	return otelhttp.NewHandler(next, "http", otelhttp.WithTracerProvider(a.provider), otelhttp.WithPropagators(propagation.TraceContext{}))
}

func (a App) Close() {
	if a.provider == nil {
		return
	}

	if err := a.provider.Shutdown(context.Background()); err != nil {
		logger.Error("shutdown trace provider: %s", err)
	}
}

func StartSpan(ctx context.Context, tracer tr.Tracer, name string, opts ...tr.SpanStartOption) (context.Context, func(options ...tr.SpanEndOption)) {
	if tracer == nil {
		return ctx, noopFunc
	}

	ctx, span := tracer.Start(ctx, name, opts...)

	return ctx, span.End
}

func CopyToBackground(ctx context.Context) context.Context {
	return CloneContext(ctx)
}

func AddTraceToLogger(span tr.Span, logger logger.Provider) logger.Provider {
	if model.IsNil(span) || !span.IsRecording() {
		return logger
	}

	spanCtx := span.SpanContext()

	if spanCtx.HasTraceID() {
		logger = logger.WithField("traceID", spanCtx.TraceID())
	}

	if spanCtx.HasSpanID() {
		logger = logger.WithField("traceID", spanCtx.SpanID())
	}

	return logger
}

func AddTracerToClient(httpClient *http.Client, tracerProvider tr.TracerProvider) *http.Client {
	if model.IsNil(tracerProvider) {
		return httpClient
	}

	httpClient.Transport = otelhttp.NewTransport(httpClient.Transport, otelhttp.WithTracerProvider(tracerProvider), otelhttp.WithPropagators(propagation.TraceContext{}))

	return httpClient
}
