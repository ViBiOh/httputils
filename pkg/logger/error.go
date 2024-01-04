package logger

import (
	"errors"
	"reflect"
)

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

	if joinErr, ok := err.(UnwrapJoin); ok {
		errorsToTest = append(errorsToTest, joinErr.Unwrap()...)
	}

	for index := 0; index < len(errorsToTest); index++ {
		testedErr := errorsToTest[index]

		if strackTracer, ok := testedErr.(StackTracer); ok {
			return strackTracer.StackTrace()
		}

		if unwraped := errors.Unwrap(testedErr); unwraped != nil {
			errorsToTest = append(errorsToTest, unwraped)
		}
	}

	return ""
}
