package cors

import (
	"flag"
	"net/http"
)

var (
	origin  = flag.String(`corsOrigin`, `*`, `Access-Control-Allow-Origin`)
	headers = flag.String(`corsHeaders`, `Content-Type`, `Access-Control-Allow-Headers`)
	methods = flag.String(`corsMethods`, http.MethodGet, `Access-Control-Allow-Methods`)
)

// Handler for net/http package allowing cors header
func Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add(`Access-Control-Allow-Origin`, *origin)
		w.Header().Add(`Access-Control-Allow-Headers`, *headers)
		w.Header().Add(`Access-Control-Allow-Methods`, *methods)

		next.ServeHTTP(w, r)
	})
}
