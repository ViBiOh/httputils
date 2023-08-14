package tracer

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/model"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	tr "go.opentelemetry.io/otel/trace"
)

var noopFunc = func(*error, ...tr.SpanEndOption) {
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
		url:  flags.New("URL", "OpenTracing gRPC endpoint (e.g. otel-exporter:4317)").Prefix(prefix).DocPrefix("tracing").String(fs, "", overrides),
		rate: flags.New("Rate", "OpenTracing sample rate, 'always', 'never' or a float value").Prefix(prefix).DocPrefix("tracing").String(fs, "always", overrides),
	}
}

func newOtelExporter(ctx context.Context, endpoint string) (trace.SpanExporter, error) {
	return otlptracegrpc.New(ctx,
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(endpoint),
	)
}

func newSampler(rate string) (trace.Sampler, error) {
	switch rate {
	case "always":
		return trace.AlwaysSample(), nil

	case "never":
		return trace.NeverSample(), nil

	default:
		rateRatio, err := strconv.ParseFloat(rate, 64)
		if err != nil {
			return nil, fmt.Errorf("parse sample rate `%s`: %w", rate, err)
		}

		return trace.TraceIDRatioBased(rateRatio), nil
	}
}

func newResource(ctx context.Context) (*resource.Resource, error) {
	newResource, err := resource.New(ctx, resource.WithFromEnv())
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

func New(ctx context.Context, config Config) (App, error) {
	url := strings.TrimSpace(*config.url)

	if len(url) == 0 {
		return App{}, nil
	}

	tracerExporter, err := newOtelExporter(ctx, url)
	if err != nil {
		return App{}, err
	}

	tracerResource, err := newResource(ctx)
	if err != nil {
		return App{}, err
	}

	sampler, err := newSampler(strings.TrimSpace(*config.rate))
	if err != nil {
		return App{}, err
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

func (a App) Close(ctx context.Context) {
	if a.provider == nil {
		return
	}

	if err := a.provider.Shutdown(ctx); err != nil {
		slog.Error("shutdown trace provider", "err", err)
	}
}

func StartSpan(ctx context.Context, tracer tr.Tracer, name string, opts ...tr.SpanStartOption) (context.Context, func(err *error, options ...tr.SpanEndOption)) {
	if tracer == nil {
		return ctx, noopFunc
	}

	ctx, span := tracer.Start(ctx, name, opts...)

	return ctx, func(err *error, options ...tr.SpanEndOption) {
		if err != nil && *err != nil {
			span.SetStatus(codes.Error, (*err).Error())
		}

		span.End(options...)
	}
}

func AddTraceToLogger(span tr.Span, logger *slog.Logger) *slog.Logger {
	if model.IsNil(span) || !span.IsRecording() {
		return logger
	}

	spanCtx := span.SpanContext()

	if spanCtx.HasTraceID() {
		logger = logger.With("traceID", spanCtx.TraceID())
	}

	if spanCtx.HasSpanID() {
		logger = logger.With("spanID", spanCtx.SpanID())
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
