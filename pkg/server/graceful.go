package server

func GracefulWait(dones ...<-chan struct{}) {
	for _, done := range dones {
		<-done
	}
}
