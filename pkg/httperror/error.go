package httperror

import (
	"net/http"

	"github.com/ViBiOh/httputils/v3/pkg/logger"
)

const (
	internalError = "Oops! Something went wrong. Server's logs contain more details."
)

func httpError(w http.ResponseWriter, status int, payload string, err error) {
	w.Header().Set("Cache-Control", "no-cache")
	http.Error(w, payload, status)

	if err == nil {
		return
	}

	if status >= http.StatusInternalServerError {
		logger.Error("HTTP/%d: %s", status, err.Error())
	} else {
		logger.Warn("HTTP/%d: %s", status, err.Error())
	}
}

// BadRequest logs error and sets BadRequest status
func BadRequest(w http.ResponseWriter, err error) {
	httpError(w, http.StatusBadRequest, err.Error(), err)
}

// Unauthorized logs error and sets Unauthorized status
func Unauthorized(w http.ResponseWriter, err error) {
	httpError(w, http.StatusUnauthorized, err.Error(), err)
}

// Forbidden sets Forbidden status
func Forbidden(w http.ResponseWriter) {
	httpError(w, http.StatusForbidden, "⛔️", nil)
}

// NotFound sets NotFound status
func NotFound(w http.ResponseWriter) {
	httpError(w, http.StatusNotFound, "¯\\_(ツ)_/¯", nil)
}

// InternalServerError logs error and sets InternalServerError status
func InternalServerError(w http.ResponseWriter, err error) {
	httpError(w, http.StatusInternalServerError, internalError, err)
}
