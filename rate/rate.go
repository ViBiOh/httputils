package rate

import (
	"flag"
	"net/http"
	"sync"
	"time"

	"github.com/ViBiOh/httputils"
)

const forwardedForHeader = `X-Forwarded-For`
const ipRateDelay = time.Second * 60

var (
	ipRateLimit = flag.Int(`rateCount`, 5000, `Rate IP limit`)
)

var ipRate = make(map[string]int)
var ipRateMutex sync.RWMutex

func init() {
	go func() {
		ticker := time.NewTicker(ipRateDelay)

		for {
			select {
			case <-ticker.C:
				ipRateMutex.Lock()
				ipRate = make(map[string]int)
				ipRateMutex.Unlock()
			}
		}
	}()
}

func getIP(r *http.Request) (ip string) {
	ip = r.Header.Get(forwardedForHeader)
	if ip == `` {
		ip = r.RemoteAddr
	}

	return
}

func checkRate(r *http.Request) bool {
	ip := getIP(r)

	ipRateMutex.Lock()
	defer ipRateMutex.Unlock()
	ipRate[ip]++

	return ipRate[ip] < *ipRateLimit
}

// Handler that check rate limit
type Handler struct {
	Handler http.Handler
}

func (handler Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !checkRate(r) {
		w.WriteHeader(http.StatusTooManyRequests)
		return
	}

	if r.Method == http.MethodGet && r.URL.Path == `/rate_limits` {
		ipRateMutex.RLock()
		defer ipRateMutex.RUnlock()

		httputils.ResponseJSON(w, http.StatusOK, ipRate)
		return
	}

	handler.Handler.ServeHTTP(w, r)
}
