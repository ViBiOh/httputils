package cache

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/cntxt"
	"github.com/ViBiOh/httputils/v4/pkg/model"
	"github.com/ViBiOh/httputils/v4/pkg/telemetry"
	"github.com/redis/go-redis/v9"
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
	fetchMany[K comparable, V any] func(context.Context, []K) (map[K]V, error)
)

type Cache[K comparable, V any] struct {
	tracer      trace.Tracer
	read        RedisClient
	write       RedisClient
	toKey       func(K) string
	serializer  Serializer[V]
	onMiss      fetch[K, V]
	onMissMany  fetchMany[K, V]
	ttl         time.Duration
	concurrency int
	extendOnHit bool
}

func New[K comparable, V any](client RedisClient, toKey func(K) string, onMiss fetch[K, V], tracerProvider trace.TracerProvider) *Cache[K, V] {
	client = getClient(client)

	cache := &Cache[K, V]{
		read:       client,
		write:      client,
		toKey:      toKey,
		serializer: JSONSerializer[V]{},
		onMiss:     onMiss,
	}

	if tracerProvider != nil {
		cache.tracer = tracerProvider.Tracer("cache")
	}

	return cache
}

func (c *Cache[K, V]) WithMissMany(cb fetchMany[K, V]) *Cache[K, V] {
	c.onMissMany = cb

	return c
}

func (c *Cache[K, V]) WithSerializer(serializer Serializer[V]) *Cache[K, V] {
	c.serializer = serializer

	return c
}

func (c *Cache[K, V]) WithRead(client RedisClient) *Cache[K, V] {
	c.read = getClient(client)

	return c
}

func (c *Cache[K, V]) WithTTL(ttl time.Duration) *Cache[K, V] {
	c.ttl = ttl

	return c
}

func (c *Cache[K, V]) WithExtendOnHit() *Cache[K, V] {
	c.extendOnHit = true

	return c
}

func (c *Cache[K, V]) WithMaxConcurrency(concurrency int) *Cache[K, V] {
	c.concurrency = concurrency

	return c
}

func getClient(client RedisClient) RedisClient {
	if !model.IsNil(client) && client.Enabled() {
		return client
	}

	return nil
}

func (c *Cache[K, V]) Get(ctx context.Context, id K) (V, error) {
	if c.read == nil || IsBypassed(ctx) {
		return c.fetch(ctx, id)
	}

	var err error

	ctx, end := telemetry.StartSpan(ctx, c.tracer, "get", trace.WithSpanKind(trace.SpanKindInternal))
	defer end(&err)

	key := c.toKey(id)

	loadCtx, cancel := context.WithTimeout(ctx, syncActionTimeout)
	defer cancel()

	if content, err := c.read.Load(loadCtx, key); err != nil {
		if errors.Is(err, context.Canceled) {
			loggerWithTrace(ctx, key).Warn("load from cache", "err", err)
		} else {
			loggerWithTrace(ctx, key).Error("load from cache", "err", err)
		}
	} else if value, ok, err := c.decode([]byte(content)); err != nil {
		logUnmarshalError(ctx, key, err)
	} else if ok {
		c.extendTTL(ctx, key)

		return value, nil
	}

	return c.fetch(ctx, id)
}

func (c *Cache[K, V]) fetch(ctx context.Context, id K) (V, error) {
	var err error

	ctx, end := telemetry.StartSpan(ctx, c.tracer, "fetch", trace.WithSpanKind(trace.SpanKindInternal))
	defer end(&err)

	value, err := c.onMiss(ctx, id)

	if err == nil && c.write != nil {
		go doInBackground(cntxt.WithoutDeadline(ctx), func(ctx context.Context) error {
			return c.store(ctx, id, value)
		})
	}

	return value, err
}

func (c *Cache[K, V]) decode(content []byte) (value V, ok bool, err error) {
	if len(content) == 0 {
		return
	}

	value, err = c.serializer.Decode(content)
	ok = err == nil

	return
}

func (c *Cache[K, V]) extendTTL(ctx context.Context, keys ...string) {
	if c.write == nil || !c.extendOnHit || c.ttl == 0 || len(keys) == 0 {
		return
	}

	go doInBackground(cntxt.WithoutDeadline(ctx), func(ctx context.Context) error {
		return c.write.Expire(ctx, c.ttl, keys...)
	})
}

func logUnmarshalError(ctx context.Context, key string, err error) {
	loggerWithTrace(ctx, key).Error("unmarshal from cache", "err", err)
}

func loggerWithTrace(ctx context.Context, key string) *slog.Logger {
	return telemetry.AddTraceToLogger(trace.SpanFromContext(ctx), slog.Default()).With("key", key)
}
