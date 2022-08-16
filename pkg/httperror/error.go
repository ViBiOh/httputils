package httperror

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/httputils/v4/pkg/model"
)

const (
	internalError = "Oops! Something went wrong. Server's logs contain more details."
)

func httpError(w http.ResponseWriter, status int, payload string, err error) {
	w.Header().Add("Cache-Control", "no-cache")
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

// HandleError return a status code according to given error
func HandleError(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}

	switch {
	case errors.Is(err, model.ErrInvalid):
		BadRequest(w, err)
	case errors.Is(err, model.ErrUnauthorized):
		Unauthorized(w, err)
	case errors.Is(err, model.ErrForbidden):
		httpError(w, http.StatusForbidden, err.Error(), err)
	case errors.Is(err, model.ErrNotFound):
		NotFound(w)
	case errors.Is(err, model.ErrMethodNotAllowed):
		httpError(w, http.StatusMethodNotAllowed, err.Error(), err)
	default:
		InternalServerError(w, err)
	}

	return true
}

// ErrorStatus guess HTTP status and message from given error
func ErrorStatus(err error) (status int, message string) {
	status = http.StatusInternalServerError
	if err == nil {
		return
	}

	message = err.Error()

	switch {
	case errors.Is(err, model.ErrInvalid):
		status = http.StatusBadRequest
	case errors.Is(err, model.ErrUnauthorized):
		status = http.StatusUnauthorized
	case errors.Is(err, model.ErrForbidden):
		status = http.StatusForbidden
	case errors.Is(err, model.ErrNotFound):
		status = http.StatusNotFound
	case errors.Is(err, model.ErrMethodNotAllowed):
		status = http.StatusMethodNotAllowed
	default:
		message = internalError
	}

	return
}

// FromStatus wraps model's error according to status
func FromStatus(status int, err error) error {
	switch status {
	case http.StatusBadRequest:
		return fmt.Errorf("%s: %w", err, model.ErrInvalid)
	case http.StatusUnauthorized:
		return fmt.Errorf("%s: %w", err, model.ErrUnauthorized)
	case http.StatusForbidden:
		return fmt.Errorf("%s: %w", err, model.ErrForbidden)
	case http.StatusNotFound:
		return fmt.Errorf("%s: %w", err, model.ErrNotFound)
	case http.StatusMethodNotAllowed:
		return fmt.Errorf("%s: %w", err, model.ErrMethodNotAllowed)
	case http.StatusInternalServerError:
		return fmt.Errorf("%s: %w", err, model.ErrInternalError)
	default:
		return err
	}
}
