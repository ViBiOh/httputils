package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/tracer"
)

func Load[V any](ctx context.Context, client RedisClient, key string, onMiss func(context.Context) (V, error), ttl time.Duration) (V, error) {
	if !client.Enabled() {
		return onMiss(ctx)
	}

	loadCtx, cancel := context.WithTimeout(ctx, syncActionTimeout)
	defer cancel()

	if content, err := client.Load(loadCtx, key); err != nil {
		if errors.Is(err, context.Canceled) {
			loggerWithTrace(ctx, key).Warn("load: %s", err)
		} else {
			loggerWithTrace(ctx, key).Error("load: %s", err)
		}
	} else if value, ok, err := unmarshal[V](ctx, content); err != nil {
		logUnmarshallError(ctx, key, err)
	} else if ok {
		return value, nil
	}

	value, err := onMiss(ctx)

	if err == nil {
		go func(ctx context.Context) {
			payload, err := json.Marshal(value)
			if err != nil {
				loggerWithTrace(ctx, key).Error("marshal: %s", err)
				return
			}

			storeCtx, cancel := context.WithTimeout(ctx, asyncActionTimeout)
			defer cancel()

			if err = client.Store(storeCtx, key, payload, ttl); err != nil {
				loggerWithTrace(ctx, key).Error("store: %s", err)
			}
		}(tracer.CopyToBackground(ctx))
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