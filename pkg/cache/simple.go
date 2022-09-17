package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/model"
)

func Load[V any](ctx context.Context, client RedisClient, key string, onMiss func(context.Context) (V, error), ttl time.Duration) (V, error) {
	if model.IsNil(client) {
		return onMiss(ctx)
	}

	loadCtx, cancel := context.WithTimeout(ctx, syncActionTimeout)
	defer cancel()

	if content, err := client.Load(loadCtx, key); err != nil {
		loggerWithTrace(ctx, key).Error("load: %s", err)
	} else if value, ok := unmarshal[V](ctx, key, content); ok {
		return value, nil
	}

	value, err := onMiss(ctx)

	if err == nil {
		go func() {
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
		}()
	}

	return value, err
}

func EvictOnSucces(ctx context.Context, client RedisClient, key string, err error) error {
	if err != nil || model.IsNil(client) {
		return err
	}

	if err = client.Delete(ctx, key); err != nil {
		return fmt.Errorf("evict key `%s` from cache: %w", key, err)
	}

	return nil
}
