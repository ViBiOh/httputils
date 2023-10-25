package concurrent

import (
	"context"
)

func ChanUntilDone[T any](ctx context.Context, source <-chan T, onSource func(T), onDone func()) {
	done := ctx.Done()

	for {
		select {
		case <-done:
			goto done

		case item, ok := <-source:
			if !ok {
				goto done
			}

			onSource(item)
		}
	}

done:
	onDone()

	for item := range source {
		onSource(item)
	}
}
