package memory

import (
	"container/heap"
	"context"
	"time"
)

const idleDelay = time.Hour * 24

func (c *Cache[K, V]) startEvicter(ctx context.Context) {
	timer := time.NewTimer(idleDelay)
	defer timer.Stop()

	for {
		timer.Reset(c.evictExpired())

		select {
		case <-ctx.Done():
			return
		case <-timer.C:
		case <-c.wake:
		}
	}
}

func (c *Cache[K, V]) evictExpired() time.Duration {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for len(c.expiration) > 0 {
		first := c.expiration[0]

		if remaining := time.Until(first.expiration); remaining > 0 {
			return remaining
		}

		c.remove(first)
	}

	return idleDelay
}

func (c *Cache[K, V]) setExpiration(item *entry[K, V], ttl time.Duration) {
	if ttl == 0 {
		item.expiration = time.Time{}

		if item.index != -1 {
			heap.Remove(&c.expiration, item.index)
		}

		return
	}

	item.expiration = time.Now().Add(ttl)

	if item.index == -1 {
		heap.Push(&c.expiration, item)
	} else {
		heap.Fix(&c.expiration, item.index)
	}

	if c.expiration[0] == item {
		select {
		case c.wake <- struct{}{}:
		default:
		}
	}
}
