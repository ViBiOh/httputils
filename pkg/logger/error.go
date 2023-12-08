package logger

import (
	"reflect"
)

type errorField struct {
	Kind    string `json:"kind"`
	Message string `json:"message"`
}

func ErrorField(err error) errorField {
	return errorField{
		Kind:    reflect.TypeOf(err).String(),
		Message: err.Error(),
	}
}
