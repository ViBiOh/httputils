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
}

func NewExtender(ttl, interval time.Duration, redis RedisClient) *TTLExtender {
	extender := &TTLExtender{
		ttl:      ttl,
		interval: interval,
		redis:    redis,
	}

	if interval == 0 {
		return extender
	}

	extender.batch = make(map[string]struct{})

	return extender
}

func (be *TTLExtender) Extend(ctx context.Context, keys ...string) error {
	if be.interval == 0 {
		return be.redis.Expire(ctx, be.ttl, keys...)
	}

	be.mutex.Lock()
	for _, key := range keys {
		be.batch[key] = struct{}{}
	}
	be.mutex.Unlock()

	return nil
}

func (be *TTLExtender) Start(ctx context.Context) {
	if be.interval == 0 {
		return
	}

	ticker := time.NewTicker(be.interval)

	concurrent.ChanUntilDone(ctx, ticker.C, func(_ time.Time) {
		be.mutex.Lock()

		keys := make([]string, 0, len(be.batch))

		for key := range be.batch {
			keys = append(keys, key)
			delete(be.batch, key)
		}

		be.mutex.Unlock()

		if err := be.redis.Expire(ctx, be.ttl, keys...); err != nil {
			slog.Error("extend keys", "err", err)
		}
	}, ticker.Stop)
}
