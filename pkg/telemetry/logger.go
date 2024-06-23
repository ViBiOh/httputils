package telemetry

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (s *Service) AddTraceToLogHandler(handler slog.Handler) slog.Handler {
	if s == nil {
		return handler
	}

	return OtlpLogger{
		Handler:    handler,
		Uint64:     s.TraceUint64,
		Attributes: otelAttributeToSlogAttr(s.resource.Attributes()),
	}
}

func otelAttributeToSlogAttr(attributes []attribute.KeyValue) []slog.Attr {
	output := make([]slog.Attr, len(attributes))

	for index, attribute := range attributes {
		output[index] = slog.String(string(attribute.Key), attribute.Value.AsString())
	}

	return output
}

type OtlpLogger struct {
	slog.Handler
	Attributes []slog.Attr
	Uint64     bool
}

func (tl OtlpLogger) Handle(ctx context.Context, r slog.Record) error {
	spanCtx := trace.SpanContextFromContext(ctx)

	if spanCtx.HasTraceID() {
		r.AddAttrs(slog.String("trace_id", tl.getTraceID(spanCtx.TraceID().String())))
	}

	if spanCtx.HasSpanID() {
		r.AddAttrs(slog.String("span_id", tl.getTraceID(spanCtx.SpanID().String())))
	}

	r.AddAttrs(tl.Attributes...)

	return tl.Handler.Handle(ctx, r)
}

func (tl OtlpLogger) getTraceID(id string) string {
	if tl.Uint64 {
		return uint64TraceId(id)
	}

	return id
}
