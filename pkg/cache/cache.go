package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/httputils/v4/pkg/tracer"
	"go.opentelemetry.io/otel/trace"
)

//go:generate mockgen -source cache.go -destination ../mocks/cache.go -package mocks -mock_names RedisClient=RedisClient

// RedisClient for caching response
type RedisClient interface {
	Load(ctx context.Context, key string) (string, error)
	Store(ctx context.Context, key string, value any, duration time.Duration) error
	Delete(ctx context.Context, keys ...string) error
}

// Retrieve loads an item from the cache for given key or retrieve it (and store it in cache after)
func Retrieve[T any](ctx context.Context, redisClient RedisClient, key string, onMiss func(context.Context) (T, error), duration time.Duration) (item T, err error) {
	content, err := redisClient.Load(ctx, key)
	if err != nil {
		loggerWithTrace(ctx).Error("read from cache: %s", err)
	} else if len(content) != 0 {
		if err = json.Unmarshal([]byte(content), &item); err != nil {
			loggerWithTrace(ctx).Error("unmarshal from cache: %s", err)
		} else {
			return item, nil
		}
	}

	item, err = onMiss(ctx)

	if err == nil {
		go func() {
			storeCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()

			if payload, err := json.Marshal(item); err != nil {
				loggerWithTrace(ctx).Error("marshal to cache: %s", err)
			} else if err = redisClient.Store(storeCtx, key, payload, duration); err != nil {
				loggerWithTrace(ctx).Error("write to cache: %s", err)
			}
		}()
	}

	return item, err
}

// EvictOnSuccess evict the given key if there is no error
func EvictOnSuccess(ctx context.Context, redisClient RedisClient, key string, err error) error {
	if err != nil {
		return err
	}

	if err = redisClient.Delete(ctx, key); err != nil {
		return fmt.Errorf("evict key `%s` from cache: %w", key, err)
	}

	return nil
}

func loggerWithTrace(ctx context.Context) logger.Provider {
	return tracer.AddTraceToLogger(trace.SpanFromContext(ctx), logger.GetGlobal())
}
