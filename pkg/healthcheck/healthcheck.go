package healthcheck

import (
	"net/http"
)

// App stores informations
type App struct {
	closed  bool
	handler http.Handler
}

// NewApp creates new App for given handler
func NewApp(handler http.Handler) *App {
	return &App{
		closed:  false,
		handler: handler,
	}
}

// Handler for Health request. Should be use with net/http
func (a *App) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if a.closed {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		a.handler.ServeHTTP(w, r)
	})
}

// Close set all healthchecks to be unavailable
func (a *App) Close() {
	a.closed = true
}
