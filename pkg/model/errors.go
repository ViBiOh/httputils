package model

import (
	"errors"
	"fmt"
	"strings"
)

var (
	// ErrInvalid occurs when checks fails.
	ErrInvalid = errors.New("invalid")

	// ErrUnauthorized occurs when authorization header is missing.
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden occurs when authorization header is missing.
	ErrForbidden = errors.New("forbidden")

	// ErrNotFound occurs when something is not found.
	ErrNotFound = errors.New("not found")

	// ErrMethodNotAllowed occurs when method is not allowed.
	ErrMethodNotAllowed = errors.New("method not allowed")

	// ErrInternalError occurs when shit happens.
	ErrInternalError = errors.New("internal error")
)

// WrapError wraps err with the given wrapper.
func WrapError(err, wrapper error) error {
	if err == nil {
		return wrapper
	}

	return fmt.Errorf("%s: %w", err, wrapper)
}

// WrapInvalid wraps given error with invalid error.
func WrapInvalid(err error) error {
	return WrapError(err, ErrInvalid)
}

// WrapUnauthorized wraps given error with unauthorized error.
func WrapUnauthorized(err error) error {
	return WrapError(err, ErrUnauthorized)
}

// WrapForbidden wraps given error with forbidden error.
func WrapForbidden(err error) error {
	return WrapError(err, ErrForbidden)
}

// WrapNotFound wraps given error with not found error.
func WrapNotFound(err error) error {
	return WrapError(err, ErrNotFound)
}

// WrapMethodNotAllowed wraps given error with method not allowed error.
func WrapMethodNotAllowed(err error) error {
	return WrapError(err, ErrMethodNotAllowed)
}

// WrapInternal wraps given error with internal error.
func WrapInternal(err error) error {
	return WrapError(err, ErrInternalError)
}

// ConcatError concat errors to a single string.
func ConcatError(errs []error) error {
	if len(errs) == 0 {
		return nil
	}

	values := make([]string, len(errs))
	for index, err := range errs {
		values[index] = err.Error()
	}

	return errors.New(strings.Join(values, ", "))
}
