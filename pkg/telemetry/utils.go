package telemetry

import (
	"context"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"

	"github.com/ViBiOh/httputils/v4/pkg/model"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	meter "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	tr "go.opentelemetry.io/otel/trace"
)

type FinishSpan = func(err *error, options ...tr.SpanEndOption)

type redactedURL struct{}

var noopFunc = func(*error, ...tr.SpanEndOption) {
	// Nothing to do
}

type noTrace struct{}

func NoTrace(ctx context.Context) context.Context {
	return context.WithValue(ctx, noTrace{}, true)
}

func IsNoTrace(ctx context.Context) bool {
	value := ctx.Value(noTrace{})
	if value == nil {
		return false
	}

	boolValue, _ := value.(bool)

	return boolValue
}

// WithRedactedURL marks the HTTP requests made with the returned context: the instrumented client records only the scheme and the host of their URL.
// It is meant for URLs that hold a secret, like the webhooks of Slack, Discord or Telegram.
func WithRedactedURL(ctx context.Context) context.Context {
	return context.WithValue(ctx, redactedURL{}, true)
}

func isURLRedacted(ctx context.Context) bool {
	value, _ := ctx.Value(redactedURL{}).(bool)
	return value
}

func RedactURL(target *url.URL) string {
	if target == nil || len(target.Host) == 0 {
		return "<redacted>"
	}

	output := target.Scheme + "://" + target.Host

	if len(target.Path) > 1 || len(target.RawQuery) != 0 || len(target.Fragment) != 0 {
		output += "/<redacted>"
	}

	return output
}

// urlRedactingTransport runs inside the OpenTelemetry transport, once its client span has been started with the full URL.
type urlRedactingTransport struct {
	next http.RoundTripper
}

func (urt urlRedactingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if ctx := req.Context(); isURLRedacted(ctx) {
		tr.SpanFromContext(ctx).SetAttributes(attribute.String("url.full", RedactURL(req.URL)))
	}

	return urt.next.RoundTrip(req)
}

func StartSpan(ctx context.Context, tracer tr.Tracer, name string, opts ...tr.SpanStartOption) (context.Context, FinishSpan) {
	if tracer == nil || IsNoTrace(ctx) {
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

func AddOpenTelemetryToDefaultLogger(telemetry *Service) {
	slog.SetDefault(slog.New(telemetry.AddTraceToLogHandler(slog.Default().Handler())))
}

func AddOpenTelemetryToClient(httpClient *http.Client, meterProvider meter.MeterProvider, tracerProvider tr.TracerProvider) *http.Client {
	if model.IsNil(tracerProvider) {
		return httpClient
	}

	next := httpClient.Transport
	if next == nil {
		next = http.DefaultTransport
	}

	httpClient.Transport = otelhttp.NewTransport(
		urlRedactingTransport{next: next},
		otelhttp.WithTracerProvider(tracerProvider),
		otelhttp.WithMeterProvider(meterProvider),
		otelhttp.WithPropagators(propagation.TraceContext{}),
	)

	return httpClient
}

func uint64TraceId(id string) string {
	if len(id) < 16 {
		return ""
	}

	if len(id) > 16 {
		id = id[16:]
	}

	intValue, err := strconv.ParseUint(id, 16, 64)
	if err != nil {
		return ""
	}

	return strconv.FormatUint(intValue, 10)
}
