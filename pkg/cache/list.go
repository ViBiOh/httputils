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
func (a *Cache[K, V]) List(ctx context.Context, onMissError func(K, error) bool, items ...K) (outputs []V, err error) {
	if len(items) == 0 {
		return nil, nil
	}

	if a.read == nil || IsBypassed(ctx) {
		if a.onMissMany == nil {
			return a.listRaw(ctx, onMissError, items...)
		}

		return a.listRawMany(ctx, items)
	}

	ctx, end := telemetry.StartSpan(ctx, a.tracer, "list", trace.WithSpanKind(trace.SpanKindInternal))
	defer end(&err)

	keys, values := a.getValues(ctx, items)

	if a.onMissMany == nil {
		return a.handleListSingle(ctx, onMissError, items, keys, values)
	}

	return a.handleListMany(ctx, items, keys, values)
}

func (a *Cache[K, V]) listRawMany(ctx context.Context, items []K) ([]V, error) {
	values, err := a.onMissMany(ctx, items)
	if err != nil {
		return nil, err
	}

	output := make([]V, len(values))
	for index, item := range items {
		output[index] = values[item]
	}

	return output, nil
}

func (a *Cache[K, V]) listRaw(ctx context.Context, onMissError func(K, error) bool, items ...K) ([]V, error) {
	output := make([]V, len(items))

	for index, item := range items {
		value, err := a.fetch(ctx, item)
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

func (a *Cache[K, V]) handleListSingle(ctx context.Context, onMissError func(K, error) bool, items []K, keys, values []string) ([]V, error) {
	output := make([]V, len(items))
	wg := concurrent.NewFailFast(a.concurrency)
	ctx = wg.WithContext(ctx)

	var extendKeys []string

	for index, item := range items {
		index, item := index, item

		wg.Go(func() error {
			value, ok, err := a.decode([]byte(values[index]))
			if ok {
				output[index] = value

				if a.ttl != 0 && a.extendOnHit {
					extendKeys = append(extendKeys, keys[index])
				}

				return nil
			}

			if err != nil {
				logUnmarshalError(ctx, a.toKey(item), err)
			}

			if output[index], err = a.fetch(ctx, item); err != nil && !onMissError(item, err) {
				return err
			}

			return nil
		})
	}

	a.extendTTL(ctx, extendKeys...)

	return output, wg.Wait()
}

// Param fetchMany has to return the same number of values as requested and in the same order
func (a *Cache[K, V]) handleListMany(ctx context.Context, items []K, keys, values []string) ([]V, error) {
	var extendKeys []string

	missingKeys := make(IndexedItems[K])
	output := make([]V, len(items))

	for index, item := range items {
		if value, ok, err := a.decode([]byte(values[index])); ok {
			output[index] = value

			if a.ttl != 0 && a.extendOnHit {
				extendKeys = append(extendKeys, keys[index])
			}

			continue
		} else if err != nil {
			logUnmarshalError(ctx, a.toKey(item), err)
		}

		missingKeys[item] = index
	}

	a.extendTTL(ctx, extendKeys...)

	missingValues, err := a.onMissMany(ctx, missingKeys.Items())
	if err != nil {
		return output, fmt.Errorf("fetch many: %w", err)
	}

	for key, value := range missingValues {
		output[missingKeys[key]] = value
	}

	go doInBackground(cntxt.WithoutDeadline(ctx), func(ctx context.Context) error {
		return a.storeMany(ctx, items, output, missingKeys)
	})

	return output, nil
}

func (a *Cache[K, V]) getValues(ctx context.Context, ids []K) ([]string, []string) {
	keys := make([]string, len(ids))
	for index, id := range ids {
		keys[index] = a.toKey(id)
	}

	loadCtx, cancel := context.WithTimeout(ctx, syncActionTimeout)
	defer cancel()

	values, err := a.read.LoadMany(loadCtx, keys...)
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
