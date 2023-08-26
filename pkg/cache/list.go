package cache

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/ViBiOh/httputils/v4/pkg/cntxt"
	"github.com/ViBiOh/httputils/v4/pkg/concurrent"
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

// If onMissError returns false, List stops by returning an error
func (c *Cache[K, V]) List(ctx context.Context, onMissError func(K, error) bool, items ...K) (outputs []V, err error) {
	if len(items) == 0 {
		return nil, nil
	}

	if c.read == nil || IsBypassed(ctx) {
		if c.onMissMany == nil {
			return c.listRaw(ctx, onMissError, items...)
		}

		return c.listRawMany(ctx, items)
	}

	ctx, end := telemetry.StartSpan(ctx, c.tracer, "list", trace.WithSpanKind(trace.SpanKindInternal))
	defer end(&err)

	keys, values := c.getValues(ctx, items)

	if c.onMissMany == nil {
		return c.handleListSingle(ctx, onMissError, items, keys, values)
	}

	return c.handleListMany(ctx, items, keys, values)
}

func (c *Cache[K, V]) listRawMany(ctx context.Context, items []K) ([]V, error) {
	values, err := c.onMissMany(ctx, items)
	if err != nil {
		return nil, err
	}

	output := make([]V, len(values))
	for index, item := range items {
		output[index] = values[item]
	}

	return output, nil
}

func (c *Cache[K, V]) listRaw(ctx context.Context, onMissError func(K, error) bool, items ...K) ([]V, error) {
	output := make([]V, len(items))

	for index, item := range items {
		value, err := c.fetch(ctx, item)
		if err != nil {
			if !onMissError(item, err) {
				return nil, err
			}

			continue
		}

		output[index] = value
	}

	return output, nil
}

func (c *Cache[K, V]) handleListSingle(ctx context.Context, onMissError func(K, error) bool, items []K, keys, values []string) ([]V, error) {
	output := make([]V, len(items))
	wg := concurrent.NewFailFast(c.concurrency)
	ctx = wg.WithContext(ctx)

	var extendKeys []string

	for index, item := range items {
		index, item := index, item

		wg.Go(func() error {
			value, ok, err := c.decode([]byte(values[index]))
			if ok {
				output[index] = value

				if c.ttl != 0 && c.extendOnHit {
					extendKeys = append(extendKeys, keys[index])
				}

				return nil
			}

			if err != nil {
				logUnmarshalError(ctx, c.toKey(item), err)
			}

			if output[index], err = c.fetch(ctx, item); err != nil && !onMissError(item, err) {
				return err
			}

			return nil
		})
	}

	c.extendTTL(ctx, extendKeys...)

	return output, wg.Wait()
}

// Param fetchMany has to return the same number of values as requested and in the same order
func (c *Cache[K, V]) handleListMany(ctx context.Context, items []K, keys, values []string) ([]V, error) {
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
