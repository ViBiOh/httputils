package memory

import (
	"container/heap"
	"context"
	"sync"
	"time"
)

type Cache[K comparable, V any] struct {
	content    map[K]V
	expiration *ExpirationQueue[K]
	mutex      sync.RWMutex
}

func New[K comparable, V any]() *Cache[K, V] {
	return &Cache[K, V]{
		content:    make(map[K]V),
		expiration: &ExpirationQueue[K]{},
	}
}

func (c *Cache[K, V]) Get(ctx context.Context, id K) (V, bool) {
	c.mutex.RLock()

	output, ok := c.content[id]

	c.mutex.RUnlock()

	return output, ok
}

func (c *Cache[K, V]) Set(ctx context.Context, id K, value V, ttl time.Duration) {
	if c.content == nil {
		return
	}

	c.mutex.Lock()

	c.content[id] = value

	if ttl != 0 {
		heap.Push(c.expiration, Item[K]{id: id, expiration: time.Now().Add(ttl)})
	}

	c.mutex.Unlock()
}

func (c *Cache[K, V]) Delete(ctx context.Context, id K) {
	if c.content == nil {
		return
	}

	c.mutex.Lock()

	delete(c.content, id)

	c.mutex.Unlock()
}
