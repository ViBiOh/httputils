package memory

import (
	"context"
)

func (c *Cache[K, V]) startLRU(ctx context.Context) {
	if c.maxSize == 0 {
		return
	}

	for {
		select {
		case <-ctx.Done():
			return

		case id := <-c.touches:
			c.refresh(id)
		}
	}
}

func (c *Cache[K, V]) touch(id K) {
	if c.maxSize == 0 {
		return
	}

	select {
	case c.touches <- id:
	default:
	}
}

func (c *Cache[K, V]) refresh(id K) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if item, ok := c.content[id]; ok && item.element != nil {
		c.lru.MoveToFront(item.element)
	}
}

func (c *Cache[K, V]) markUsed(item *entry[K, V]) {
	if c.maxSize == 0 {
		return
	}

	if item.element == nil {
		item.element = c.lru.PushFront(item)
	} else {
		c.lru.MoveToFront(item.element)
	}

	for c.lru.Len() > c.maxSize {
		c.remove(c.lru.Back().Value.(*entry[K, V]))
	}
}
