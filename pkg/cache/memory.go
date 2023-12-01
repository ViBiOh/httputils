package cache

import (
	"context"
	"log/slog"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/redis"
)

func (c *Cache[K, V]) memoryRead(id K) (V, bool) {
	var output V

	if c.memory == nil {
		return output, false
	}

	return c.memory.Get(id)
}

func (c *Cache[K, V]) memoryWrite(id K, value V, ttl time.Duration) {
	if c.memory == nil {
		return
	}

	c.memory.Set(id, value, ttl)
}

func (c *Cache[K, V]) subscribe(ctx context.Context) {
	if c.read == nil {
		return
	}

	redis.SubscribeFor(ctx, c.read, c.channel, func(id K, err error) {
		slog.InfoContext(ctx, "evicting from memory cache", "id", id, "channel", c.channel)
		c.memory.Delete(id)
	})
}
