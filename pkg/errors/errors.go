package errors

import (
	"fmt"
	"io"
	"runtime"
	"strings"
)

var _ error = Error{}

// Error enhanced error
type Error struct {
	message    string
	stacktrace string
	origin     error
}

// New creates a new error with stack trace saved
func New(format string, a ...interface{}) error {
	return Error{
		message:    fmt.Sprintf(format, a...),
		stacktrace: stackTrace(3, 5),
	}
}

// WithStack wrap error with stack trace saved
func WithStack(err error) error {
	if err == nil {
		return nil
	}

	return Error{
		message:    err.Error(),
		stacktrace: stackTrace(3, 5),
	}
}

// Wrap wrap origin error into a new one
func Wrap(err error, format string, a ...interface{}) error {
	return Error{
		message:    fmt.Sprintf(format, a...),
		stacktrace: stackTrace(3, 5),
		origin:     err,
	}
}

// Error return string representation of error
func (e Error) Error() string {
	return e.message
}

// Format formats error
func (e Error) Format(state fmt.State, verb rune) {
	switch verb {
	case 'v':
		if state.Flag('+') {
			safeWriteString(state, e.message)

			if e.stacktrace != "" {
				safeWriteString(state, "\n")
				safeWriteString(state, e.stacktrace)
			}

			if e.origin != nil {
				safeWriteString(state, "\n\nfrom\n")
				safeWriteString(state, fmt.Sprintf("%+v", e.origin))
			}
			break
		}
		fallthrough
	case 's':
		safeWriteString(state, e.message)
	case 'q':
		if _, err := fmt.Fprintf(state, "%q", e.message); err != nil {
			fmt.Print(err)
		}
	}
}

// Cause returns origin error
func (e Error) Cause() error {
	return e.origin
}

func safeWriteString(w io.Writer, s string) {
	if _, err := io.WriteString(w, s); err != nil {
		fmt.Print(err)
	}
}

func stackTrace(skip, depth int) string {
	pc := make([]uintptr, depth)
	n := runtime.Callers(skip, pc)

	frames := runtime.CallersFrames(pc[:n])
	stacktraces := make([]string, 0)

	for {
		frame, more := frames.Next()
		if strings.Contains(frame.File, "runtime/") {
			break
		}

		stacktraces = append(stacktraces, fmt.Sprintf("%s\n\t%s:%d", frame.Function, frame.File, frame.Line))
		if !more {
			break
		}
	}

	if len(stacktraces) == 0 {
		return ""
	}

	return fmt.Sprintf("%s", strings.Join(stacktraces, "\n"))
}
