package gzip

import (
	"net/http"

	"github.com/NYTimes/gziphandler"
	"github.com/ViBiOh/httputils/v3/pkg/model"
)

var _ model.Middleware = Middleware

// Middleware for request. Should be use with net/http
func Middleware(next http.Handler) http.Handler {
	return gziphandler.GzipHandler(next)
}
