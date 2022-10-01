package concurrent

import (
	"sync"
)

type Simple struct {
	wg sync.WaitGroup
}

func NewSimple() *Simple {
	return &Simple{}
}

func (s *Simple) Go(f func()) {
	s.wg.Add(1)

	go func() {
		defer s.wg.Done()

		f()
	}()
}

func (s *Simple) Wait() {
	s.wg.Wait()
}
