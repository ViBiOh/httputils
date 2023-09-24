package cache

import (
	"context"
	"log/slog"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/cntxt"
	"github.com/ViBiOh/httputils/v4/pkg/redis"
)

type entry[V any] struct {
	expiration time.Time
	content    V
}

func (e entry[V]) isValid(now time.Time) bool {
	return e.expiration.IsZero() || now.Before(e.expiration)
}

func (c *Cache[K, V]) subscribe(ctx context.Context) {
	done, close := redis.SubscribeFor(ctx, c.write, c.channel, func(id K, err error) {
		slog.Info("evicting from memory cache", "id", id, "channel", c.channel)
		c.memoryDelete(ctx, id)
	})

	<-done

	if err := close(cntxt.WithoutDeadline(ctx)); err != nil {
		slog.Error("close subscriber", "err", err)
	}
}

func (c *Cache[K, V]) memoryRead(ctx context.Context, id K, now time.Time) (V, bool) {
	c.mutex.RLock()

	output, ok := c.memory[id]

	c.mutex.RUnlock()

	if ok {
		if output.isValid(now) {
			return output.content, ok
		} else {
			c.memoryDelete(ctx, id)
		}
	}

	return output.content, ok
}

func (c *Cache[K, V]) memorySet(ctx context.Context, id K, value V, expiration time.Time) {
	if c.memory == nil {
		return
	}

	c.mutex.Lock()

	c.memory[id] = entry[V]{
		expiration: expiration,
		content:    value,
	}

	c.mutex.Unlock()
}

func (c *Cache[K, V]) memoryDelete(ctx context.Context, id K) {
	if c.memory == nil {
		return
	}

	c.mutex.Lock()

	delete(c.memory, id)

	c.mutex.Unlock()
}
