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

	if spanCtx.HasSpanID() {
		payload.Headers["span_id"] = spanCtx.SpanID().String()
	}

	return payload
}

func FromAmqp(ctx context.Context, message amqp.Delivery) context.Context {
	id, ok := message.Headers["trace_id"]
	if !ok {
		return ctx
	}

	var spanID trace.SpanID
	if otherID, ok := message.Headers["span_id"]; ok {
		spanID = parseSpanID(otherID)
	}

	spanCtx := trace.NewSpanContext(
		trace.SpanContextConfig{
			TraceID: parseTraceID(id),
			SpanID:  spanID,
			Remote:  true,
		},
	)

	return trace.ContextWithSpanContext(ctx, spanCtx)
}

func parseTraceID(input any) trace.TraceID {
	idHex, ok := input.(string)
	if !ok {
		return trace.TraceID{}
	}

	traceID, err := trace.TraceIDFromHex(idHex)
	if err != nil {
		slog.Warn("parse trace id", "error", err, "trace_id", input)

		return trace.TraceID{}
	}

	return traceID
}

func parseSpanID(input any) trace.SpanID {
	idHex, ok := input.(string)
	if !ok {
		return trace.SpanID{}
	}

	spanID, err := trace.SpanIDFromHex(idHex)
	if err != nil {
		slog.Warn("parse trace id", "error", err, "span_id", input)

		return trace.SpanID{}
	}

	return spanID
}
