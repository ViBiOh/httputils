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

	var closedCount int

	timerCh := timer.C
	expirationCh := c.expirationUpdates

	for closedCount < 3 {
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
			c.expirationUpdates = nil
			close(expirationCh)

			done = nil
			closedCount++

		case update, ok := <-expirationCh:
			if !ok {
				closedCount++
				continue
			}

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

		case _, ok := <-timerCh:
			if !ok {
				timerCh = nil
				closedCount++
				continue
			}

			if toExpire == nil {
				continue
			}

			c.mutex.Lock()
			delete(c.content, toExpire.id)
			c.mutex.Unlock()
		}
	}
}

func (c *Cache[K, V]) sendExpirationAction(id K, ttl time.Duration) {
	if ttl == 0 {
		return
	}

	select {
	case c.expirationUpdates <- ExpirationQueueAction[K]{id: id, ttl: ttl, action: AddItem}:
	default:
	}
}
