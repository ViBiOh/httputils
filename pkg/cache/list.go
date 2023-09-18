package cache

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/ViBiOh/httputils/v4/pkg/cntxt"
	"github.com/ViBiOh/httputils/v4/pkg/telemetry"
	"go.opentelemetry.io/otel/trace"
)

type IndexedItems[K comparable] map[K]int

func (ii IndexedItems[K]) Items() []K {
	output := make([]K, len(ii))
	index := 0

	for item := range ii {
		output[index] = item
		index++
	}

	return output
}

func (c *Cache[K, V]) List(ctx context.Context, items ...K) (outputs []V, err error) {
	if len(items) == 0 {
		return nil, nil
	}

	if c.read == nil || IsBypassed(ctx) {
		return c.listRaw(ctx, items)
	}

	ctx, end := telemetry.StartSpan(ctx, c.tracer, "list", trace.WithSpanKind(trace.SpanKindInternal))
	defer end(&err)

	keys, values := c.getValues(ctx, items)

	return c.handleList(ctx, items, keys, values)
}

func (c *Cache[K, V]) listRaw(ctx context.Context, items []K) ([]V, error) {
	values, err := c.fetchAll(ctx, items)
	if err != nil {
		return nil, err
	}

	output := make([]V, len(values))
	for index, item := range items {
		output[index] = values[item]
	}

	return output, nil
}

func (c *Cache[K, V]) handleList(ctx context.Context, items []K, keys, values []string) ([]V, error) {
	var extendKeys []string

	missingKeys := make(IndexedItems[K])
	output := make([]V, len(items))

	for index, item := range items {
		if value, ok, err := c.decode([]byte(values[index])); ok {
			output[index] = value

			if c.ttl != 0 && c.extendOnHit {
				extendKeys = append(extendKeys, keys[index])
			}

			continue
		} else if err != nil {
			logUnmarshalError(ctx, c.toKey(item), err)
		}

		missingKeys[item] = index
	}

	c.extendTTL(ctx, extendKeys...)

	missingValues, err := c.onMissMany(ctx, missingKeys.Items())
	if err != nil {
		return output, fmt.Errorf("fetch many: %w", err)
	}

	for key, value := range missingValues {
		output[missingKeys[key]] = value
	}

	go doInBackground(cntxt.WithoutDeadline(ctx), func(ctx context.Context) error {
		return c.storeMany(ctx, items, output, missingKeys)
	})

	return output, nil
}

func (c *Cache[K, V]) getValues(ctx context.Context, ids []K) ([]string, []string) {
	keys := make([]string, len(ids))
	for index, id := range ids {
		keys[index] = c.toKey(id)
	}

	loadCtx, cancel := context.WithTimeout(ctx, syncActionTimeout)
	defer cancel()

	values, err := c.read.LoadMany(loadCtx, keys...)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			loggerWithTrace(ctx, strconv.Itoa(len(keys))).Warn("load many from cache", "err", err)
		} else {
			loggerWithTrace(ctx, strconv.Itoa(len(keys))).Error("load many from cache", "err", err)
		}

		values = make([]string, len(ids))
	}

	return keys, values
}
