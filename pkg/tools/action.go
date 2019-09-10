package tools

import (
	"runtime"
	"sync"

	"github.com/ViBiOh/httputils/pkg/logger"
	"github.com/ViBiOh/httputils/pkg/uuid"
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

// ConcurrentAction create a pool of goroutines for executing action with concurrency limits (default to NumCPU)
func ConcurrentAction(maxConcurrent uint, action func(interface{}) (interface{}, error)) (chan<- interface{}, <-chan ConcurentOutput) {
	if maxConcurrent == 0 {
		maxConcurrent = uint(runtime.NumCPU())
	}

	id, err := uuid.New()
	if err != nil {
		logger.Warn("unable to generate uuid: %#v", err)
	}
	logger.Info("Worker %s: starting %d in parallel", id, maxConcurrent)

	wg := sync.WaitGroup{}
	inputs := make(chan interface{}, maxConcurrent)
	results := make(chan ConcurentOutput, maxConcurrent)

	for i := uint(0); i < maxConcurrent; i++ {
		wg.Add(1)
		go doAction(&wg, inputs, action, results)
	}

	go func() {
		wg.Wait()
		close(results)
		logger.Info("Worker %s: ended", id)
	}()

	return inputs, results
}
