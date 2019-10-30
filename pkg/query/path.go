package query

import (
	"net/http"
	"strings"
)

// IsRoot checks if current path is root (empty or only trailing slash)
func IsRoot(r *http.Request) bool {
	return r.URL.Path == "/" || r.URL.Path == ""
}

// GetID return ID of URL (first section between two slashes)
func GetID(r *http.Request) string {
	return strings.Split(strings.Trim(r.URL.Path, "/"), "/")[0]
}
