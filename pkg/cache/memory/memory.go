package memory

import (
	"container/list"
	"runtime"
	"sync"
	"time"
)

type Cache[K comparable, V any] struct {
	content           map[K]V
	expiration        *ExpirationQueue[K]
	lru               *list.List
	expirationUpdates chan ExpirationQueueAction[K]
	mutex             sync.RWMutex
	maxSize           int
}

func New[K comparable, V any](maxSize int) *Cache[K, V] {
	return &Cache[K, V]{
		content:           make(map[K]V),
		expiration:        &ExpirationQueue[K]{},
		expirationUpdates: make(chan ExpirationQueueAction[K], runtime.NumCPU()),
		lru:               list.New(),
		maxSize:           maxSize,
	}
}

func (c *Cache[K, V]) Get(id K) (V, bool) {
	doesUpdate, unlock := c.readLock()

	output, ok := c.content[id]
	if doesUpdate && ok {
		c.touchEntry(id)
	}

	unlock()

	return output, ok
}

func (c *Cache[K, V]) readLock() (bool, func()) {
	if c.maxSize == 0 {
		c.mutex.RLock()
		return false, c.mutex.RUnlock
	}

	c.mutex.Lock()
	return true, c.mutex.Unlock
}

func (c *Cache[K, V]) GetAll(ids []K, output []V) []K {
	var missingIDs []K

	doesUpdate, unlock := c.readLock()

	for index, id := range ids {
		if value, ok := c.content[id]; ok {
			output[index] = value

			if doesUpdate {
				c.touchEntry(id)
			}
		} else {
			missingIDs = append(missingIDs, id)
		}
	}

	unlock()

	return missingIDs
}

func (c *Cache[K, V]) Set(id K, value V, ttl time.Duration) {
	c.mutex.Lock()

	c.updateLru(id)

	c.content[id] = value

	if ttl != 0 {
		c.expirationUpdates <- ExpirationQueueAction[K]{id: id, ttl: ttl, action: AddItem}
	}

	c.mutex.Unlock()
}

func (c *Cache[K, V]) Delete(id K) {
	c.mutex.Lock()

	c.delete(id)

	c.mutex.Unlock()
}

func (c *Cache[K, V]) delete(id K) {
	delete(c.content, id)
	c.expirationUpdates <- ExpirationQueueAction[K]{id: id, action: RemoveItem}
}

func (c *Cache[K, V]) updateLru(id K) {
	if c.maxSize == 0 {
		return
	}

	if _, ok := c.content[id]; ok {
		c.touchEntry(id)
		return
	}

	if c.lru.Len() == c.maxSize {
		back := c.lru.Back()
		c.lru.Remove(back)

		c.delete(back.Value.(K))
	}

	c.lru.PushFront(id)
}

func (c *Cache[K, V]) touchEntry(id K) {
	for element := c.lru.Front(); element != nil; element = element.Next() {
		if element.Value != id {
			continue
		}

		c.lru.MoveToFront(element)

		return
	}
}
