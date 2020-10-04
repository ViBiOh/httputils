package model

import "net/http"

// Middleware describe a middleware in the net/http package form
type Middleware func(http.Handler) http.Handler

// Pinger describes a function to check liveness of app
type Pinger = func() error
