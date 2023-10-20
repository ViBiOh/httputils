package memory

import (
	"container/list"
	"context"
	"runtime"
	"sync"
	"time"
)

type Cache[K comparable, V any] struct {
	content           map[K]V
	expiration        *ExpirationQueue[K]
	lru               *list.List
	expirationUpdates chan ExpirationQueueAction[K]
	lruUpdates        chan LeastRecentlyUsedAction[K]
	mutex             sync.RWMutex
	maxSize           int
}

func New[K comparable, V any](maxSize int) *Cache[K, V] {
	return &Cache[K, V]{
		content:           make(map[K]V),
		expiration:        &ExpirationQueue[K]{},
		expirationUpdates: make(chan ExpirationQueueAction[K], runtime.NumCPU()),
		lruUpdates:        make(chan LeastRecentlyUsedAction[K], runtime.NumCPU()*10), // read ratio 10:1
		lru:               list.New(),
		maxSize:           maxSize,
	}
}

func (c *Cache[K, V]) Start(ctx context.Context) {
	go c.startLRU(ctx)

	c.startEvicter(ctx.Done())
}

func (c *Cache[K, V]) Get(id K) (V, bool) {
	c.mutex.RLock()

	output, ok := c.content[id]
	if ok {
		c.touchLRU(id)
	}

	c.mutex.RUnlock()

	return output, ok
}

func (c *Cache[K, V]) GetAll(ids []K, output []V) []K {
	var missingIDs []K

	c.mutex.RLock()

	for index, id := range ids {
		if value, ok := c.content[id]; ok {
			output[index] = value
			c.touchLRU(id)
		} else {
			missingIDs = append(missingIDs, id)
		}
	}

	c.mutex.RUnlock()

	return missingIDs
}

func (c *Cache[K, V]) Set(id K, value V, ttl time.Duration) {
	c.mutex.Lock()

	c.content[id] = value

	if ttl != 0 {
		c.expirationUpdates <- ExpirationQueueAction[K]{id: id, ttl: ttl, action: AddItem}
	}

	c.addLRU(id)

	c.mutex.Unlock()
}

func (c *Cache[K, V]) Delete(id K) {
	c.mutex.Lock()

	delete(c.content, id)
	c.expirationUpdates <- ExpirationQueueAction[K]{id: id, action: RemoveItem}

	c.mutex.Unlock()
}
