package healthcheck

import (
	"net/http"
)

// App stores informations
type App struct {
	handler http.Handler
	closed  bool
}

// New creates new App for given handler
func New() *App {
	return &App{
		closed: false,
	}
}

// Handler for Health request. Should be use with net/http
func (a *App) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if a.closed {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}

			if a.handler != nil {
				a.handler.ServeHTTP(w, r)
				return
			}

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	})
}

// NextHealthcheck define sub healthcheck
func (a *App) NextHealthcheck(next http.Handler) *App {
	a.handler = next

	return a
}

// Close set all healthchecks to be unavailable
func (a *App) Close() *App {
	a.closed = true

	return a
}
