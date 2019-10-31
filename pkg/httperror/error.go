package httperror

import (
	"net/http"

	"github.com/ViBiOh/httputils/v3/pkg/logger"
)

// BadRequest logs error and sets BadRequest status
func BadRequest(w http.ResponseWriter, err error) {
	logger.Warn("HTTP/400 %s: %#v", err.Error(), err)
	http.Error(w, err.Error(), http.StatusBadRequest)
}

// Unauthorized logs error and sets Unauthorized status
func Unauthorized(w http.ResponseWriter, err error) {
	logger.Warn("HTTP/401 %s: %#v", err.Error(), err)
	http.Error(w, err.Error(), http.StatusUnauthorized)
}

// Forbidden sets Forbidden status
func Forbidden(w http.ResponseWriter) {
	http.Error(w, "⛔️", http.StatusForbidden)
}

// NotFound sets NotFound status
func NotFound(w http.ResponseWriter) {
	http.Error(w, "¯\\_(ツ)_/¯", http.StatusNotFound)
}

// InternalServerError logs error and sets InternalServerError status
func InternalServerError(w http.ResponseWriter, err error) {
	logger.Error("HTTP/500 %s: %#v", err.Error(), err)
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
