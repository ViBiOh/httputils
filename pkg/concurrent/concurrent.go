package concurrent

import (
	"context"
	"sync"
)

// Group describes a task group with a fail-fast approach
type Group struct {
	err     error
	done    chan struct{}
	limiter chan bool
	cancel  context.CancelFunc
	once    sync.Once
	wg      sync.WaitGroup
}

// NewGroup creates a Group with given concurrency limit
func NewGroup(limit uint64) *Group {
	return &Group{
		done:    make(chan struct{}),
		limiter: make(chan bool, limit),
	}
}

// WithContext make the given context cancelable for the group
func (g *Group) WithContext(ctx context.Context) context.Context {
	if g.cancel != nil {
		panic("cancelable context already set-up")
	}

	ctx, g.cancel = context.WithCancel(ctx)
	return ctx
}

// Go run given function in a goroutine according to limiter and current status
func (g *Group) Go(f func() error) {
	select {
	case <-g.done:
	case g.limiter <- true:
		g.wg.Add(1)

		go func() {
			defer g.wg.Done()
			defer func() { <-g.limiter }()

			if err := f(); err != nil {
				g.close(err)
			}
		}()
	}
}

// Wait for Group to end
func (g *Group) Wait() error {
	g.wg.Wait()

	select {
	case <-g.done:
	default:
		close(g.done)
	}

	close(g.limiter)

	return g.err
}

func (g *Group) close(err error) {
	g.once.Do(func() {
		close(g.done)

		if g.cancel != nil {
			g.cancel()
		}

		g.err = err
	})
}
