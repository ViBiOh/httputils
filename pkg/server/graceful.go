package server

// GracefulWait wait for all done chan to be closed.
func GracefulWait(dones ...<-chan struct{}) {
	for _, done := range dones {
		<-done
	}
}
