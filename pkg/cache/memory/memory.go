package memory

import (
	"runtime"
	"sync"
	"time"
)

type Cache[K comparable, V any] struct {
	content           map[K]V
	expiration        *ExpirationQueue[K]
	expirationUpdates chan ExpirationQueueAction[K]
	mutex             sync.RWMutex
}

func New[K comparable, V any]() *Cache[K, V] {
	return &Cache[K, V]{
		content:           make(map[K]V),
		expiration:        &ExpirationQueue[K]{},
		expirationUpdates: make(chan ExpirationQueueAction[K], runtime.NumCPU()),
	}
}

func (c *Cache[K, V]) Get(id K) (V, bool) {
	c.mutex.RLock()

	output, ok := c.content[id]

	c.mutex.RUnlock()

	return output, ok
}

func (c *Cache[K, V]) GetAll(ids []K, output []V) []K {
	var missingIDs []K

	c.mutex.RLock()

	for index, id := range ids {
		if value, ok := c.content[id]; ok {
			output[index] = value
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

	c.mutex.Unlock()
}

func (c *Cache[K, V]) Delete(id K) {
	c.mutex.Lock()

	delete(c.content, id)
	c.expirationUpdates <- ExpirationQueueAction[K]{id: id, action: RemoveItem}

	c.mutex.Unlock()
}
