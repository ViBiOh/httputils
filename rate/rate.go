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

var userRate = sync.Map{}

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
	rate, ok := userRate.Load(ip)

	if !ok {
		return ip, make([]*rateLimit, 0)
	}

	return ip, rate.([]*rateLimit)
}

func cleanRateLimits(rateLimits []*rateLimit, nowMinusDelaySecond int64) []*rateLimit {
	for len(rateLimits) > 0 && rateLimits[0].unix < nowMinusDelaySecond {
		rateLimits = rateLimits[1:]
	}

	return rateLimits
}

func cleanUserRate() {
	_, nowMinusDelay := getUnix()

	userRate.Range(func(key, value interface{}) bool {
		rateLimits := cleanRateLimits(value.([]*rateLimit), nowMinusDelay)
		if len(rateLimits) == 0 {
			userRate.Delete(key)
		} else {
			userRate.Store(key, rateLimits)
		}

		return true
	})
}

func sumRateLimitsCount(rateLimits []*rateLimit) (count int) {
	for _, rateLimit := range rateLimits {
		count = count + rateLimit.count
	}

	return
}

func checkRate(r *http.Request) bool {
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

	userRate.Store(ip, rateLimits)

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
		userRate.Range(func(key, value interface{}) bool {
			output[key.(string)] = sumRateLimitsCount(value.([]*rateLimit))
			return true
		})

		httputils.ResponseJSON(w, output)
		return
	}

	if !checkRate(r) {
		w.WriteHeader(http.StatusTooManyRequests)
		return
	}

	handler.Handler.ServeHTTP(w, r)
}
