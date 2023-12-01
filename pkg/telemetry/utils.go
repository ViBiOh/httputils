package telemetry

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/ViBiOh/httputils/v4/pkg/model"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/codes"
	meter "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	tr "go.opentelemetry.io/otel/trace"
)

type FinishSpan = func(err *error, options ...tr.SpanEndOption)

var noopFunc FinishSpan = func(*error, ...tr.SpanEndOption) {
	// Nothing to do
}

func StartSpan(ctx context.Context, tracer tr.Tracer, name string, opts ...tr.SpanStartOption) (context.Context, FinishSpan) {
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

type OtlpLogger struct {
	slog.Handler
}

func AddOpenTelemetryToLogHandler(handler slog.Handler) slog.Handler {
	return OtlpLogger{handler}
}

func (tl OtlpLogger) Handle(ctx context.Context, r slog.Record) error {
	spanCtx := trace.SpanContextFromContext(ctx)

	if spanCtx.HasTraceID() {
		r.AddAttrs(slog.String("trace_id", spanCtx.TraceID().String()))
	}

	if spanCtx.HasSpanID() {
		r.AddAttrs(slog.String("span_id", spanCtx.SpanID().String()))
	}

	return tl.Handler.Handle(ctx, r)
}

func AddOpenTelemetryToClient(httpClient *http.Client, meterProvider meter.MeterProvider, tracerProvider tr.TracerProvider) *http.Client {
	if model.IsNil(tracerProvider) {
		return httpClient
	}

	httpClient.Transport = otelhttp.NewTransport(httpClient.Transport,
		otelhttp.WithTracerProvider(tracerProvider),
		otelhttp.WithMeterProvider(meterProvider),
		otelhttp.WithPropagators(propagation.TraceContext{}),
	)

	return httpClient
}
