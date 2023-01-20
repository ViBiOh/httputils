package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ViBiOh/httputils/v4/pkg/tracer"
	"github.com/go-redis/redis/v8"
	"go.opentelemetry.io/otel/trace"
)

func (a App) Publish(ctx context.Context, channel string, value any) error {
	ctx, end := tracer.StartSpan(ctx, a.tracer, "publish", trace.WithSpanKind(trace.SpanKindProducer))
	defer end()

	count, err := a.redisClient.Publish(ctx, channel, value).Result()
	if err != nil {
		a.increase("error")

		return fmt.Errorf("publish: %w", err)
	}

	if count == 0 {
		return ErrNoSubscriber
	}

	return nil
}

func (a App) Subscribe(ctx context.Context, channel string) (<-chan *redis.Message, func(context.Context) error) {
	ctx, end := tracer.StartSpan(ctx, a.tracer, "subscribe", trace.WithSpanKind(trace.SpanKindConsumer))
	defer end()

	pubsub := a.redisClient.Subscribe(ctx, channel)

	return pubsub.Channel(), func(ctx context.Context) error {
		return pubsub.Unsubscribe(ctx, channel)
	}
}

func SubscribeFor[T any](ctx context.Context, app App, channel string, handler func(T, error)) func(context.Context) error {
	subscription, unsubscribe := app.Subscribe(ctx, channel)

	output := make(chan T, len(subscription))

	go func() {
		defer close(output)

		for item := range subscription {
			var instance T
			handler(instance, json.Unmarshal([]byte(item.Payload), &instance))
		}
	}()

	return unsubscribe
}
