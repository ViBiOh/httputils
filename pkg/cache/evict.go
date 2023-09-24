package cache

import (
	"context"
	"fmt"

	"github.com/ViBiOh/httputils/v4/pkg/telemetry"
	"go.opentelemetry.io/otel/trace"
)

func (c *Cache[K, V]) EvictOnSuccess(ctx context.Context, id K, err error) error {
	if err != nil || c.write == nil {
		return err
	}

	ctx, end := telemetry.StartSpan(ctx, c.tracer, "evict", trace.WithSpanKind(trace.SpanKindInternal))
	defer end(&err)

	key := c.toKey(id)

	if err = c.write.Delete(ctx, key); err != nil {
		return fmt.Errorf("evict key `%s` from cache: %w", key, err)
	}

	if c.memory != nil {
		if err = c.write.PublishJSON(ctx, c.channel, id); err != nil {
			return fmt.Errorf("pubsub publish for `%v`: %w", id, err)
		}
	}

	return nil
}
