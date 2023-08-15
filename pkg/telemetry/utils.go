package telemetry

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/ViBiOh/httputils/v4/pkg/model"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	tr "go.opentelemetry.io/otel/trace"
)

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
