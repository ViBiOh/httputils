package model

import "net/http"

// Middleware describe a middleware in the net/http package form
type Middleware interface {
	Handler(http.Handler) http.Handler
}
