package wait

import (
	"net"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/logger"
)

func Wait(network, addr string, timeout time.Duration) bool {
	if timeout == 0 {
		return dial(network, addr)
	}

	timeoutTimer := time.NewTimer(timeout)
	defer func() {
		timeoutTimer.Stop()

		select {
		case <-timeoutTimer.C:
		default:
		}
	}()

	for {
		select {
		case <-timeoutTimer.C:
			return false

		default:
			if dial(network, addr) {
				return true
			}

			time.Sleep(time.Second)
		}
	}
}

func dial(network, addr string) bool {
	conn, err := net.DialTimeout(network, addr, time.Second)
	if err != nil {
		logger.Warn("dial `%s` on `%s`: %s", addr, network, err)
		return false
	}

	if closeErr := conn.Close(); closeErr != nil {
		logger.Warn("close dial: %s", err)
	}

	return true
}
