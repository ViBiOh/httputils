package concurrent

import (
	"runtime"
)

// Action defines a concurrent action with takes input and return output or error
type Action func(interface{}) (interface{}, error)

// Run create a pool of goroutines for executing action with concurrency limits (default to NumCPU)
func Run(maxConcurrent uint, action Action, onSuccess func(interface{}), onError func(error)) chan<- interface{} {
	if maxConcurrent == 0 {
		maxConcurrent = uint(runtime.NumCPU())
	}

	inputs := make(chan interface{}, maxConcurrent)

	for i := uint(0); i < maxConcurrent; i++ {
		go func() {
			for input := range inputs {
				if output, err := action(input); err != nil {
					onError(err)
				} else {
					onSuccess(output)
				}
			}
		}()
	}

	return inputs
}

// FireAndForget run concurrent action without taking care of output
func FireAndForget(maxConcurrent uint, action Action) chan<- interface{} {
	return Run(maxConcurrent, action, func(interface{}) {}, func(error) {})
}
