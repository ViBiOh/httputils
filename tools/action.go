package tools

import (
	"sync"
)

// Error contains ID in error and error desc
type Error struct {
	Input []byte
	Err   error
}

func doAction(wg *sync.WaitGroup, inputs <-chan []byte, action func([]byte) (interface{}, error), results chan<- interface{}, errors chan<- *Error) {
	defer wg.Done()

	for input := range inputs {
		if result, err := action(input); err == nil {
			results <- result
		} else {
			errors <- &Error{input, err}
		}
	}
}

// ConcurrentAction create a pool of goroutines for executing action with concurrency limits
func ConcurrentAction(maxConcurrent int, action func([]byte) (interface{}, error)) (chan<- []byte, <-chan interface{}, <-chan *Error) {
	inputs := make(chan []byte, maxConcurrent)
	results := make(chan interface{}, maxConcurrent)
	errors := make(chan *Error, maxConcurrent)

	var wg sync.WaitGroup

	for i := 0; i < maxConcurrent; i++ {
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
