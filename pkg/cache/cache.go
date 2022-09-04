package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/concurrent"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/httputils/v4/pkg/model"
	"github.com/ViBiOh/httputils/v4/pkg/tracer"
	"go.opentelemetry.io/otel/trace"
)

//go:generate mockgen -source cache.go -destination ../mocks/cache.go -package mocks -mock_names RedisClient=RedisClient

var (
	ErrIgnore = errors.New("ignore error")

	syncActionTimeout  = time.Millisecond * 150
	asyncActionTimeout = time.Second * 5
)

type RedisClient interface {
	Load(ctx context.Context, key string) ([]byte, error)
	LoadMany(ctx context.Context, keys ...string) ([]string, error)
	Store(ctx context.Context, key string, value any, ttl time.Duration) error
	Delete(ctx context.Context, keys ...string) error
}

type App[K any, V any] struct {
	tracer      trace.Tracer
	client      RedisClient
	toKey       func(K) string
	onMiss      func(context.Context, K) (V, error)
	ttl         time.Duration
	concurrency uint64
}

func New[K any, V any](client RedisClient, toKey func(K) string, onMiss func(context.Context, K) (V, error), ttl time.Duration, concurrency uint64, tracer trace.Tracer) App[K, V] {
	return App[K, V]{
		client:      client,
		toKey:       toKey,
		onMiss:      onMiss,
		ttl:         ttl,
		concurrency: concurrency,
	}
}

func (a App[K, V]) Get(ctx context.Context, id K) (V, error) {
	if model.IsNil(a.client) {
		return a.onMiss(ctx, id)
	}

	ctx, end := tracer.StartSpan(ctx, a.tracer, "get")
	defer end()

	key := a.toKey(id)

	loadCtx, cancel := context.WithTimeout(ctx, syncActionTimeout)
	defer cancel()

	content, err := a.client.Load(loadCtx, key)
	if len(content) > 0 {
		var value V

		err := json.Unmarshal(content, &value)
		if err == nil {
			return value, nil
		}

		loggerWithTrace(ctx, key).Error("unmarshal from cache: %s", err)
	} else if err != nil {
		loggerWithTrace(ctx, key).Error("load from cache: %s", err)
	}

	value, err := a.onMiss(ctx, id)

	if err == nil {
		go func() {
			if storeErr := a.store(context.Background(), key, value); storeErr != nil {
				loggerWithTrace(ctx, key).Error("store to cache: %s", storeErr)
			}
		}()
	} else if errors.Is(err, ErrIgnore) {
		err = nil
	}

	return value, err
}

func (a App[K, V]) List(ctx context.Context, items ...K) ([]V, error) {
	ctx, end := tracer.StartSpan(ctx, a.tracer, "list")
	defer end()

	values := a.getValues(ctx, items)
	valuesLen := len(values)
	wg := concurrent.NewLimited(a.concurrency)

	output := make([]V, len(items))
	for index, item := range items {
		index, item := index, item

		wg.Go(func() {
			if index < valuesLen && len(values[index]) > 0 {
				var value V

				err := json.Unmarshal([]byte(values[index]), &value)
				if err == nil {
					output[index] = value
					return
				}

				loggerWithTrace(ctx, a.toKey(item)).Error("unmarshal from cache: %s", err)
			}

			value, err := a.onMiss(ctx, item)
			if err != nil {
				if !errors.Is(err, ErrIgnore) {
					loggerWithTrace(ctx, a.toKey(item)).Error("onMiss to cache: %s", err)
				}
				return
			}

			output[index] = value

			go func() {
				if storeErr := a.Store(context.Background(), item, value); storeErr != nil {
					loggerWithTrace(ctx, a.toKey(item)).Error("store to cache: %s", storeErr)
				}
			}()
		})
	}

	wg.Wait()

	return output, nil
}

func (a App[K, V]) EvictOnSuccess(ctx context.Context, item K, err error) error {
	ctx, end := tracer.StartSpan(ctx, a.tracer, "evict")
	defer end()

	if err != nil || model.IsNil(a.client) {
		return err
	}

	key := a.toKey(item)

	if err = a.client.Delete(ctx, key); err != nil {
		return fmt.Errorf("evict key `%s` from cache: %w", key, err)
	}

	return nil
}

func (a App[K, V]) Store(ctx context.Context, id K, value V) error {
	return a.store(ctx, a.toKey(id), value)
}

func (a App[K, V]) getValues(ctx context.Context, items []K) []string {
	if model.IsNil(a.client) {
		return nil
	}

	keys := make([]string, len(items))
	for index, id := range items {
		keys[index] = a.toKey(id)
	}

	loadCtx, cancel := context.WithTimeout(ctx, syncActionTimeout)
	defer cancel()

	values, err := a.client.LoadMany(loadCtx, keys...)
	if err != nil {
		loggerWithTrace(ctx, strconv.Itoa(len(keys))).Error("load many from cache: %s", err)
	}

	return values
}

func (a App[k, V]) store(ctx context.Context, key string, item any) error {
	ctx, end := tracer.StartSpan(ctx, a.tracer, "store")
	defer end()

	storeCtx, cancel := context.WithTimeout(ctx, asyncActionTimeout)
	defer cancel()

	payload, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	if err = a.client.Store(storeCtx, key, payload, a.ttl); err != nil {
		return fmt.Errorf("store: %w", err)
	}

	return nil
}

func loggerWithTrace(ctx context.Context, key string) logger.Provider {
	return tracer.AddTraceToLogger(trace.SpanFromContext(ctx), logger.GetGlobal()).WithField("key", key)
}
