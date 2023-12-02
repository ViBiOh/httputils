package httperror

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/ViBiOh/httputils/v4/pkg/model"
)

const (
	internalError = "Oops! Something went wrong. Server's logs contain more details."
)

func httpError(ctx context.Context, w http.ResponseWriter, status int, payload string, err error) {
	w.Header().Add("Cache-Control", "no-cache")
	http.Error(w, payload, status)

	if err == nil {
		return
	}

	if status >= http.StatusInternalServerError {
		slog.ErrorContext(ctx, err.Error(), "status", status)
	} else {
		slog.WarnContext(ctx, err.Error(), "status", status)
	}
}

func BadRequest(ctx context.Context, w http.ResponseWriter, err error) {
	httpError(ctx, w, http.StatusBadRequest, err.Error(), err)
}

func Unauthorized(ctx context.Context, w http.ResponseWriter, err error) {
	httpError(ctx, w, http.StatusUnauthorized, err.Error(), err)
}

func Forbidden(ctx context.Context, w http.ResponseWriter) {
	httpError(ctx, w, http.StatusForbidden, "⛔️", nil)
}

func NotFound(ctx context.Context, w http.ResponseWriter) {
	httpError(ctx, w, http.StatusNotFound, "¯\\_(ツ)_/¯", nil)
}

func InternalServerError(ctx context.Context, w http.ResponseWriter, err error) {
	httpError(ctx, w, http.StatusInternalServerError, internalError, err)
}

func HandleError(ctx context.Context, w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}

	switch {
	case errors.Is(err, model.ErrInvalid):
		BadRequest(ctx, w, err)
	case errors.Is(err, model.ErrUnauthorized):
		Unauthorized(ctx, w, err)
	case errors.Is(err, model.ErrForbidden):
		httpError(ctx, w, http.StatusForbidden, err.Error(), err)
	case errors.Is(err, model.ErrNotFound):
		NotFound(ctx, w)
	case errors.Is(err, model.ErrMethodNotAllowed):
		httpError(ctx, w, http.StatusMethodNotAllowed, err.Error(), err)
	default:
		InternalServerError(ctx, w, err)
	}

	return true
}

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

func FromStatus(status int, err error) error {
	switch status {
	case http.StatusBadRequest:
		return errors.Join(err, model.ErrInvalid)
	case http.StatusUnauthorized:
		return errors.Join(err, model.ErrUnauthorized)
	case http.StatusForbidden:
		return errors.Join(err, model.ErrForbidden)
	case http.StatusNotFound:
		return errors.Join(err, model.ErrNotFound)
	case http.StatusMethodNotAllowed:
		return errors.Join(err, model.ErrMethodNotAllowed)
	case http.StatusInternalServerError:
		return errors.Join(err, model.ErrInternalError)
	default:
		return err
	}
}

func FromResponse(resp *http.Response, err error) error {
	if resp == nil {
		return err
	}

	return FromStatus(resp.StatusCode, err)
}
