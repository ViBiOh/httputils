package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/concurrent"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/httputils/v4/pkg/tracer"
	"go.opentelemetry.io/otel/trace"
)

//go:generate mockgen -source cache.go -destination ../mocks/cache.go -package mocks -mock_names RedisClient=RedisClient

var (
	syncActionTimeout  = time.Millisecond * 150
	asyncActionTimeout = time.Second * 5
)

type RedisClient interface {
	Enabled() bool
	Load(ctx context.Context, key string) ([]byte, error)
	LoadMany(ctx context.Context, keys ...string) ([]string, error)
	Store(ctx context.Context, key string, value any, ttl time.Duration) error
	Delete(ctx context.Context, keys ...string) error
}

type App[K any, V any] struct {
	tracer      trace.Tracer
	client      RedisClient
	toKey       func(K) string
	onMiss      func(context.Context, K) (V, error)
	ttl         time.Duration
	concurrency uint64
}

func New[K any, V any](client RedisClient, toKey func(K) string, onMiss func(context.Context, K) (V, error), ttl time.Duration, concurrency uint64, tracer trace.Tracer) App[K, V] {
	return App[K, V]{
		client:      client,
		toKey:       toKey,
		onMiss:      onMiss,
		ttl:         ttl,
		concurrency: concurrency,
		tracer:      tracer,
	}
}

func (a App[K, V]) Get(ctx context.Context, id K) (V, error) {
	if !a.client.Enabled() {
		return a.onMiss(ctx, id)
	}

	ctx, end := tracer.StartSpan(ctx, a.tracer, "get", trace.WithSpanKind(trace.SpanKindInternal))
	defer end()

	key := a.toKey(id)

	loadCtx, cancel := context.WithTimeout(ctx, syncActionTimeout)
	defer cancel()

	if content, err := a.client.Load(loadCtx, key); err != nil {
		loggerWithTrace(ctx, key).Error("load from cache: %s", err)
	} else if value, ok, err := a.unmarshal(ctx, content); err != nil {
		loggerWithTrace(ctx, key).Error("unmarshal from cache: %s", err)
	} else if ok {
		return value, nil
	}

	return a.fetch(ctx, id)
}

// If onMissError returns false, List stops by returning an error
func (a App[K, V]) List(ctx context.Context, onMissError func(K, error) bool, items ...K) ([]V, error) {
	if !a.client.Enabled() {
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

	ctx, end := tracer.StartSpan(ctx, a.tracer, "list", trace.WithSpanKind(trace.SpanKindInternal))
	defer end()

	values := a.getValues(ctx, items)

	if valuesLen := len(values); valuesLen != len(items) {
		return nil, fmt.Errorf("get returned %d values while expecting %d", valuesLen, len(items))
	}

	output := make([]V, len(items))
	wg := concurrent.NewFailFast(a.concurrency)

	for index, item := range items {
		index, item := index, item

		wg.Go(func() error {
			value, ok, err := a.unmarshal(ctx, []byte(values[index]))
			if ok {
				output[index] = value

				return nil
			}

			if err != nil {
				loggerWithTrace(ctx, a.toKey(item)).Error("unmarshal from cache: %s", err)
			}

			value, err = a.fetch(ctx, item)
			if err != nil {
				if !onMissError(item, err) {
					return err
				}

				return nil
			}

			output[index] = value

			return nil
		})
	}

	return output, wg.Wait()
}

// Param fetchMany has to return the same number of values as requested and in the same order
func (a App[K, V]) ListMany(ctx context.Context, fetchMany func(context.Context, []K) ([]V, error), items ...K) ([]V, error) {
	if !a.client.Enabled() {
		return fetchMany(ctx, items)
	}

	ctx, end := tracer.StartSpan(ctx, a.tracer, "list", trace.WithSpanKind(trace.SpanKindInternal))
	defer end()

	values := a.getValues(ctx, items)

	if valuesLen := len(values); valuesLen != len(items) {
		return nil, fmt.Errorf("get returned %d values while expecting %d", valuesLen, len(items))
	}

	var missingIds []K
	var missingIndex []int

	output := make([]V, len(items))
	for index, item := range items {
		value, ok, err := a.unmarshal(ctx, []byte(values[index]))

		if ok {
			output[index] = value

			continue
		}

		if err != nil {
			loggerWithTrace(ctx, a.toKey(item)).Error("unmarshal from cache: %s", err)
		}

		missingIds = append(missingIds, item)
		missingIndex = append(missingIndex, index)
	}

	missingValues, err := fetchMany(ctx, missingIds)
	if err != nil {
		return output, fmt.Errorf("fetch: %w", err)
	}

	if valuesLen := len(missingValues); valuesLen != len(missingIndex) {
		return output, fmt.Errorf("fetch returned %d values while expecting %d", valuesLen, len(missingIndex))
	}

	for index, value := range missingValues {
		output[missingIndex[index]] = value
	}

	return output, nil
}

func (a App[K, V]) EvictOnSuccess(ctx context.Context, item K, err error) error {
	if err != nil || !a.client.Enabled() {
		return err
	}

	ctx, end := tracer.StartSpan(ctx, a.tracer, "evict", trace.WithSpanKind(trace.SpanKindInternal))
	defer end()

	key := a.toKey(item)

	if err = a.client.Delete(ctx, key); err != nil {
		return fmt.Errorf("evict key `%s` from cache: %w", key, err)
	}

	return nil
}

func (a App[K, V]) Store(ctx context.Context, id K, value V) error {
	if !a.client.Enabled() {
		return nil
	}

	return a.store(ctx, id, value)
}

func (a App[K, V]) store(ctx context.Context, id K, value V) error {
	ctx, end := tracer.StartSpan(ctx, a.tracer, "store", trace.WithSpanKind(trace.SpanKindInternal))
	defer end()

	payload, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	storeCtx, cancel := context.WithTimeout(ctx, asyncActionTimeout)
	defer cancel()

	if err = a.client.Store(storeCtx, a.toKey(id), payload, a.ttl); err != nil {
		return fmt.Errorf("store: %w", err)
	}

	return nil
}

func (a App[K, V]) fetch(ctx context.Context, id K) (V, error) {
	ctx, end := tracer.StartSpan(ctx, a.tracer, "fetch", trace.WithSpanKind(trace.SpanKindInternal))
	defer end()

	value, err := a.onMiss(ctx, id)

	if err == nil {
		go func(ctx context.Context) {
			if storeErr := a.store(ctx, id, value); storeErr != nil {
				loggerWithTrace(ctx, a.toKey(id)).Error("store to cache: %s", storeErr)
			}
		}(tracer.CopyToBackground(ctx))
	}

	return value, err
}

func (a App[K, V]) unmarshal(ctx context.Context, content []byte) (V, bool, error) {
	ctx, end := tracer.StartSpan(ctx, a.tracer, "unmarshal", trace.WithSpanKind(trace.SpanKindInternal))
	defer end()

	return unmarshal[V](ctx, content)
}

func unmarshal[V any](ctx context.Context, content []byte) (value V, ok bool, err error) {
	if len(content) == 0 {
		return
	}

	err = json.Unmarshal(content, &value)
	if err == nil {
		ok = true
	}

	return
}

func (a App[K, V]) getValues(ctx context.Context, ids []K) []string {
	keys := make([]string, len(ids))
	for index, id := range ids {
		keys[index] = a.toKey(id)
	}

	loadCtx, cancel := context.WithTimeout(ctx, syncActionTimeout)
	defer cancel()

	values, err := a.client.LoadMany(loadCtx, keys...)
	if err != nil {
		loggerWithTrace(ctx, strconv.Itoa(len(keys))).Error("load many from cache: %s", err)
	}

	return values
}

func loggerWithTrace(ctx context.Context, key string) logger.Provider {
	return tracer.AddTraceToLogger(trace.SpanFromContext(ctx), logger.GetGlobal()).WithField("key", key)
}
