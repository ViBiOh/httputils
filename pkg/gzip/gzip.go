package gzip

import (
	"net/http"

	"github.com/NYTimes/gziphandler"
	"github.com/ViBiOh/httputils/pkg/model"
)

var _ model.Middleware = &App{}

// App stores informations
type App struct {
}

// NewApp creates new App from Flags' config
func NewApp() *App {
	return &App{}
}

// Handler for request. Should be use with net/http
func (a App) Handler(next http.Handler) http.Handler {
	return gziphandler.GzipHandler(next)
}
