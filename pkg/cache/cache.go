package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/concurrent"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/httputils/v4/pkg/model"
	"github.com/ViBiOh/httputils/v4/pkg/tracer"
	"go.opentelemetry.io/otel/trace"
)

//go:generate mockgen -source cache.go -destination ../mocks/cache.go -package mocks -mock_names RedisClient=RedisClient

var (
	syncActionTimeout  = time.Millisecond * 150
	asyncActionTimeout = time.Second * 5
	errEmptyContent    = errors.New("empty content")
)

type RedisClient interface {
	Load(ctx context.Context, key string) ([]byte, error)
	LoadMany(ctx context.Context, keys ...string) ([]string, error)
	Store(ctx context.Context, key string, value any, ttl time.Duration) error
	Delete(ctx context.Context, keys ...string) error
}

func Retrieve[T any](ctx context.Context, client RedisClient, onMiss func(context.Context) (T, error), ttl time.Duration, key string) (item T, err error) {
	if model.IsNil(client) {
		return onMiss(ctx)
	}

	loadCtx, cancel := context.WithTimeout(ctx, syncActionTimeout)
	defer cancel()

	content, err := client.Load(loadCtx, key)
	if err != nil {
		loggerWithTrace(loadCtx, key).Error("load from cache: %s", err)
	} else if item, err = unmarshal[T](content); err != nil {
		loggerWithTrace(loadCtx, key).Error("unmarshal from cache: %s", err)
	} else {
		return item, nil
	}

	item, err = onMiss(ctx)

	if err == nil {
		go store(context.Background(), client, key, item, ttl)
	}

	return item, err
}

func RetrieveMany[T any](ctx context.Context, client RedisClient, onMiss func(context.Context, string) (T, error), ttl time.Duration, concurrency uint64, keys ...string) ([]T, error) {
	var values []string
	var err error

	if !model.IsNil(client) {
		loadCtx, cancel := context.WithTimeout(ctx, syncActionTimeout)
		defer cancel()

		values, err = client.LoadMany(loadCtx, keys...)
		if err != nil {
			logger.Error("load many from cache: %s", err)
		}
	}

	valuesLen := len(values)
	wg := concurrent.NewLimited(concurrency)

	output := make([]T, len(keys))
	for index, key := range keys {
		index, key := index, key

		wg.Go(func() {
			if index < valuesLen {
				if item, err := unmarshal[T]([]byte(values[index])); err != nil {
					loggerWithTrace(ctx, key).Error("unmarshal from cache: %s", err)
				} else {
					output[index] = item
					return
				}
			}

			item, err := onMiss(ctx, key)
			if err != nil {
				loggerWithTrace(ctx, key).Error("onMiss to cache: %s", err)
				return
			}

			output[index] = item
			go store(context.Background(), client, key, item, ttl)
		})
	}

	wg.Wait()

	return output, nil
}

func unmarshal[T any](content []byte) (item T, err error) {
	if len(content) == 0 {
		err = errEmptyContent
		return
	}

	return item, json.Unmarshal(content, &item)
}

func store(ctx context.Context, client RedisClient, key string, item any, ttl time.Duration) {
	storeCtx, cancel := context.WithTimeout(ctx, asyncActionTimeout)
	defer cancel()

	if payload, err := json.Marshal(item); err != nil {
		loggerWithTrace(ctx, key).Error("marshal to cache: %s", err)
	} else if err = client.Store(storeCtx, key, payload, ttl); err != nil {
		loggerWithTrace(ctx, key).Error("write to cache: %s", err)
	}
}

func EvictOnSuccess(ctx context.Context, redisClient RedisClient, key string, err error) error {
	if err != nil || model.IsNil(redisClient) {
		return err
	}

	if err = redisClient.Delete(ctx, key); err != nil {
		return fmt.Errorf("evict key `%s` from cache: %w", key, err)
	}

	return nil
}

func loggerWithTrace(ctx context.Context, key string) logger.Provider {
	return tracer.AddTraceToLogger(trace.SpanFromContext(ctx), logger.GetGlobal()).WithField("key", key)
}
