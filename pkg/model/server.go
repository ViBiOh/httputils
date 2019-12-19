package model

import "net/http"

// Middleware describe a middleware in the net/http package form
type Middleware func(http.Handler) http.Handler
