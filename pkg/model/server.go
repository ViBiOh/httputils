package model

import (
	"net/http"
)

// Middleware describe a middleware in the net/http package form.
type Middleware func(http.Handler) http.Handler

// Pinger describes a function to check liveness of app.
type Pinger = func() error

// ChainMiddlewares chain middlewares call from last to first (so first item is the first called).
func ChainMiddlewares(handler http.Handler, middlewares ...Middleware) http.Handler {
	result := handler

	for i := len(middlewares) - 1; i >= 0; i-- {
		result = middlewares[i](result)
	}

	return result
}
