package cache

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/telemetry"
	"go.opentelemetry.io/otel/trace"
)

func (c *Cache[K, V]) Store(ctx context.Context, id K, value V) error {
	ctx, cancel := context.WithTimeout(ctx, syncActionTimeout)
	defer cancel()

	return c.store(ctx, id, value)
}

func (c *Cache[K, V]) store(ctx context.Context, id K, value V) error {
	c.memoryWrite(id, value, c.ttl)

	if err := c.redisWrite(ctx, id, value, c.ttl); err != nil {
		return err
	}

	return nil
}

func (c *Cache[K, V]) storeMany(ctx context.Context, ids []K, values []V, indexedIDs IndexedIDs[K]) error {
	if c.write == nil && c.memory == nil {
		return nil
	}

	var err error

	ctx, end := telemetry.StartSpan(ctx, c.tracer, "store_many", trace.WithSpanKind(trace.SpanKindInternal))
	defer end(&err)

	toSet := make(map[string]any)

	for _, indexed := range indexedIDs {
		id := indexed.id
		index := indexed.index

		c.memoryWrite(id, values[index], c.ttl)

		if c.write == nil {
			continue
		}

		key := c.toKey(id)

		payload, err := c.serializer.Encode(values[index])
		if err != nil {
			slog.ErrorContext(ctx, "encoding", "error", err, "key", key)

			continue
		}

		toSet[key] = payload
	}

	if c.write != nil {
		return c.write.StoreMany(ctx, toSet, c.ttl)
	}

	return nil
}

func (c *Cache[K, V]) redisWrite(ctx context.Context, id K, value V, ttl time.Duration) (err error) {
	if c.write == nil {
		return nil
	}

	ctx, end := telemetry.StartSpan(ctx, c.tracer, "store", trace.WithSpanKind(trace.SpanKindInternal))
	defer end(&err)

	payload, err := c.serializer.Encode(value)
	if err != nil {
		return fmt.Errorf("encoding: %w", err)
	}

	if err = c.write.Store(ctx, c.toKey(id), payload, c.ttl); err != nil {
		return fmt.Errorf("store: %w", err)
	}

	return nil
}
