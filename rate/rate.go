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

type rate struct {
	unix  int64
	count int
}

var ipRate = sync.Map{}

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

func getRates(r *http.Request) (string, []*rate) {
	ip := getIP(r)
	rates, ok := ipRate.Load(ip)

	if !ok {
		return ip, make([]*rate, 0)
	}

	return ip, rates.([]*rate)
}

func cleanRates(rates []*rate, nowMinusDelaySecond int64) []*rate {
	for len(rates) > 0 && rates[0].unix < nowMinusDelaySecond {
		rates = rates[1:]
	}

	return rates
}

func cleanIPRate() {
	_, nowMinusDelay := getUnix()

	ipRate.Range(func(ip, rates interface{}) bool {
		cleanedRates := cleanRates(rates.([]*rate), nowMinusDelay)
		if len(cleanedRates) == 0 {
			ipRate.Delete(ip)
		} else {
			ipRate.Store(ip, cleanedRates)
		}

		return true
	})
}

func sumRates(rates []*rate) int {
	sum := 0

	for i := 0; i < len(rates) && sum < *ipRateLimit; i++ {
		sum = sum + rates[i].count
	}

	return sum
}

func checkRate(r *http.Request) bool {
	ip, rates := getRates(r)
	now, nowMinusDelay := getUnix()

	rates = cleanRates(rates, nowMinusDelay)
	lastIndex := len(rates) - 1

	if lastIndex >= 0 && rates[lastIndex].unix == now {
		rates[lastIndex].count++
	} else {
		rates = append(rates, &rate{now, 1})
	}
	sum := sumRates(rates)

	ipRate.Store(ip, rates)

	return sum < *ipRateLimit
}

// Handler that check rate limit
type Handler struct {
	Handler http.Handler
}

func (handler Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet && r.URL.Path == `/rate_limits` {
		cleanIPRate()

		output := map[string]int{}
		ipRate.Range(func(ip, rates interface{}) bool {
			output[ip.(string)] = sumRates(rates.([]*rate))
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
