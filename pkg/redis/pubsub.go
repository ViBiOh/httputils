package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

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

func (s Service) Subscribe(ctx context.Context, channel string) (<-chan *redis.Message, func(context.Context) error) {
	pubsub := s.client.Subscribe(ctx, channel)

	return pubsub.Channel(), func(ctx context.Context) (err error) {
		defer func() {
			if closeErr := pubsub.Close(); closeErr != nil {
				err = errors.Join(err, closeErr)
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
