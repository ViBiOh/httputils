package healthcheck

import "net/http"

// Basic handler, return always true for healthcheck. Should be use with net/http
func Basic() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
}
