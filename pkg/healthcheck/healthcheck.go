package healthcheck

import (
	"net/http"
)

// App stores informations
type App struct {
	closed bool
}

// NewApp creates new App for given handler
func NewApp() *App {
	return &App{
		closed: false,
	}
}

// Handler for Health request. Should be use with net/http
func (a *App) Handler(next http.Handler) http.Handler {
	handler := next
	if handler == nil {
		handler = Basic()
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if a.closed {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		handler.ServeHTTP(w, r)
	})
}

// Close set all healthchecks to be unavailable
func (a *App) Close() {
	a.closed = true
}
