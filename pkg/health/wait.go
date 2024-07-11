package health

import "sync"

func WaitAll(dones ...<-chan struct{}) {
	for _, done := range dones {
		<-done
	}
}

func WaitFirst(dones ...<-chan struct{}) {
	switch len(dones) {
	case 0:
	case 1:
		<-dones[0]

	default:
		found := make(chan struct{})

		var closer sync.Once

		for _, done := range dones {
			go func(done <-chan struct{}) {
				select {
				case <-found:
				case <-done:
					closer.Do(func() { close(found) })
				}
			}(done)
		}

		<-found
	}
}
