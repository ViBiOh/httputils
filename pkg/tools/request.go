package tools

import "net/http"

// IsRoot checks if current path is root (empty or only trailing slash)
func IsRoot(r *http.Request) bool {
	return r.URL.Path == `/` || r.URL.Path == ``
}
