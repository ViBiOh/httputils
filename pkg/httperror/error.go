package httperror

import (
	"net/http"

	"github.com/ViBiOh/httputils/v3/pkg/logger"
)

func httpError(w http.ResponseWriter, status int, payload string) {
	w.Header().Set("Cache-Control", "no-cache")
	http.Error(w, payload, status)
}

// BadRequest logs error and sets BadRequest status
func BadRequest(w http.ResponseWriter, err error) {
	logger.Warn("HTTP/400 %s: %#v", err.Error(), err)
	httpError(w, http.StatusBadRequest, err.Error())
}

// Unauthorized logs error and sets Unauthorized status
func Unauthorized(w http.ResponseWriter, err error) {
	logger.Warn("HTTP/401 %s: %#v", err.Error(), err)
	httpError(w, http.StatusUnauthorized, err.Error())
}

// Forbidden sets Forbidden status
func Forbidden(w http.ResponseWriter) {
	httpError(w, http.StatusForbidden, "⛔️")
}

// NotFound sets NotFound status
func NotFound(w http.ResponseWriter) {
	httpError(w, http.StatusNotFound, "¯\\_(ツ)_/¯")
}

// InternalServerError logs error and sets InternalServerError status
func InternalServerError(w http.ResponseWriter, err error) {
	logger.Error("HTTP/500 %s: %#v", err.Error(), err)
	httpError(w, http.StatusInternalServerError, err.Error())
}
