package telemetry

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/ViBiOh/flags"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	meter "go.opentelemetry.io/otel/metric"
	noop_meter "go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	tr "go.opentelemetry.io/otel/trace"
)

type App struct {
	tracerProvider *trace.TracerProvider
	meterProvider  *metric.MeterProvider
}

type Config struct {
	url  *string
	rate *string
}

func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) Config {
	return Config{
		url:  flags.New("URL", "OpenTelemetry gRPC endpoint (e.g. otel-exporter:4317)").Prefix(prefix).DocPrefix("telemetry").String(fs, "", overrides),
		rate: flags.New("Rate", "OpenTelemetry sample rate, 'always', 'never' or a float value").Prefix(prefix).DocPrefix("telemetry").String(fs, "always", overrides),
	}
}

func New(ctx context.Context, config Config) (App, error) {
	url := strings.TrimSpace(*config.url)

	if len(url) == 0 {
		return App{}, nil
	}

	otelResource, err := newResource(ctx)
	if err != nil {
		return App{}, fmt.Errorf("otel resource: %w", err)
	}

	tracerExporter, err := newTraceExporter(ctx, url)
	if err != nil {
		return App{}, fmt.Errorf("trace exporter: %w", err)
	}

	sampler, err := newSampler(strings.TrimSpace(*config.rate))
	if err != nil {
		return App{}, fmt.Errorf("sampler: %w", err)
	}

	tracerProvider := trace.NewTracerProvider(
		trace.WithBatcher(tracerExporter),
		trace.WithResource(otelResource),
		trace.WithSampler(sampler),
	)

	metricExporter, err := newMetricExporter(ctx, url)
	if err != nil {
		return App{}, fmt.Errorf("metric exporter: %w", err)
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithResource(otelResource),
		metric.WithReader(metric.NewPeriodicReader(metricExporter)),
		metric.WithView(
			metric.NewView(
				metric.Instrument{Scope: instrumentation.Scope{
					Name: "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp",
				}},
				metric.Stream{
					AttributeFilter: allowedHttpAttr(
						"http.method",
						"http.status_code",
					),
				},
			),
		),
	)

	if err := runtime.Start(runtime.WithMeterProvider(meterProvider)); err != nil {
		return App{}, fmt.Errorf("runtime: %w", err)
	}

	return App{
		tracerProvider: tracerProvider,
		meterProvider:  meterProvider,
	}, nil
}

func allowedHttpAttr(v ...string) attribute.Filter {
	m := make(map[string]any, len(v))

	for _, s := range v {
		m[s] = struct{}{}
	}

	return func(kv attribute.KeyValue) bool {
		_, ok := m[string(kv.Key)]
		return ok
	}
}

func (a App) MeterProvider() meter.MeterProvider {
	if a.meterProvider == nil {
		return noop_meter.MeterProvider{}
	}

	return a.meterProvider
}

func (a App) TracerProvider() tr.TracerProvider {
	if a.meterProvider == nil {
		return tr.NewNoopTracerProvider()
	}

	return a.tracerProvider
}

func (a App) Middleware(name string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if next == nil {
			return next
		}

		return otelhttp.NewHandler(next, name,
			otelhttp.WithTracerProvider(a.TracerProvider()),
			otelhttp.WithPropagators(propagation.TraceContext{}),
			otelhttp.WithMeterProvider(a.MeterProvider()),
		)
	}
}

func (a App) Close(ctx context.Context) {
	if a.tracerProvider != nil {
		if err := a.tracerProvider.Shutdown(ctx); err != nil {
			slog.Error("shutdown trace provider", "err", err)
		}
	}

	if a.meterProvider != nil {
		if err := a.meterProvider.Shutdown(ctx); err != nil {
			slog.Error("shutdown meter provider", "err", err)
		}
	}
}

func newTraceExporter(ctx context.Context, endpoint string) (trace.SpanExporter, error) {
	return otlptracegrpc.New(ctx,
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(endpoint),
	)
}

func newMetricExporter(ctx context.Context, endpoint string) (metric.Exporter, error) {
	return otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithInsecure(),
		otlpmetricgrpc.WithEndpoint(endpoint),
	)
}

func newResource(ctx context.Context) (*resource.Resource, error) {
	newResource, err := resource.New(ctx, resource.WithFromEnv())
	if err != nil {
		return nil, fmt.Errorf("create resource: %w", err)
	}

	r, err := resource.Merge(resource.Default(), newResource)
	if err != nil {
		return nil, fmt.Errorf("merge resource with default: %w", err)
	}

	return r, nil
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
