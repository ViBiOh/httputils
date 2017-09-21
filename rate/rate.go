package rate

import (
	"flag"
	"net/http"
	"sync"
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
var userRateMutex sync.RWMutex

func getIP(r *http.Request) (ip string) {
	ip = r.Header.Get(forwardedForHeader)
	if ip == `` {
		ip = r.RemoteAddr
	}

	return
}

func getUnix() (int64, int64) {
	ts := time.Now()
	return ts.Unix(), ts.Add(*ipRateDelay * -1).Unix()
}

func getRateLimits(r *http.Request) (string, []*rateLimit) {
	ip := getIP(r)
	rate, ok := userRate[ip]

	if !ok {
		return ip, make([]*rateLimit, 0)
	}

	return ip, rate
}

func cleanRateLimits(rateLimits []*rateLimit, nowMinusDelaySecond int64) []*rateLimit {
	for len(rateLimits) > 0 && rateLimits[0].unix < nowMinusDelaySecond {
		rateLimits = rateLimits[1:]
	}

	return rateLimits
}

func cleanUserRate() {
	userRateMutex.Lock()
	defer userRateMutex.Unlock()

	_, nowMinusDelay := getUnix()

	for key, value := range userRate {
		rateLimits := cleanRateLimits(value, nowMinusDelay)
		if len(rateLimits) == 0 {
			delete(userRate, key)
		} else {
			userRate[key] = rateLimits
		}
	}
}

func sumRateLimitsCount(rateLimits []*rateLimit) (count int) {
	for _, rateLimit := range rateLimits {
		count = count + rateLimit.count
	}

	return
}

func checkRate(r *http.Request) bool {
	userRateMutex.Lock()
	ip, rateLimits := getRateLimits(r)
	now, nowMinusDelay := getUnix()

	rateLimits = cleanRateLimits(rateLimits, nowMinusDelay)
	lastIndex := len(rateLimits) - 1

	if lastIndex >= 0 && rateLimits[lastIndex].unix == now {
		rateLimits[lastIndex].count++
	} else {
		rateLimits = append(rateLimits, &rateLimit{now, 1})
	}
	sum := sumRateLimitsCount(rateLimits)

	userRate[ip] = rateLimits
	userRateMutex.Unlock()

	return sum < *ipRateLimit
}

// Handler that check rate limit
type Handler struct {
	Handler http.Handler
}

func (handler Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet && r.URL.Path == `/rate_limits` {
		cleanUserRate()

		output := map[string]int{}
		for key, value := range userRate {
			output[key] = sumRateLimitsCount(value)
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
