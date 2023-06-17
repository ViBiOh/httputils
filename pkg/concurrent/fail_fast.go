package concurrent

import (
	"context"
	"sync"

	"github.com/ViBiOh/httputils/v4/pkg/recoverer"
)

type FailFast struct {
	err     error
	done    chan struct{}
	limiter chan struct{}
	cancel  context.CancelFunc
	once    sync.Once
	wg      sync.WaitGroup
}

func NewFailFast(limit uint64) *FailFast {
	return &FailFast{
		done:    make(chan struct{}),
		limiter: make(chan struct{}, limit),
	}
}

func (ff *FailFast) WithContext(ctx context.Context) context.Context {
	if ff.cancel != nil {
		panic("cancelable context already set-up")
	}

	ctx, ff.cancel = context.WithCancel(ctx)

	return ctx
}

func (ff *FailFast) Go(f func() error) {
	select {
	case <-ff.done:
		return
	case ff.limiter <- struct{}{}:
		ff.wg.Add(1)

		go func() {
			defer ff.wg.Done()
			defer func() { <-ff.limiter }()

			var err error

			defer func() {
				if err != nil {
					ff.close(err)
				}
			}()

			defer recoverer.Error(&err)

			err = f()
		}()
	}
}

func (ff *FailFast) Wait() error {
	ff.wg.Wait()

	select {
	case <-ff.done:
	default:
		close(ff.done)
	}

	close(ff.limiter)

	return ff.err
}

func (ff *FailFast) close(err error) {
	ff.once.Do(func() {
		close(ff.done)

		if ff.cancel != nil {
			ff.cancel()
		}

		ff.err = err
	})
}
