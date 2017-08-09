package cors

import (
	"flag"
	"net/http"
)

var (
	origin  = flag.String(`corsOrigin`, `*`, `Access-Control-Allow-Origin`)
	headers = flag.String(`corsHeaders`, `Content-Type`, `Access-Control-Allow-Headers`)
	methods = flag.String(`corsMethods`, `GET`, `Access-Control-Allow-Methods`)
)

// Handler for net/http package allowing cors header
type Handler struct {
	http.Handler
}

func (handler Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add(`Access-Control-Allow-Origin`, *origin)
	w.Header().Add(`Access-Control-Allow-Headers`, *headers)
	w.Header().Add(`Access-Control-Allow-Methods`, *methods)

	handler.Handler.ServeHTTP(w, r)
}
