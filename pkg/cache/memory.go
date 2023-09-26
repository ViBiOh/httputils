package cache

import (
	"context"
	"time"
)

func (c *Cache[K, V]) memoryRead(ctx context.Context, id K) (V, bool) {
	var output V

	if c.memory == nil {
		return output, false
	}

	return c.memory.Get(ctx, id)
}

func (c *Cache[K, V]) memoryWrite(ctx context.Context, id K, value V, ttl time.Duration) {
	if c.memory == nil {
		return
	}

	c.memory.Set(ctx, id, value, ttl)
}
