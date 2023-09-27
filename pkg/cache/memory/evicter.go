package memory

import (
	"container/heap"
	"context"
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

func (c *Cache[K, V]) Start(ctx context.Context) {
	timer := time.NewTimer(defaultTimer)

	var toExpire *Item[K]

	for {
		if len(*c.expiration) > 0 {
			firstItem := c.expiration.Pop().(Item[K])
			toExpire = &firstItem

			timer.Reset(time.Since(toExpire.expiration))
		} else {
			timer.Reset(defaultTimer)
		}

		select {
		case <-ctx.Done():
			goto exit

		case update, ok := <-c.expirationUpdates:
			if !ok {
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

		case <-timer.C:
			if toExpire == nil {
				continue
			}

			c.mutex.Lock()

			delete(c.content, toExpire.id)

			c.mutex.Unlock()
		}
	}

exit:
	if !timer.Stop() {
		<-timer.C
	}

	close(c.expirationUpdates)
	for range c.expirationUpdates {
		// drain the channel
	}
}
