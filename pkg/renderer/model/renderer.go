package model

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/ViBiOh/httputils/v3/pkg/httperror"
)

const (
	internalError = "Oops! Something went wrong."
)

var (
	// ErrInvalid occurs when checks fails
	ErrInvalid = errors.New("invalid")

	// ErrNotFound occurs when something is not found
	ErrNotFound = errors.New("not found")

	// ErrMethodNotAllowed occurs when method is not allowed
	ErrMethodNotAllowed = errors.New("method not allowed")

	// ErrInternalError occurs when shit happens
	ErrInternalError = errors.New("internal error")
)

// TemplateFunc handle a request and returns which template to render with which status and datas
type TemplateFunc = func(*http.Request) (string, int, map[string]interface{}, error)

// Message for render
type Message struct {
	Level   string
	Content string
}

func newMessage(level, format string, a ...interface{}) Message {
	return Message{
		Level:   level,
		Content: fmt.Sprintf(format, a...),
	}
}

func (m Message) String() string {
	if len(m.Content) == 0 {
		return ""
	}

	return fmt.Sprintf("messageContent=%s&messageLevel=%s", url.QueryEscape(m.Content), url.QueryEscape(m.Level))
}

// ParseMessage parses messages from request
func ParseMessage(r *http.Request) Message {
	values := r.URL.Query()

	return Message{
		Level:   values.Get("messageLevel"),
		Content: values.Get("messageContent"),
	}
}

// NewSuccessMessage create a success message
func NewSuccessMessage(format string, a ...interface{}) Message {
	return newMessage("success", format, a...)
}

// NewErrorMessage create a error message
func NewErrorMessage(format string, a ...interface{}) Message {
	return newMessage("error", format, a...)
}

// ErrorStatus guess HTTP status and message from given error
func ErrorStatus(err error) (status int, message string) {
	status = http.StatusInternalServerError
	if err == nil {
		return
	}

	message = err.Error()

	switch {
	case errors.Is(err, ErrInvalid):
		status = http.StatusBadRequest
	case errors.Is(err, ErrNotFound):
		status = http.StatusNotFound
	case errors.Is(err, ErrMethodNotAllowed):
		status = http.StatusMethodNotAllowed
	default:
		message = internalError
	}

	return
}

// HandleError return a status code according to given error
func HandleError(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}

	switch {
	case errors.Is(err, ErrInvalid):
		httperror.BadRequest(w, err)
	case errors.Is(err, ErrNotFound):
		httperror.NotFound(w)
	case errors.Is(err, ErrMethodNotAllowed):
		w.WriteHeader(http.StatusMethodNotAllowed)
	default:
		httperror.InternalServerError(w, err)
	}

	return true
}

func wrapError(err, wrapper error) error {
	return fmt.Errorf("%s: %w", err, wrapper)
}

// WrapInvalid wraps given error with invalid err
func WrapInvalid(err error) error {
	return wrapError(err, ErrInvalid)
}

// WrapNotFound wraps given error with not found err
func WrapNotFound(err error) error {
	return wrapError(err, ErrNotFound)
}

// WrapMethodNotAllowed wraps given error with not method not allowed err
func WrapMethodNotAllowed(err error) error {
	return wrapError(err, ErrMethodNotAllowed)
}

// WrapInternal wraps given error with internal err
func WrapInternal(err error) error {
	return wrapError(err, ErrInternalError)
}

// ConcatError concat errors to a single string
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
