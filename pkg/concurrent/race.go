package concurrent

import (
	"context"
)

func ChanUntilDone[T any](ctx context.Context, sourceCh <-chan T, onSource func(T), onDone func()) {
	var closedCount uint
	done := ctx.Done()

	for closedCount < 2 {
		select {
		case <-done:
			onDone()

			done = nil
			closedCount++

		case item, ok := <-sourceCh:
			if ok {
				onSource(item)
			} else {
				closedCount++
			}
		}
	}
}
