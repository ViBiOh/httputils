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

type App[K any, V any] struct {
	client RedisClient
	toKey  func(K) string
	onMiss func(context.Context, K) (V, error)
	ttl    time.Duration
}

func New[K any, V any](client RedisClient, toKey func(K) string, onMiss func(context.Context, K) (V, error), ttl time.Duration) App[K, V] {
	return App[K, V]{
		client: client,
		toKey:  toKey,
		onMiss: onMiss,
		ttl:    ttl,
	}
}

func (a App[K, V]) Get(ctx context.Context, item K) (V, error) {
	if model.IsNil(a.client) {
		return a.onMiss(ctx, item)
	}

	loadCtx, cancel := context.WithTimeout(ctx, syncActionTimeout)
	defer cancel()

	key := a.toKey(item)

	content, err := a.client.Load(loadCtx, key)
	if err != nil {
		loggerWithTrace(loadCtx, key).Error("load from cache: %s", err)
	} else if value, err := unmarshal[V](content); err != nil {
		loggerWithTrace(loadCtx, key).Error("unmarshal from cache: %s", err)
	} else {
		return value, nil
	}

	value, err := a.onMiss(ctx, item)

	if err == nil {
		go store(context.Background(), a.client, key, value, a.ttl)
	}

	return value, err
}

func (a App[K, V]) List(ctx context.Context, concurrency uint64, items ...K) ([]V, error) {
	var values []string
	var err error

	if !model.IsNil(a.client) {
		loadCtx, cancel := context.WithTimeout(ctx, syncActionTimeout)
		defer cancel()

		keys := make([]string, len(items))
		for index, id := range items {
			keys[index] = a.toKey(id)
		}

		values, err = a.client.LoadMany(loadCtx, keys...)
		if err != nil {
			logger.Error("load many from cache: %s", err)
		}
	}

	valuesLen := len(values)
	wg := concurrent.NewLimited(concurrency)

	output := make([]V, len(items))
	for index, item := range items {
		index, item := index, item

		wg.Go(func() {
			if index < valuesLen {
				if value, err := unmarshal[V]([]byte(values[index])); err != nil {
					loggerWithTrace(ctx, a.toKey(item)).Error("unmarshal from cache: %s", err)
				} else {
					output[index] = value
					return
				}
			}

			value, err := a.onMiss(ctx, item)
			if err != nil {
				loggerWithTrace(ctx, a.toKey(item)).Error("onMiss to cache: %s", err)
				return
			}

			output[index] = value
			go store(context.Background(), a.client, a.toKey(item), value, a.ttl)
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
