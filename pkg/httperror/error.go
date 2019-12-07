package httperror

import (
	"net/http"

	"github.com/ViBiOh/httputils/v3/pkg/logger"
)

func httpError(w http.ResponseWriter, status int, payload string) {
	w.Header().Set("Cache-Control", "no-cache")
	http.Error(w, payload, status)
}

func logError(status int, err error) {
	logger.Warn("HTTP/%d %s: %#v", status, err.Error(), err)
}

// BadRequest logs error and sets BadRequest status
func BadRequest(w http.ResponseWriter, err error) {
	logError(http.StatusBadRequest, err)
	httpError(w, http.StatusBadRequest, err.Error())
}

// Unauthorized logs error and sets Unauthorized status
func Unauthorized(w http.ResponseWriter, err error) {
	logError(http.StatusUnauthorized, err)
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
	logError(http.StatusInternalServerError, err)
	httpError(w, http.StatusInternalServerError, err.Error())
}
