package query

import (
	"net/http"
)

// IsRoot checks if current path is root (empty or only trailing slash).
func IsRoot(r *http.Request) bool {
	return len(r.URL.Path) == 0 || r.URL.Path == "/"
}
