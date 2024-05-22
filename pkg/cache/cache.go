package cache

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/cache/memory"
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
	StoreMany(ctx context.Context, values map[string]any, ttl time.Duration) error
	Delete(ctx context.Context, keys ...string) error
	Expire(ctx context.Context, ttl time.Duration, keys ...string) error
	Pipeline() redis.Pipeliner

	PublishJSON(ctx context.Context, channel string, value any) error
	Subscribe(ctx context.Context, channel string) (<-chan *redis.Message, func(context.Context))
}

type (
	keyer[K comparable]            func(K) string
	fetch[K comparable, V any]     func(context.Context, K) (V, error)
	fetchMany[K comparable, V any] func(context.Context, []K) ([]V, error)
)

type Cache[K comparable, V any] struct {
	serializer  Serializer[V]
	read        RedisClient
	write       RedisClient
	tracer      trace.Tracer
	toKey       keyer[K]
	onMiss      fetch[K, V]
	onMissMany  fetchMany[K, V]
	memory      *memory.Cache[K, V]
	extender    *TTLExtender
	channel     string
	ttl         time.Duration
	concurrency int
}

func New[K comparable, V any](client RedisClient, toKey keyer[K], onMiss fetch[K, V], tracerProvider trace.TracerProvider) *Cache[K, V] {
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

func (c *Cache[K, V]) WithExtendOnHit(ctx context.Context, interval time.Duration, maxSize int) *Cache[K, V] {
	c.extender = NewExtender(c.ttl, interval, maxSize, c.write)

	go c.extender.Start(ctx)

	return c
}

func (c *Cache[K, V]) WithMaxConcurrency(concurrency int) *Cache[K, V] {
	c.concurrency = concurrency

	return c
}

func (c *Cache[K, V]) WithClientSideCaching(ctx context.Context, channel string, size int) *Cache[K, V] {
	c.memory = memory.New[K, V](size)
	c.channel = channel

	go c.subscribe(ctx)
	go c.memory.Start(ctx)

	return c
}

func getClient(client RedisClient) RedisClient {
	if !model.IsNil(client) && client.Enabled() {
		return client
	}

	return nil
}

func (c *Cache[K, V]) Get(ctx context.Context, id K) (V, error) {
	if IsBypassed(ctx) {
		return c.fetch(ctx, id)
	}

	if cached, ok := c.memoryRead(id); ok {
		c.extendTTL(ctx, c.toKey(id))

		return cached, nil
	}

	if c.read == nil {
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
			slog.LogAttrs(ctx, slog.LevelWarn, "load from cache", slog.String("key", key), slog.Any("error", err))
		} else {
			slog.LogAttrs(ctx, slog.LevelError, "load from cache", slog.String("key", key), slog.Any("error", err))
		}
	} else if value, ok, err := c.decode([]byte(content)); err != nil {
		logUnmarshalError(ctx, key, err)
	} else if ok {
		c.memoryWrite(id, value, c.ttl)
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

	if err == nil {
		go doInBackground(context.WithoutCancel(ctx), func(ctx context.Context) error {
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
	if c.write == nil || c.extender == nil || c.ttl == 0 || len(keys) == 0 {
		return
	}

	go doInBackground(context.WithoutCancel(ctx), func(ctx context.Context) error {
		return c.extender.Extend(ctx, keys...)
	})
}

func logUnmarshalError(ctx context.Context, key string, err error) {
	slog.LogAttrs(ctx, slog.LevelError, "unmarshal from cache", slog.String("key", key), slog.Any("error", err))
}
