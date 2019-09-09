package tools

import (
	"sync"
)

// ConcurentOutput contains input, output and error from action
type ConcurentOutput struct {
	Input  interface{}
	Output interface{}
	Err    error
}

func doAction(wg *sync.WaitGroup, inputs <-chan interface{}, action func(interface{}) (interface{}, error), results chan<- ConcurentOutput) {
	defer wg.Done()

	for input := range inputs {
		output, err := action(input)
		results <- ConcurentOutput{
			Input:  input,
			Output: output,
			Err:    err,
		}
	}
}

// ConcurrentAction create a pool of goroutines for executing action with concurrency limits
func ConcurrentAction(maxConcurrent uint, action func(interface{}) (interface{}, error)) (chan<- interface{}, <-chan ConcurentOutput) {
	inputs := make(chan interface{}, maxConcurrent)
	results := make(chan ConcurentOutput, maxConcurrent)

	var wg sync.WaitGroup

	for i := uint(0); i < maxConcurrent; i++ {
		wg.Add(1)
		go doAction(&wg, inputs, action, results)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	return inputs, results
}
