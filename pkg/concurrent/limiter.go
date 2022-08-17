package concurrent

import (
	"sync"
)

// Limited describes a task group with limited parallelism.
type Limited struct {
	limiter chan bool
	wg      sync.WaitGroup
}

// NewLimited creates a Limited with given concurrency limit.
func NewLimited(limit uint64) *Limited {
	return &Limited{
		limiter: make(chan bool, limit),
	}
}

// Go run given function in a goroutine according to limiter.
func (g *Limited) Go(f func()) {
	g.wg.Add(1)
	g.limiter <- true

	go func() {
		defer g.wg.Done()
		defer func() { <-g.limiter }()

		f()
	}()
}

// Wait for Limited to end.
func (g *Limited) Wait() {
	g.wg.Wait()
	close(g.limiter)
}
