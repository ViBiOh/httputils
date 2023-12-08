package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/ViBiOh/httputils/v4/pkg/cntxt"
	"github.com/ViBiOh/httputils/v4/pkg/concurrent"
	"github.com/redis/go-redis/v9"
)

func (s Service) PublishJSON(ctx context.Context, channel string, value any) error {
	payload, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	return s.Publish(ctx, channel, payload)
}

func (s Service) Publish(ctx context.Context, channel string, value any) (err error) {
	count, err := s.client.Publish(ctx, channel, value).Result()
	if err != nil {
		return fmt.Errorf("publish: %w", err)
	}

	if count == 0 {
		return ErrNoSubscriber
	}

	return nil
}

func (s Service) Subscribe(ctx context.Context, channel string) (<-chan *redis.Message, func(context.Context)) {
	pubsub := s.client.Subscribe(ctx, channel)

	return pubsub.Channel(), func(ctx context.Context) {
		if err := pubsub.Unsubscribe(ctx, channel); err != nil {
			slog.ErrorContext(ctx, "unsubscribe pubsub", "error", err, "channel", channel)
		}

		if err := pubsub.Close(); err != nil {
			slog.ErrorContext(ctx, "close pubsub", "error", err, "channel", channel)
		}
	}
}

func SubscribeFor[T any](ctx context.Context, client Subscriber, channel string, handler func(T, error)) {
	subscription, unsubscribe := client.Subscribe(ctx, channel)

	concurrent.ChanUntilDone(ctx, subscription, func(item *redis.Message) {
		var instance T
		handler(instance, json.Unmarshal([]byte(item.Payload), &instance))
	}, func() {
		unsubscribe(cntxt.WithoutDeadline(ctx))
	})
}
