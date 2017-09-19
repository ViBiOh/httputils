package rate

import (
	"flag"
	"net/http"
	"time"

	"github.com/ViBiOh/httputils"
)

const forwardedForHeader = `X-Forwarded-For`

var (
	ipRateDelay = flag.Duration(`rateDelay`, time.Second*60, `Rate IP delay`)
	ipRateLimit = flag.Int(`rateCount`, 5000, `Rate IP limit`)
)

type rateLimit struct {
	unix  int64
	count int
}

var userRate = make(map[string][]*rateLimit, 0)

func getIP(r *http.Request) (ip string) {
	ip = r.Header.Get(forwardedForHeader)
	if ip == `` {
		ip = r.RemoteAddr
	}

	return
}

func getRateLimits(r *http.Request) []*rateLimit {
	ip := getIP(r)
	rate, ok := userRate[ip]

	if !ok {
		rate = make([]*rateLimit, 0)
		userRate[ip] = rate
	}

	return rate
}

func checkRate(r *http.Request) bool {
	rateLimits := getRateLimits(r)

	now := time.Now()
	nowSecond := now.Unix()
	nowMinusDelaySecond := now.Add(*ipRateDelay * -1).Unix()

	total := 0

	for _, rateLimit := range rateLimits {
		if rateLimit.unix < nowMinusDelaySecond {
			rateLimits = rateLimits[1:]
		} else {
			if rateLimit.unix == nowSecond {
				rateLimit.count++
			}

			total = total + rateLimit.count
		}
	}

	return total < *ipRateLimit
}

// Handler that check rate limit
type Handler struct {
	Handler http.Handler
}

func (handler Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet && r.URL.Path == `/rate_limits` {
		output := map[string]int{}
		var count int

		for key, value := range userRate {
			count = 0
			for _, rateLimit := range value {
				count = count + rateLimit.count
			}

			output[key] = count
		}

		httputils.ResponseJSON(w, output)
		return
	}

	if !checkRate(r) {
		w.WriteHeader(http.StatusTooManyRequests)
		return
	}

	handler.Handler.ServeHTTP(w, r)
}
