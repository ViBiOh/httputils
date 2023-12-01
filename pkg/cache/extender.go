package cache

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/concurrent"
)

type TTLExtender struct {
	redis    RedisClient
	batch    map[string]struct{}
	interval time.Duration
	ttl      time.Duration
	mutex    sync.Mutex
	maxSize  int
}

func NewExtender(ttl, interval time.Duration, maxSize int, redis RedisClient) *TTLExtender {
	extender := &TTLExtender{
		ttl:      ttl,
		interval: interval,
		redis:    redis,
	}

	if interval == 0 {
		return extender
	}

	extender.batch = make(map[string]struct{})
	extender.maxSize = maxSize

	return extender
}

func (te *TTLExtender) Extend(ctx context.Context, keys ...string) error {
	if te.interval == 0 {
		return te.redis.Expire(ctx, te.ttl, keys...)
	}

	te.mutex.Lock()

	for _, key := range keys {
		te.batch[key] = struct{}{}
	}

	if te.maxSize != 0 && len(te.batch) > te.maxSize {
		te.flush(ctx)
	}

	te.mutex.Unlock()

	return nil
}

func (te *TTLExtender) Start(ctx context.Context) {
	if te.interval == 0 {
		return
	}

	ticker := time.NewTicker(te.interval)

	concurrent.ChanUntilDone(ctx, ticker.C, func(_ time.Time) { te.flush(ctx) }, ticker.Stop)
}

func (te *TTLExtender) flush(ctx context.Context) {
	te.mutex.Lock()

	if len(te.batch) == 0 {
		return
	}

	keys := make([]string, 0, len(te.batch))

	for key := range te.batch {
		keys = append(keys, key)
		delete(te.batch, key)
	}

	te.mutex.Unlock()

	if err := te.redis.Expire(ctx, te.ttl, keys...); err != nil {
		slog.ErrorContext(ctx, "extend keys", "err", err)
	}
}
