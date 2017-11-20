package tools

import (
	"sync"
)

// Error contains ID in error and error desc
type Error struct {
	Input interface{}
	Err   error
}

func doAction(wg *sync.WaitGroup, inputs <-chan interface{}, action func(interface{}) (interface{}, error), results chan<- interface{}, errors chan<- *Error) {
	defer wg.Done()

	for input := range inputs {
		if result, err := action(input); err == nil {
			results <- result
		} else {
			errors <- &Error{Input: input, Err: err}
		}
	}
}

// ConcurrentAction create a pool of goroutines for executing action with concurrency limits
func ConcurrentAction(maxConcurrent uint, action func(interface{}) (interface{}, error)) (chan<- interface{}, <-chan interface{}, <-chan *Error) {
	inputs := make(chan interface{}, maxConcurrent)
	results := make(chan interface{}, maxConcurrent)
	errors := make(chan *Error, maxConcurrent)

	var wg sync.WaitGroup

	for i := uint(0); i < maxConcurrent; i++ {
		wg.Add(1)
		go doAction(&wg, inputs, action, results, errors)
	}

	go func() {
		wg.Wait()
		close(results)
		close(errors)
	}()

	return inputs, results, errors
}
