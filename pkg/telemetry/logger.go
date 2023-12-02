package telemetry

import (
	"context"
	"log/slog"
	"strconv"

	"go.opentelemetry.io/otel/trace"
)

func (s Service) AddTraceToLogHandler(handler slog.Handler) slog.Handler {
	return OtlpLogger{
		Handler: handler,
		Uint64:  s.TraceUint64,
	}
}

type OtlpLogger struct {
	slog.Handler
	Uint64 bool
}

func (tl OtlpLogger) getTraceID(id string) string {
	if tl.Uint64 {
		return uint64TraceId(id)
	}

	return id
}

func (tl OtlpLogger) Handle(ctx context.Context, r slog.Record) error {
	spanCtx := trace.SpanContextFromContext(ctx)

	if spanCtx.HasTraceID() {
		r.AddAttrs(slog.String("trace_id", tl.getTraceID(spanCtx.TraceID().String())))
	}

	if spanCtx.HasSpanID() {
		r.AddAttrs(slog.String("span_id", tl.getTraceID(spanCtx.SpanID().String())))
	}

	return tl.Handler.Handle(ctx, r)
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
