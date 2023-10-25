package memory

import (
	"container/heap"
	"time"
)

type ExpirationAction int

const (
	AddItem    ExpirationAction = iota
	RemoveItem ExpirationAction = iota

	defaultTimer = time.Hour * 24
)

type ExpirationQueueAction[K comparable] struct {
	id     K
	ttl    time.Duration
	action ExpirationAction
}

func (c *Cache[K, V]) startEvicter(done <-chan struct{}) {
	timer := time.NewTimer(defaultTimer)

	var toExpire *Item[K]

	for {
		if len(*c.expiration) > 0 {
			firstItem := c.expiration.Pop().(Item[K])
			toExpire = &firstItem

			timer.Reset(time.Until(toExpire.expiration))
		} else {
			timer.Reset(defaultTimer)
		}

		select {
		case <-done:
			timer.Stop()
			done = nil

		case _, ok := <-timer.C:
			if !ok {
				goto done
			}

			if toExpire == nil {
				continue
			}

			c.mutex.Lock()
			delete(c.content, toExpire.id)
			c.mutex.Unlock()

		case update, ok := <-c.expirationUpdates:
			if !ok {
				continue
			}

			c.handleExpirationUpdate(update, toExpire)
		}
	}

done:
	for update := range c.expirationUpdates {
		c.handleExpirationUpdate(update, toExpire)
	}
}

func (c *Cache[K, V]) handleExpirationUpdate(update ExpirationQueueAction[K], toExpire *Item[K]) {
	switch update.action {
	case AddItem:
		if toExpire != nil {
			heap.Push(c.expiration, *toExpire)
		}

		heap.Push(c.expiration, Item[K]{
			id:         update.id,
			expiration: time.Now().Add(update.ttl),
		})

	case RemoveItem:
		if index := c.expiration.Index(update.id); index != -1 {
			heap.Remove(c.expiration, index)
		}
	}
}

func (c *Cache[K, V]) addExpiration(id K, ttl time.Duration) {
	if ttl == 0 {
		return
	}

	c.expirationUpdates <- ExpirationQueueAction[K]{id: id, ttl: ttl, action: AddItem}
}
