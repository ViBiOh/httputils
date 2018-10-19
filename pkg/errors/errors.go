package errors

import (
	"fmt"
	"io"
	"runtime"
	"strings"
)

var _ error = New(``)

type enhancedError struct {
	message string
	callers []uintptr
}

// New creates a new error with stack trace saved
func New(format string, a ...interface{}) error {
	return enhancedError{
		message: fmt.Sprintf(format, a...),
		callers: callers(3, 5),
	}
}

// WithStack wrap error with stack trace saved
func WithStack(err error) error {
	return enhancedError{
		message: err.Error(),
		callers: callers(3, 5),
	}
}

func (e enhancedError) Error() string {
	return e.message
}

func (e enhancedError) Format(state fmt.State, verb rune) {
	switch verb {
	case 'v':
		if state.Flag('+') {
			safeWriteString(state, e.message)
			safeWriteString(state, "\n")
			safeWriteString(state, stackTrace(e.callers))
			break
		}
		fallthrough
	case 's':
		safeWriteString(state, e.message)
	case 'q':
		if _, err := fmt.Fprintf(state, `%q`, e.message); err != nil {
			fmt.Print(err)
		}
	}
}

func safeWriteString(w io.Writer, s string) {
	if _, err := io.WriteString(w, s); err != nil {
		fmt.Print(err)
	}
}

func callers(skip, depth int) []uintptr {
	pc := make([]uintptr, depth)
	n := runtime.Callers(skip, pc)

	if n == 0 {
		return nil
	}

	return pc[:n]
}

func stackTrace(pc []uintptr) string {
	frames := runtime.CallersFrames(pc)
	stacktraces := make([]string, 0)

	for {
		frame, more := frames.Next()
		if strings.Contains(frame.File, `runtime/`) {
			break
		}

		stacktraces = append(stacktraces, fmt.Sprintf("%s\n\t%s:%d", frame.Function, frame.File, frame.Line))
		if !more {
			break
		}
	}

	if len(stacktraces) == 0 {
		return ``
	}

	return fmt.Sprintf("%s", strings.Join(stacktraces, "\n"))
}
