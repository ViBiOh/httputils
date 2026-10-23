package memory

import (
	"container/heap"
	"container/list"
	"context"
	"runtime"
	"sync"
	"time"
)

type entry[K comparable, V any] struct {
	expiration time.Time
	id         K
	value      V
	element    *list.Element
	index      int
}

func (item *entry[K, V]) isExpired() bool {
	return !item.expiration.IsZero() && time.Now().After(item.expiration)
}

type Cache[K comparable, V any] struct {
	content    map[K]*entry[K, V]
	lru        *list.List
	touches    chan K
	wake       chan struct{}
	expiration expirationQueue[K, V]
	mutex      sync.RWMutex
	maxSize    int
}

func New[K comparable, V any](maxSize int) *Cache[K, V] {
	return &Cache[K, V]{
		content: make(map[K]*entry[K, V]),
		lru:     list.New(),
		touches: make(chan K, runtime.NumCPU()*10), // read ratio 10:1
		wake:    make(chan struct{}, 1),
		maxSize: maxSize,
	}
}

func (c *Cache[K, V]) Start(ctx context.Context) {
	var wait sync.WaitGroup

	wait.Go(func() { c.startLRU(ctx) })

	c.startEvicter(ctx)

	wait.Wait()
}

func (c *Cache[K, V]) Get(id K) (V, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return c.guardedGet(id)
}

func (c *Cache[K, V]) GetAll(ids []K, output []V) []K {
	var missingIDs []K

	c.mutex.RLock()
	defer c.mutex.RUnlock()

	for index, id := range ids {
		if item, ok := c.guardedGet(id); !ok {
			missingIDs = append(missingIDs, id)
		} else {
			output[index] = item
		}
	}

	return missingIDs
}

func (c *Cache[K, V]) guardedGet(id K) (V, bool) {
	item, ok := c.content[id]
	if !ok || item.isExpired() {
		var output V

		return output, false
	}

	c.touch(id)

	return item.value, true
}

func (c *Cache[K, V]) Set(id K, value V, ttl time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	item, ok := c.content[id]
	if !ok {
		item = &entry[K, V]{id: id, index: -1}
		c.content[id] = item
	}

	item.value = value

	c.setExpiration(item, ttl)
	c.markUsed(item)
}

func (c *Cache[K, V]) Delete(id K) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if item, ok := c.content[id]; ok {
		c.remove(item)
	}
}

func (c *Cache[K, V]) remove(item *entry[K, V]) {
	delete(c.content, item.id)

	if item.element != nil {
		c.lru.Remove(item.element)
	}

	if item.index != -1 {
		heap.Remove(&c.expiration, item.index)
	}
}
