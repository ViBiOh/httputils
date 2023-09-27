package cache

import (
	"context"
	"log/slog"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/cntxt"
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
	done, close := redis.SubscribeFor(ctx, c.write, c.channel, func(id K, err error) {
		slog.Info("evicting from memory cache", "id", id, "channel", c.channel)
		c.memory.Delete(id)
	})

	<-ctx.Done()

	if err := close(cntxt.WithoutDeadline(ctx)); err != nil {
		slog.Error("close subscriber", "err", err)
	}

	<-done
}

func (c *Cache[K, V]) notify(ctx context.Context, id K) {
	if c.memory == nil {
		return
	}

	c.memory.Delete(id)

	if err := c.write.PublishJSON(ctx, c.channel, id); err != nil {
		slog.Error("notify delete on pubsub", "err", err, "id", id)
	}
}
