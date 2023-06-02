package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/cntxt"
)

func Load[V any](ctx context.Context, client RedisClient, key string, onMiss func(context.Context) (V, error), ttl time.Duration) (V, error) {
	if !client.Enabled() || IsBypassed(ctx) {
		return onMiss(ctx)
	}

	loadCtx, cancel := context.WithTimeout(ctx, syncActionTimeout)
	defer cancel()

	if content, err := client.Load(loadCtx, key); err != nil {
		if errors.Is(err, context.Canceled) {
			loggerWithTrace(ctx, key).Warn("load from cache: %s", err)
		} else {
			loggerWithTrace(ctx, key).Error("load from cache: %s", err)
		}
	} else if value, ok, err := unmarshal[V](content); err != nil {
		logUnmarshallError(ctx, key, err)
	} else if ok {
		go doInBackground(cntxt.WithoutDeadline(ctx), "extend ttl", func(ctx context.Context) error {
			return client.Expire(ctx, ttl, key)
		})

		return value, nil
	}

	value, err := onMiss(ctx)

	if err == nil {
		go doInBackground(cntxt.WithoutDeadline(ctx), "store", func(ctx context.Context) error {
			return store(ctx, client, key, value, ttl)
		})
	}

	return value, err
}

func EvictOnSucces(ctx context.Context, client RedisClient, key string, err error) error {
	if err != nil || !client.Enabled() {
		return err
	}

	if err = client.Delete(ctx, key); err != nil {
		return fmt.Errorf("evict key `%s` from cache: %w", key, err)
	}

	return nil
}
