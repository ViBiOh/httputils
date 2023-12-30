package cache

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ViBiOh/httputils/v4/pkg/telemetry"
	"go.opentelemetry.io/otel/trace"
)

func (c *Cache[K, V]) EvictOnSuccess(ctx context.Context, id K, err error) error {
	if err != nil {
		return err
	}

	if err = c.redisEvict(ctx, id, err); err != nil {
		return err
	}

	if err = c.memoryEvict(ctx, id, err); err != nil {
		return err
	}

	return nil
}

func (c *Cache[K, V]) redisEvict(ctx context.Context, id K, err error) error {
	if c.write == nil {
		return nil
	}

	ctx, end := telemetry.StartSpan(ctx, c.tracer, "evict", trace.WithSpanKind(trace.SpanKindInternal))
	defer end(&err)

	key := c.toKey(id)

	slog.DebugContext(ctx, "evicting from redis cache", "id", id)

	if err = c.write.Delete(ctx, key); err != nil {
		return fmt.Errorf("evict key `%s` from cache: %w", key, err)
	}

	return nil
}

func (c *Cache[K, V]) memoryEvict(ctx context.Context, id K, err error) error {
	if c.memory == nil {
		return nil
	}

	c.memory.Delete(id)

	if c.write == nil {
		return nil
	}

	ctx, end := telemetry.StartSpan(ctx, c.tracer, "evict notify", trace.WithSpanKind(trace.SpanKindInternal))
	defer end(&err)

	if err = c.write.PublishJSON(ctx, c.channel, id); err != nil {
		return fmt.Errorf("evict notify for id %v: %w", id, err)
	}

	return nil
}
