package query

import "net/http"

var ipHeaders = []string{
	"Cf-Connecting-Ip",
	"X-Forwarded-For",
	"X-Real-Ip",
}

func GetIP(r *http.Request) string {
	for _, header := range ipHeaders {
		if ip := r.Header.Get(header); len(ip) != 0 {
			return ip
		}
	}

	return r.RemoteAddr
}
