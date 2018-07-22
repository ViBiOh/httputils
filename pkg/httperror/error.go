package httperror

import (
	"log"
	"net/http"

	"github.com/ViBiOh/httputils/pkg/rollbar"
)

// BadRequest logs error and sets BadRequest status
func BadRequest(w http.ResponseWriter, err error) {
	log.Printf(`HTTP/400 %v`, err)
	rollbar.Warning(`HTTP/400 %v`, err)
	http.Error(w, err.Error(), http.StatusBadRequest)
}

// Unauthorized logs error and sets Unauthorized status
func Unauthorized(w http.ResponseWriter, err error) {
	log.Printf(`HTTP/401 %v`, err)
	rollbar.Warning(`HTTP/401 %v`, err)
	http.Error(w, err.Error(), http.StatusUnauthorized)
}

// Forbidden sets Forbidden status
func Forbidden(w http.ResponseWriter) {
	http.Error(w, `⛔️`, http.StatusForbidden)
}

// NotFound sets NotFound status
func NotFound(w http.ResponseWriter) {
	http.Error(w, `¯\_(ツ)_/¯`, http.StatusNotFound)
}

// InternalServerError logs error and sets InternalServerError status
func InternalServerError(w http.ResponseWriter, err error) {
	log.Printf(`HTTP/500 %v`, err)
	rollbar.Error(`HTTP/500 %v`, err)
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
