package logger

import (
	"errors"
	"reflect"
)

// Interface to unwrap joined errors.Join https://pkg.go.dev/errors#Join
type UnwrapJoin interface {
	Unwrap() []error
}

type StackTracer interface {
	StackTrace() string
}

type errorField struct {
	Kind    string `json:"kind"`
	Message string `json:"message"`
	Stack   string `json:"stack,omitempty"`
}

func ErrorField(err error) errorField {
	output := errorField{
		Kind:    reflect.TypeOf(err).String(),
		Message: err.Error(),
	}

	if stacktrace := getStacktrace(err); len(stacktrace) != 0 {
		output.Stack = stacktrace
	}

	return output
}

func getStacktrace(err error) string {
	errorsToTest := []error{err}

	for index := 0; index < len(errorsToTest); index++ {
		testedErr := errorsToTest[index]

		if stackTracer, ok := testedErr.(StackTracer); ok {
			return stackTracer.StackTrace()
		}

		if joinErr, ok := testedErr.(UnwrapJoin); ok {
			errorsToTest = append(errorsToTest, joinErr.Unwrap()...)
		} else if unwraped := errors.Unwrap(testedErr); unwraped != nil {
			errorsToTest = append(errorsToTest, unwraped)
		}
	}

	return ""
}
