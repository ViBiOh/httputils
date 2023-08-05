package concurrent

import (
	"sync"

	"github.com/ViBiOh/httputils/v4/pkg/recoverer"
)

type Limiter struct {
	limiter chan struct{}
	wg      sync.WaitGroup
}

func NewLimiter(limit int) *Limiter {
	var limiter chan struct{}

	if limit > 0 {
		limiter = make(chan struct{}, limit)
	}

	return &Limiter{
		limiter: limiter,
	}
}

func (l *Limiter) Go(f func()) {
	l.wg.Add(1)

	if l.limiter != nil {
		l.limiter <- struct{}{}
	}

	go l.run(f)
}

func (l *Limiter) run(f func()) {
	defer l.wg.Done()
	if l.limiter != nil {
		defer func() { <-l.limiter }()
	}

	defer recoverer.Logger()

	f()
}

func (l *Limiter) Wait() {
	l.wg.Wait()

	if l.limiter != nil {
		close(l.limiter)
	}
}
