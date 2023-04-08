package waitcp

import (
	"net"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/logger"
)

func Wait(scheme, addr string, timeout time.Duration) bool {
	if timeout == 0 {
		return dial(scheme, addr)
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
			if dial(scheme, addr) {
				return true
			}

			time.Sleep(time.Second)
		}
	}
}

func dial(scheme, addr string) bool {
	conn, err := net.DialTimeout(scheme, addr, time.Second)
	if err != nil {
		logger.Warn("dial `%s` on `%s`: %s", addr, scheme, err)
		return false
	}

	if closeErr := conn.Close(); closeErr != nil {
		logger.Warn("close dial: %s", err)
	}

	return true
}
