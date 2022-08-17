package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ViBiOh/httputils/v4/pkg/tracer"
	"github.com/go-redis/redis/v8"
)

// Publish a message to a given channel.
func (a App) Publish(ctx context.Context, channel string, value any) error {
	if !a.enabled() {
		return nil
	}

	ctx, end := tracer.StartSpan(ctx, a.tracer, "publish")
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

// Subscribe to a given channel.
func (a App) Subscribe(ctx context.Context, channel string) (<-chan *redis.Message, func(context.Context) error) {
	if !a.enabled() {
		return nil, func(_ context.Context) error { return nil }
	}

	ctx, end := tracer.StartSpan(ctx, a.tracer, "subscribe")
	defer end()

	pubsub := a.redisClient.Subscribe(ctx, channel)

	return pubsub.Channel(), func(ctx context.Context) error {
		return pubsub.Unsubscribe(ctx, channel)
	}
}

// SubscribeFor pubsub with unmarshal of given type.
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
