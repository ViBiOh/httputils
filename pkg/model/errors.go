package model

import (
	"errors"
	"strings"
)

var (
	ErrInvalid          = errors.New("invalid")
	ErrUnauthorized     = errors.New("unauthorized")
	ErrForbidden        = errors.New("forbidden")
	ErrNotFound         = errors.New("not found")
	ErrMethodNotAllowed = errors.New("method not allowed")
	ErrInternalError    = errors.New("internal error")
)

func WrapInvalid(err error) error {
	return errors.Join(err, ErrInvalid)
}

func WrapUnauthorized(err error) error {
	return errors.Join(err, ErrUnauthorized)
}

func WrapForbidden(err error) error {
	return errors.Join(err, ErrForbidden)
}

func WrapNotFound(err error) error {
	return errors.Join(err, ErrNotFound)
}

func WrapMethodNotAllowed(err error) error {
	return errors.Join(err, ErrMethodNotAllowed)
}

func WrapInternal(err error) error {
	return errors.Join(err, ErrInternalError)
}

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
