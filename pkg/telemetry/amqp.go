package telemetry

import (
	"context"
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel/trace"
)

func AddToAmqp(ctx context.Context, payload amqp.Publishing) amqp.Publishing {
	payload.Headers["otel_link"] = trace.LinkFromContext(ctx)

	return payload
}

func GetAmqpLink(ctx context.Context, message amqp.Delivery) trace.SpanStartOption {
	rawLink, ok := message.Headers["otel_link"]
	if !ok {
		return nil
	}

	link, ok := rawLink.(trace.Link)
	if !ok {
		slog.Warn("link is not in expected format")

		return nil
	}

	return trace.WithLinks(link)
}
