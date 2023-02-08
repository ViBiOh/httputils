package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/tracer"
	"go.opentelemetry.io/otel/trace"
)

func (a App[K, V]) Store(ctx context.Context, id K, value V) error {
	if !a.client.Enabled() {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, syncActionTimeout)
	defer cancel()

	return a.store(ctx, id, value)
}

func (a App[K, V]) store(ctx context.Context, id K, value V) error {
	ctx, end := tracer.StartSpan(ctx, a.tracer, "store", trace.WithSpanKind(trace.SpanKindInternal))
	defer end()

	return store(ctx, a.client, a.toKey(id), value, a.ttl)
}

func store(ctx context.Context, client RedisClient, key string, value any, ttl time.Duration) error {
	payload, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	if err = client.Store(ctx, key, payload, ttl); err != nil {
		return fmt.Errorf("store: %w", err)
	}

	return nil
}

func (a App[K, V]) storeMany(ctx context.Context, ids []K, values []V, indexes []int) error {
	ctx, end := tracer.StartSpan(ctx, a.tracer, "store_many", trace.WithSpanKind(trace.SpanKindInternal))
	defer end()

	pipeline := a.client.Pipeline()

	for _, index := range indexes {
		id := ids[index]
		key := a.toKey(id)

		payload, err := json.Marshal(values[index])
		if err != nil {
			loggerWithTrace(ctx, key).Error("marshal: %s", err)

			continue
		}

		if err := pipeline.Set(ctx, key, payload, a.ttl).Err(); err != nil {
			loggerWithTrace(ctx, key).Error("pipeline set: %s", err)

			continue
		}
	}

	if _, err := pipeline.Exec(ctx); err != nil {
		return fmt.Errorf("pipeline exec: %s", err)
	}

	return nil
}