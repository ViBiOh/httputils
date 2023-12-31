package cache

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/ViBiOh/httputils/v4/pkg/cntxt"
	"github.com/ViBiOh/httputils/v4/pkg/concurrent"
	"github.com/ViBiOh/httputils/v4/pkg/telemetry"
	"go.opentelemetry.io/otel/trace"
)

type FetchErrHandler[K comparable] func(context.Context, K, error) error

type IndexedID[K comparable] struct {
	id    K
	index int
}

type IndexedIDs[K comparable] []IndexedID[K]

func (ii IndexedIDs[K]) IDs() []K {
	output := make([]K, len(ii))

	for index, indexed := range ii {
		output[index] = indexed.id
	}

	return output
}

func (c *Cache[K, V]) List(ctx context.Context, onFetchErr FetchErrHandler[K], ids ...K) (outputs []V, err error) {
	if len(ids) == 0 {
		return nil, nil
	}

	if IsBypassed(ctx) {
		return c.fetchAll(ctx, onFetchErr, ids)
	}

	ctx, end := telemetry.StartSpan(ctx, c.tracer, "list", trace.WithSpanKind(trace.SpanKindInternal))
	defer end(&err)

	output, remainingIDs := c.memoryValues(ids)
	keys, values := c.redisValues(ctx, remainingIDs)

	return c.handleList(ctx, onFetchErr, ids, output, remainingIDs, keys, values)
}

func (c *Cache[K, V]) fetchAll(ctx context.Context, onFetchErr FetchErrHandler[K], ids []K) ([]V, error) {
	if c.onMissMany != nil {
		return c.onMissMany(ctx, ids)
	}

	wg := concurrent.NewFailFast(c.concurrency)

	output := make([]V, len(ids))

	for index, id := range ids {
		index := index
		id := id

		wg.Go(func() error {
			value, err := c.fetch(ctx, id)
			if err != nil {
				if onFetchErr != nil {
					return onFetchErr(ctx, id, err)
				}

				slog.ErrorContext(ctx, "fetch id", "error", err, "id", id)
			} else {
				output[index] = value
			}

			return nil
		})
	}

	return output, wg.Wait()
}

func (c *Cache[K, V]) handleList(ctx context.Context, onFetchErr FetchErrHandler[K], ids []K, output []V, remainings []K, keys, values []string) ([]V, error) {
	var extendKeys []string
	var missingIDs IndexedIDs[K]

	remainingsLength, remainingsPos := len(remainings), 0

	for index, id := range ids {
		if remainingsPos >= remainingsLength || remainings[remainingsPos] != id {
			if c.ttl != 0 && c.extender != nil {
				extendKeys = append(extendKeys, c.toKey(id))
			}

			continue
		}

		if value, ok, err := c.decode([]byte(values[remainingsPos])); ok {
			output[index] = value

			if c.ttl != 0 && c.extender != nil {
				extendKeys = append(extendKeys, keys[remainingsPos])
			}

			c.memoryWrite(id, value, c.ttl)

			remainingsPos++

			continue
		} else if err != nil {
			logUnmarshalError(ctx, c.toKey(id), err)
		}

		remainingsPos++

		missingIDs = append(missingIDs, IndexedID[K]{id: id, index: index})
	}

	c.extendTTL(ctx, extendKeys...)

	if len(missingIDs) == 0 {
		return output, nil
	}

	missingValues, err := c.fetchAll(ctx, onFetchErr, missingIDs.IDs())
	if err != nil {
		return output, fmt.Errorf("fetch many: %w", err)
	}

	for index, missing := range missingIDs {
		output[missing.index] = missingValues[index]
	}

	go doInBackground(cntxt.WithoutDeadline(ctx), func(ctx context.Context) error {
		return c.storeMany(ctx, ids, output, missingIDs)
	})

	return output, nil
}

func (c *Cache[K, V]) memoryValues(ids []K) ([]V, []K) {
	output := make([]V, len(ids))

	if c.memory == nil {
		return output, ids
	}

	return output, c.memory.GetAll(ids, output)
}

func (c *Cache[K, V]) redisValues(ctx context.Context, ids []K) ([]string, []string) {
	if len(ids) == 0 {
		return nil, nil
	}

	keys := make([]string, len(ids))
	for index, id := range ids {
		keys[index] = c.toKey(id)
	}

	if c.read == nil {
		return keys, make([]string, len(ids))
	}

	loadCtx, cancel := context.WithTimeout(ctx, syncActionTimeout)
	defer cancel()

	values, err := c.read.LoadMany(loadCtx, keys...)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			slog.WarnContext(ctx, "load many from cache", "error", err, "key", strconv.Itoa(len(keys)))
		} else {
			slog.ErrorContext(ctx, "load many from cache", "error", err, "key", strconv.Itoa(len(keys)))
		}

		values = make([]string, len(ids))
	}

	return keys, values
}
