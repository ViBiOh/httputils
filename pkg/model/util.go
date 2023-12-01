package model

import (
	"os"
	"reflect"
)

func IsNil(i any) bool {
	if i == nil {
		return true
	}

	switch reflect.TypeOf(i).Kind() {
	case reflect.Ptr, reflect.Map, reflect.Array, reflect.Chan, reflect.Slice:
		return reflect.ValueOf(i).IsNil()
	}

	return false
}

func Version() string {
	if version := os.Getenv("VERSION"); len(version) != 0 {
		return version
	}

	return "development"
}
