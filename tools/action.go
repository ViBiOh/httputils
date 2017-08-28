package tools

// Error contains ID in error and error desc
type Error struct {
	Input []byte
	Err   error
}

func doAction(inputs <-chan []byte, action func([]byte) (interface{}, error), results chan<- interface{}, errors chan<- *Error) {
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

	for i := 0; i < maxConcurrent; i++ {
		go doAction(inputs, action, results, errors)
	}

	return inputs, results, errors
}
