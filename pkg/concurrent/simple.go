package concurrent

import (
	"sync"

	"github.com/ViBiOh/httputils/v4/pkg/recoverer"
)

type Runner interface {
	Go(f func())
	Wait()
}

type Simple struct {
	wg sync.WaitGroup
}

func NewSimple() Runner {
	return &Simple{}
}

func (s *Simple) Go(f func()) {
	s.wg.Add(1)

	go func() {
		defer s.wg.Done()
		defer recoverer.Logger()

		f()
	}()
}

func (s *Simple) Wait() {
	s.wg.Wait()
}
