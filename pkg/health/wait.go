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

	case 2:
		// Shortcut to avoid an extra goroutine for a common usecase
		select {
		case <-dones[0]:
		case <-dones[1]:
		}

	default:
		found := make(chan struct{})

		closer := sync.OnceFunc(func() { close(found) })

		for _, done := range dones {
			go func(done <-chan struct{}) {
				select {
				case <-found:
				case <-done:
					closer()
				}
			}(done)
		}

		<-found
	}
}
