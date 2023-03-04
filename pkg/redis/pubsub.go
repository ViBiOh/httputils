package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ViBiOh/httputils/v4/pkg/model"
	"github.com/ViBiOh/httputils/v4/pkg/tracer"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/trace"
)

func (a App) PublishJSON(ctx context.Context, channel string, value any) error {
	payload, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	return a.Publish(ctx, channel, payload)
}

func (a App) Publish(ctx context.Context, channel string, value any) (err error) {
	ctx, end := tracer.StartSpan(ctx, a.tracer, "publish", trace.WithSpanKind(trace.SpanKindProducer))
	defer end(&err)

	count, err := a.redisClient.Publish(ctx, channel, value).Result()
	if err != nil {
		return fmt.Errorf("publish: %w", err)
	}

	if count == 0 {
		return ErrNoSubscriber
	}

	return nil
}

func (a App) Subscribe(ctx context.Context, channel string) (<-chan *redis.Message, func(context.Context) error) {
	ctx, end := tracer.StartSpan(ctx, a.tracer, "subscribe", trace.WithSpanKind(trace.SpanKindConsumer))
	defer end(nil)

	pubsub := a.redisClient.Subscribe(ctx, channel)

	return pubsub.Channel(), func(ctx context.Context) (err error) {
		defer func() {
			if closeErr := pubsub.Close(); closeErr != nil {
				err = model.WrapError(err, closeErr)
			}
		}()

		err = pubsub.Unsubscribe(ctx, channel)

		return
	}
}

func SubscribeFor[T any](ctx context.Context, app Client, channel string, handler func(T, error)) (<-chan struct{}, func(context.Context) error) {
	subscription, unsubscribe := app.Subscribe(ctx, channel)

	done := make(chan struct{})

	go func() {
		defer close(done)

		for item := range subscription {
			var instance T
			handler(instance, json.Unmarshal([]byte(item.Payload), &instance))
		}
	}()

	return done, unsubscribe
}
