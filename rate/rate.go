package rate

import (
	"flag"
	"net/http"
	"sync"
	"time"

	"github.com/ViBiOh/httputils"
	"github.com/ViBiOh/httputils/tools"
)

const defaultLimit = uint(5000)
const forwardedForHeader = `X-Forwarded-For`
const ipRateDelay = time.Second * 60

// Flags add flags for given prefix
func Flags(prefix string) map[string]interface{} {
	return map[string]interface{}{
		`limit`: flag.Uint(tools.ToCamel(prefix+`Count`), defaultLimit, `[rate] IP limit`),
	}
}

var ipRate = make(map[string]uint)
var ipRateMutex sync.RWMutex

func init() {
	go func() {
		ticker := time.NewTicker(ipRateDelay)

		for {
			select {
			case <-ticker.C:
				ipRateMutex.Lock()
				ipRate = make(map[string]uint)
				ipRateMutex.Unlock()
			}
		}
	}()
}

// GetIP give remote IP
func GetIP(r *http.Request) (ip string) {
	ip = r.Header.Get(forwardedForHeader)
	if ip == `` {
		ip = r.RemoteAddr
	}

	return
}

func checkRate(r *http.Request, limit uint) bool {
	ip := GetIP(r)

	ipRateMutex.Lock()
	defer ipRateMutex.Unlock()
	ipRate[ip]++

	return ipRate[ip] < limit
}

// Handler that check rate limit
func Handler(config map[string]interface{}, next http.Handler) http.Handler {
	var (
		limit = defaultLimit
	)

	var given interface{}
	var ok bool

	if given, ok = config[`limit`]; ok {
		limit = *(given.(*uint))
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !checkRate(r, limit) {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}

		if r.Method == http.MethodGet && r.URL.Path == `/rate_limits` {
			ipRateMutex.RLock()
			defer ipRateMutex.RUnlock()

			httputils.ResponseJSON(w, http.StatusOK, ipRate, httputils.IsPretty(r.URL.RawQuery))
			return
		}

		next.ServeHTTP(w, r)
	})
}
