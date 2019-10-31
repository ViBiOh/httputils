package gzip

import (
	"net/http"

	"github.com/NYTimes/gziphandler"
	"github.com/ViBiOh/httputils/v2/pkg/model"
)

var _ model.Middleware = &app{}

// App stores informations
type App interface {
	Handler(http.Handler) http.Handler
}

type app struct {
}

// New creates new App
func New() App {
	return &app{}
}

// Handler for request. Should be use with net/http
func (a app) Handler(next http.Handler) http.Handler {
	return gziphandler.GzipHandler(next)
}
