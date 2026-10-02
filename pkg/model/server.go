package model

import (
	"context"
	"net/http"
	"slices"
)

type Middleware func(http.Handler) http.Handler

type Pinger = func(context.Context) error

func ChainMiddlewares(handler http.Handler, middlewares ...Middleware) http.Handler {
	result := handler

	for _, middleware := range slices.Backward(middlewares) {
		result = middleware(result)
	}

	return result
}
