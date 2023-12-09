package telemetry

import (
	"context"
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel/trace"
)

func AddToAmqp(ctx context.Context, payload amqp.Publishing) amqp.Publishing {
	spanCtx := trace.SpanContextFromContext(ctx)

	if spanCtx.HasTraceID() {
		payload.Headers["trace_id"] = spanCtx.TraceID().String()
	}

	return payload
}

func FromAmqp(ctx context.Context, message amqp.Delivery) context.Context {
	id, ok := message.Headers["trace_id"]
	if !ok {
		return ctx
	}

	idHex, ok := id.(string)
	if !ok {
		return ctx
	}

	traceID, err := trace.TraceIDFromHex(idHex)
	if err != nil {
		slog.Warn("parse trace id", "error", err, "trace_id", id)

		return ctx
	}

	spanCtx := trace.NewSpanContext(
		trace.SpanContextConfig{
			TraceID: traceID,
			Remote:  true,
		},
	)

	return trace.ContextWithSpanContext(ctx, spanCtx)
}
