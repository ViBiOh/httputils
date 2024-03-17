package memory

import (
	"context"

	"github.com/ViBiOh/httputils/v4/pkg/concurrent"
)

type LRUAction int

const (
	Touch LRUAction = iota
	Add   LRUAction = iota
)

type LeastRecentlyUsedAction[K comparable] struct {
	id     K
	action LRUAction
}

func (c *Cache[K, V]) startLRU(ctx context.Context) {
	defer close(c.lruUpdates)

	if c.maxSize == 0 {
		return
	}

	concurrent.ChanUntilDone(ctx, c.lruUpdates, func(update LeastRecentlyUsedAction[K]) {
		switch update.action {
		case Touch:
			c.refreshEntryLRU(update.id)
		case Add:
			c.addEntryLRU(update.id)
		}
	}, c.close)
}

func (c *Cache[K, V]) touchLRU(id K) {
	if c.maxSize == 0 {
		return
	}

	c.sendLRUAction(LeastRecentlyUsedAction[K]{id: id, action: Touch})
}

func (c *Cache[K, V]) addLRU(id K) {
	if c.maxSize == 0 {
		return
	}

	c.sendLRUAction(LeastRecentlyUsedAction[K]{id: id, action: Add})
}

func (c *Cache[K, V]) refreshEntryLRU(id K) {
	for element := c.lru.Front(); element != nil; element = element.Next() {
		if element.Value != id {
			continue
		}

		c.lru.MoveToFront(element)

		return
	}
}

func (c *Cache[K, V]) addEntryLRU(id K) {
	if c.lru.Len() == c.maxSize {
		back := c.lru.Back()
		c.lru.Remove(back)

		c.Delete(back.Value.(K))
	}

	c.lru.PushFront(id)
}

func (c *Cache[K, V]) sendLRUAction(action LeastRecentlyUsedAction[K]) {
	select {
	case <-c.done:
		return
	default:
	}

	select {
	case <-c.done:
	case c.lruUpdates <- action:
	default:
	}
}
