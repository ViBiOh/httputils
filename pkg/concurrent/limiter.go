package concurrent

import (
	"sync"

	"github.com/ViBiOh/httputils/v4/pkg/recoverer"
)

type Limited struct {
	limiter chan struct{}
	wg      sync.WaitGroup
}

func NewLimited(limit int) Runner {
	if limit < 0 {
		return &Simple{}
	}

	return &Limited{
		limiter: make(chan struct{}, limit),
	}
}

func (l *Limited) Go(f func()) {
	l.wg.Add(1)
	l.limiter <- struct{}{}

	go func() {
		defer l.wg.Done()
		defer func() { <-l.limiter }()
		defer recoverer.Logger()

		f()
	}()
}

func (l *Limited) Wait() {
	l.wg.Wait()
	close(l.limiter)
}
