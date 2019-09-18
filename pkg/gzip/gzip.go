package gzip

import (
	"net/http"

	"github.com/NYTimes/gziphandler"
	"github.com/ViBiOh/httputils/v2/pkg/model"
)

var _ model.Middleware = &App{}

// App stores informations
type App struct {
}

// New creates new App
func New() *App {
	return &App{}
}

// Handler for request. Should be use with net/http
func (a App) Handler(next http.Handler) http.Handler {
	return gziphandler.GzipHandler(next)
}
