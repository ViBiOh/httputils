package model

import (
	"net/http"
)

type Middleware func(http.Handler) http.Handler

type Pinger = func() error

func ChainMiddlewares(handler http.Handler, middlewares ...Middleware) http.Handler {
	result := handler

	for i := len(middlewares) - 1; i >= 0; i-- {
		result = middlewares[i](result)
	}

	return result
}
