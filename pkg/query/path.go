package query

import (
	"net/http"
)

func IsRoot(r *http.Request) bool {
	return len(r.URL.Path) == 0 || r.URL.Path == "/"
}
