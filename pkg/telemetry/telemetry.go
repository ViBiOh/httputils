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
	"github.com/ViBiOh/httputils/v4/pkg/model"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	meter "go.opentelemetry.io/otel/metric"
	noopmeter "go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	tr "go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

var propagator propagation.TraceContext

type Service struct {
	resource       *resource.Resource
	tracerProvider *trace.TracerProvider
	meterProvider  *metric.MeterProvider
	TraceUint64    bool
}

type Config struct {
	URL         string
	Rate        string
	TraceUint64 bool
}

func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) *Config {
	var config Config

	flags.New("URL", "OpenTelemetry gRPC endpoint (e.g. otel-exporter:4317)").Prefix(prefix).DocPrefix("telemetry").StringVar(fs, &config.URL, "", overrides)
	flags.New("Rate", "OpenTelemetry sample rate, 'always', 'never' or a float value").Prefix(prefix).DocPrefix("telemetry").StringVar(fs, &config.Rate, "always", overrides)
	flags.New("Uint64", "Change OpenTelemetry Trace ID format to an unsigned int 64").Prefix(prefix).DocPrefix("telemetry").BoolVar(fs, &config.TraceUint64, true, overrides)

	return &config
}

func New(ctx context.Context, config *Config) (*Service, error) {
	if len(config.URL) == 0 {
		return nil, nil
	}

	otelResource, err := newResource(ctx)
	if err != nil {
		return nil, fmt.Errorf("otel resource: %w", err)
	}

	tracerExporter, err := newTraceExporter(ctx, config.URL)
	if err != nil {
		return nil, fmt.Errorf("trace exporter: %w", err)
	}

	sampler, err := newSampler(strings.TrimSpace(config.Rate))
	if err != nil {
		return nil, fmt.Errorf("sampler: %w", err)
	}

	tracerProvider := trace.NewTracerProvider(
		trace.WithBatcher(tracerExporter),
		trace.WithResource(otelResource),
		trace.WithSampler(sampler),
	)

	metricExporter, err := newMetricExporter(ctx, config.URL)
	if err != nil {
		return nil, fmt.Errorf("metric exporter: %w", err)
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
		return nil, fmt.Errorf("runtime: %w", err)
	}

	return &Service{
		resource:       otelResource,
		tracerProvider: tracerProvider,
		meterProvider:  meterProvider,
	}, nil
}

func (s *Service) MeterProvider() meter.MeterProvider {
	if s == nil || s.meterProvider == nil {
		return noopmeter.MeterProvider{}
	}

	return s.meterProvider
}

func (s *Service) TracerProvider() tr.TracerProvider {
	if s == nil || s.meterProvider == nil {
		return noop.NewTracerProvider()
	}

	return s.tracerProvider
}

func (s *Service) Middleware(name string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if next == nil {
			return next
		}

		return otelhttp.NewHandler(AsyncRouteTagMiddleware(next), name,
			otelhttp.WithTracerProvider(s.TracerProvider()),
			otelhttp.WithPropagators(propagator),
			otelhttp.WithMeterProvider(s.MeterProvider()),
		)
	}
}

func (s *Service) Close(ctx context.Context) {
	if s == nil {
		return
	}

	if s.tracerProvider != nil {
		if err := s.tracerProvider.ForceFlush(ctx); err != nil {
			slog.LogAttrs(ctx, slog.LevelError, "flush trace provider", slog.Any("error", err))
		}

		if err := s.tracerProvider.Shutdown(ctx); err != nil {
			slog.LogAttrs(ctx, slog.LevelError, "shutdown trace provider", slog.Any("error", err))
		}
	}

	if s.meterProvider != nil {
		if err := s.meterProvider.ForceFlush(ctx); err != nil {
			slog.LogAttrs(ctx, slog.LevelError, "flush meter provider", slog.Any("error", err))
		}

		if err := s.meterProvider.Shutdown(ctx); err != nil {
			slog.LogAttrs(ctx, slog.LevelError, "shutdown meter provider", slog.Any("error", err))
		}
	}
}

func (s *Service) GetServiceVersionAndEnv() (service, version, env string) {
	if s == nil {
		return "", "", ""
	}

	for _, attr := range s.resource.Attributes() {
		switch attr.Key {
		case semconv.ServiceNameKey:
			service = attr.Value.AsString()

		case semconv.ServiceVersionKey:
			version = attr.Value.AsString()

		case "env":
			env = attr.Value.AsString()
		}
	}

	return
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
	newResource, err := resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithAttributes(
			semconv.ServiceVersion(model.Version()),
			attribute.String("git.commit.sha", model.GitSha()),
		),
	)
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

func allowedHttpAttr(v ...string) attribute.Filter {
	m := make(map[string]struct{}, len(v))

	for _, s := range v {
		m[s] = struct{}{}
	}

	return func(kv attribute.KeyValue) bool {
		_, ok := m[string(kv.Key)]
		return ok
	}
}
