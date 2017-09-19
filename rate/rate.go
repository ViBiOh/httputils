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
	ipRateCount = flag.Int(`rateCount`, 5000, `Rate IP count`)
)

type rateLimit struct {
	Calls []time.Time `json:"calls"`
}

var userRate = make(map[string]*rateLimit, 0)

func checkRate(r *http.Request) bool {
	ip := r.Header.Get(forwardedForHeader)
	if ip == `` {
		ip = r.RemoteAddr
	}
	rate, ok := userRate[ip]

	if !ok {
		rate = &rateLimit{make([]time.Time, 0)}
		userRate[ip] = rate
	}

	now := time.Now()
	nowMinusDelay := now.Add(*ipRateDelay * -1)

	rate.Calls = append(rate.Calls, now)
	for len(rate.Calls) > 0 && rate.Calls[0].Before(nowMinusDelay) {
		rate.Calls = rate.Calls[1:]
	}

	return len(rate.Calls) < *ipRateCount
}

// Handler that check rate limit
type Handler struct {
	Handler http.Handler
}

func (handler Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet && r.URL.Path == `/rate_limits` {
		httputils.ResponseJSON(w, userRate)
		return
	}

	if !checkRate(r) {
		w.WriteHeader(http.StatusTooManyRequests)
		return
	}

	handler.Handler.ServeHTTP(w, r)
}
