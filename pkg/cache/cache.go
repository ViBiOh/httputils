package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/httputils/v4/pkg/tracer"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

//go:generate mockgen -source cache.go -destination ../mocks/cache.go -package mocks -mock_names RedisClient=RedisClient

var syncActionTimeout = time.Millisecond * 150

type RedisClient interface {
	Enabled() bool
	Load(ctx context.Context, key string) ([]byte, error)
	LoadMany(ctx context.Context, keys ...string) ([]string, error)
	Store(ctx context.Context, key string, value any, ttl time.Duration) error
	Delete(ctx context.Context, keys ...string) error
	Expire(ctx context.Context, ttl time.Duration, keys ...string) error
	Pipeline() redis.Pipeliner
}

type (
	fetch[K comparable, V any]     func(context.Context, K) (V, error)
	fetchMany[K comparable, V any] func(context.Context, []K) ([]V, error)
)

type App[K comparable, V any] struct {
	tracer      trace.Tracer
	client      RedisClient
	toKey       func(K) string
	onMiss      fetch[K, V]
	onMissMany  fetchMany[K, V]
	ttl         time.Duration
	concurrency uint64
}

func New[K comparable, V any](client RedisClient, toKey func(K) string, onMiss fetch[K, V], ttl time.Duration, concurrency uint64, tracer trace.Tracer) App[K, V] {
	return App[K, V]{
		client:      client,
		toKey:       toKey,
		onMiss:      onMiss,
		ttl:         ttl,
		concurrency: concurrency,
		tracer:      tracer,
	}
}

func (a App[K, V]) WithMissMany(cb fetchMany[K, V]) App[K, V] {
	a.onMissMany = cb

	return a
}

func (a App[K, V]) Get(ctx context.Context, id K) (V, error) {
	if !a.client.Enabled() {
		return a.onMiss(ctx, id)
	}

	ctx, end := tracer.StartSpan(ctx, a.tracer, "get", trace.WithSpanKind(trace.SpanKindInternal))
	defer end()

	key := a.toKey(id)

	loadCtx, cancel := context.WithTimeout(ctx, syncActionTimeout)
	defer cancel()

	if content, err := a.client.Load(loadCtx, key); err != nil {
		if errors.Is(err, context.Canceled) {
			loggerWithTrace(ctx, key).Warn("load from cache: %s", err)
		} else {
			loggerWithTrace(ctx, key).Error("load from cache: %s", err)
		}
	} else if value, ok, err := a.unmarshal(ctx, content); err != nil {
		logUnmarshallError(ctx, key, err)
	} else if ok {
		a.extendTTL(ctx, key)

		return value, nil
	}

	return a.fetch(ctx, id)
}

func (a App[K, V]) EvictOnSuccess(ctx context.Context, item K, err error) error {
	if err != nil || !a.client.Enabled() {
		return err
	}

	ctx, end := tracer.StartSpan(ctx, a.tracer, "evict", trace.WithSpanKind(trace.SpanKindInternal))
	defer end()

	key := a.toKey(item)

	if err = a.client.Delete(ctx, key); err != nil {
		return fmt.Errorf("evict key `%s` from cache: %w", key, err)
	}

	return nil
}

func (a App[K, V]) fetch(ctx context.Context, id K) (V, error) {
	ctx, end := tracer.StartSpan(ctx, a.tracer, "fetch", trace.WithSpanKind(trace.SpanKindInternal))
	defer end()

	value, err := a.onMiss(ctx, id)

	if err == nil {
		go doInBackground(tracer.CopyToBackground(ctx), "store to cache", func(ctx context.Context) error {
			return a.store(ctx, id, value)
		})
	}

	return value, err
}

func (a App[K, V]) unmarshal(ctx context.Context, content []byte) (V, bool, error) {
	ctx, end := tracer.StartSpan(ctx, a.tracer, "unmarshal", trace.WithAttributes(attribute.Int("len", len(content))), trace.WithSpanKind(trace.SpanKindInternal))
	defer end()

	return unmarshal[V](ctx, content)
}

func unmarshal[V any](ctx context.Context, content []byte) (value V, ok bool, err error) {
	if len(content) == 0 {
		return
	}

	err = json.Unmarshal(content, &value)
	if err == nil {
		ok = true
	}

	return
}

func (a App[K, V]) extendTTL(ctx context.Context, keys ...string) {
	go doInBackground(tracer.CopyToBackground(ctx), "extend ttl", func(ctx context.Context) error {
		return a.client.Expire(ctx, a.ttl, keys...)
	})
}

func loggerWithTrace(ctx context.Context, key string) logger.Provider {
	return tracer.AddTraceToLogger(trace.SpanFromContext(ctx), logger.GetGlobal()).WithField("key", key)
}

func logUnmarshallError(ctx context.Context, key string, err error) {
	loggerWithTrace(ctx, key).Error("unmarshal from cache: %s", err)
}
