package cache

import (
	"context"
	"fmt"

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

func (a App[K, V]) store(ctx context.Context, id K, value V) (err error) {
	ctx, end := tracer.StartSpan(ctx, a.tracer, "store", trace.WithSpanKind(trace.SpanKindInternal))
	defer end(&err)

	payload, err := a.serializer.Encode(value)
	if err != nil {
		return fmt.Errorf("encoding: %w", err)
	}

	if err = a.client.Store(ctx, a.toKey(id), payload, a.ttl); err != nil {
		return fmt.Errorf("store: %w", err)
	}

	return nil
}

func (a App[K, V]) storeMany(ctx context.Context, ids []K, values []V, indexes IndexedItems[K]) error {
	var err error

	ctx, end := tracer.StartSpan(ctx, a.tracer, "store_many", trace.WithSpanKind(trace.SpanKindInternal))
	defer end(&err)

	pipeline := a.client.Pipeline()

	for _, index := range indexes {
		id := ids[index]
		key := a.toKey(id)

		payload, err := a.serializer.Encode(values[index])
		if err != nil {
			loggerWithTrace(ctx, key).Error("encoding: %s", err)

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
