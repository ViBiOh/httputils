package query

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
)

var (
	// ErrInvalidInteger occurs when value is not an integer
	ErrInvalidInteger = errors.New("invalid unsigned integer value for ID")
)

// IsRoot checks if current path is root (empty or only trailing slash)
func IsRoot(r *http.Request) bool {
	return len(r.URL.Path) == 0 || r.URL.Path == "/"
}

// GetID return ID of URL (first section between two slashes)
func GetID(r *http.Request) string {
	return strings.Split(strings.Trim(r.URL.Path, "/"), "/")[0]
}

// GetUintID return ID of URL (first section between two slashes) as int64
func GetUintID(r *http.Request) (uint64, error) {
	id, err := strconv.ParseUint(GetID(r), 10, 64)
	if err != nil {
		return 0, ErrInvalidInteger
	}

	return id, nil
}
