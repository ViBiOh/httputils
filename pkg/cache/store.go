package cache

import (
	"context"
	"fmt"

	"github.com/ViBiOh/httputils/v4/pkg/telemetry"
	"go.opentelemetry.io/otel/trace"
)

func (c *Cache[K, V]) Store(ctx context.Context, id K, value V) error {
	if c.write == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, syncActionTimeout)
	defer cancel()

	return c.store(ctx, id, value)
}

func (c *Cache[K, V]) store(ctx context.Context, id K, value V) (err error) {
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

func (c *Cache[K, V]) storeMany(ctx context.Context, ids []K, values []V, indexes IndexedItems[K]) error {
	var err error

	ctx, end := telemetry.StartSpan(ctx, c.tracer, "store_many", trace.WithSpanKind(trace.SpanKindInternal))
	defer end(&err)

	toSet := make(map[string]any)

	for _, index := range indexes {
		id := ids[index]
		key := c.toKey(id)

		payload, err := c.serializer.Encode(values[index])
		if err != nil {
			loggerWithTrace(ctx, key).Error("encoding", "err", err)

			continue
		}

		toSet[key] = payload
	}

	return c.write.StoreMany(ctx, toSet, c.ttl)
}
